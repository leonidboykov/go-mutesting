package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/printer"
	"go/token"
	"go/types"
	"io"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/lmittmann/tint"
	altsrc "github.com/urfave/cli-altsrc/v3"
	yamlsrc "github.com/urfave/cli-altsrc/v3/yaml"
	"github.com/urfave/cli/v3"

	"github.com/leonidboykov/go-mutesting"
	"github.com/leonidboykov/go-mutesting/astutil"
	"github.com/leonidboykov/go-mutesting/internal/importing"
	"github.com/leonidboykov/go-mutesting/internal/models"
	"github.com/leonidboykov/go-mutesting/mutator"
	_ "github.com/leonidboykov/go-mutesting/mutator/arithmetic"
	_ "github.com/leonidboykov/go-mutesting/mutator/branch"
	_ "github.com/leonidboykov/go-mutesting/mutator/expression"
	_ "github.com/leonidboykov/go-mutesting/mutator/loop"
	_ "github.com/leonidboykov/go-mutesting/mutator/numbers"
	_ "github.com/leonidboykov/go-mutesting/mutator/statement"
)

func debug(format string, args ...any) {
	slog.Debug(fmt.Sprintf(format, args...))
}

func verbose(format string, args ...any) {
	slog.Info(fmt.Sprintf(format, args...))
}

type mutatorItem struct {
	Name    string
	Mutator mutator.Mutator
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	logLevel := new(slog.LevelVar)
	logLevel.Set(slog.LevelInfo)
	slog.SetDefault(
		slog.New(tint.NewHandler(os.Stderr, &tint.Options{
			Level:      logLevel,
			TimeFormat: time.DateTime,
		})),
	)
	slog.SetLogLoggerLevel(slog.LevelError)

	var configFile string
	if err := (&cli.Command{
		EnableShellCompletion: true,
		Usage:                 "mutation testing for Go source code.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				Usage:       "path to config `FILE`",
				Destination: &configFile,
			},
			&cli.BoolFlag{
				Name:  "debug",
				Usage: "debug log output",
			},
			&cli.BoolFlag{
				Name:  "verbose",
				Usage: "verbose log output",
			},
			&cli.StringSliceFlag{
				Name:  "disable",
				Usage: "disable mutator by their name or using * as a suffix pattern (in order to check remaining enabled mutators use --verbose option)",
			},
			&cli.StringSliceFlag{
				Name:  "blacklist",
				Usage: "list of MD5 checksums of mutations which should be ignored. Each checksum must end with a new line character",
			},
			&cli.StringFlag{
				Name:  "match",
				Usage: "only functions are mutated that confirm to the arguments regex",
			},
			&cli.BoolFlag{
				Name:  "test-recursive",
				Usage: "defines if the executer should test recursively",
			},
			&cli.BoolFlag{
				Name:  "do-not-remove-tmp-folder",
				Usage: "do not remove the tmp folder where all mutations are saved to",
			},
			&cli.BoolFlag{
				Name:  "skip-without-test",
				Usage: "skip all files without related _test.go file",
				Sources: cli.NewValueSourceChain(
					yamlsrc.YAML("skip_without_test", altsrc.NewStringPtrSourcer(&configFile)),
				),
			},
			&cli.BoolFlag{
				Name:  "skip-with-build-tags",
				Usage: "skip all files with build tags in related _test.go file",
				Sources: cli.NewValueSourceChain(
					yamlsrc.YAML("skip_with_build_tags", altsrc.NewStringPtrSourcer(&configFile)),
				),
			},
			&cli.StringFlag{
				Name:  "exec",
				Usage: "execute this command for every mutation (by default the built-in exec command is used)",
			},
			&cli.BoolFlag{
				Name:  "no-exec",
				Usage: "skip the built-in exec command and just generate the mutations",
			},
			&cli.UintFlag{
				Name:  "exec-timeout",
				Usage: "sets a timeout for the command execution in seconds",
				Value: 10,
			},
			&cli.BoolFlag{
				Name:  "silent-mode",
				Usage: "suppress output",
				Sources: cli.NewValueSourceChain(
					yamlsrc.YAML("silent_mode", altsrc.NewStringPtrSourcer(&configFile)),
				),
			},
			&cli.StringSliceFlag{
				Name:  "exclude-dirs",
				Usage: "exclude dirs from analyze",
				Sources: cli.NewValueSourceChain(
					yamlsrc.YAML("exclude_dirs", altsrc.NewStringPtrSourcer(&configFile)),
				),
			},
			&cli.StringFlag{
				Name:  "git-branch",
				Usage: "check only files changed against specified git branch",
			},
		},
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			switch {
			case c.Bool("verbose"):
				logLevel.Set(slog.LevelInfo)
			case c.Bool("debug"):
				logLevel.Set(slog.LevelDebug)
			default:
				logLevel.Set(slog.LevelWarn)
			}
			return ctx, nil
		},
		Commands: []*cli.Command{
			listFilesCommand,
			listMutatorsCommand,
			printASTCommand,
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			return executeMutesting(ctx, options{
				args:                 c.Args().Slice(),
				disableMutators:      c.StringSlice("disable"),
				blacklist:            c.StringSlice("blacklist"),
				match:                c.String("match"),
				silentMode:           c.Bool("silent-mode"),
				doNotRemoveTmpFolder: c.Bool("do-not-remove-tmp-folder"),
				exec:                 c.String("exec"),
				noExec:               c.Bool("no-exec"),
				execTimeout:          c.Uint("exec-timeout"),
				importingOpts: importing.Options{
					SkipFileWithoutTest:  c.Bool("skip-without-test"),
					SkipFileWithBuildTag: c.Bool("skip-with-build-tags"),
					GitMainBranch:        c.String("git-branch"),
					ExcludeDirs:          c.StringArgs("exclude-dirs"),
				},
				debug:   c.Bool("debug"),
				verbose: c.Bool("verbose"),
			})
		},
	}).Run(ctx, os.Args); err != nil {
		log.Println(err)
	}
}

type options struct {
	args                 []string
	importingOpts        importing.Options
	disableMutators      []string
	blacklist            []string
	match                string
	silentMode           bool
	testRecursive        bool
	doNotRemoveTmpFolder bool
	exec                 string
	noExec               bool
	execTimeout          uint
	debug                bool
	verbose              bool
}

func executeMutesting(ctx context.Context, opts options) error {
	var mutationBlackList = map[string]struct{}{}

	files, err := importing.FilesOfArgs(opts.args, opts.importingOpts)
	if err != nil {
		return fmt.Errorf("load packages: %w", err)
	}
	if len(files) == 0 {
		return errors.New("could not find any suitable Go source files")
	}

	if len(opts.blacklist) > 0 {
		for _, f := range opts.blacklist {
			c, err := os.ReadFile(f)
			if err != nil {
				return fmt.Errorf("read blacklist file %q: %w", f, err)
			}

			for line := range strings.SplitSeq(string(c), "\n") {
				if line == "" {
					continue
				}

				if len(line) != 32 {
					return fmt.Errorf("%q is not a MD5 checksum", line)
				}

				mutationBlackList[line] = struct{}{}
			}
		}
	}

	var mutators []mutatorItem

MUTATOR:
	for _, name := range mutator.List() {
		if len(opts.disableMutators) > 0 {
			for _, d := range opts.disableMutators {
				if ok, _ := filepath.Match(d, name); ok {
					continue MUTATOR
				}
			}
		}

		verbose("Enable mutator %q", name)

		m, _ := mutator.New(name)
		mutators = append(mutators, mutatorItem{
			Name:    name,
			Mutator: m,
		})
	}

	tmpDir, err := os.MkdirTemp("", "go-mutesting-")
	if err != nil {
		panic(err)
	}
	verbose("Save mutations into %q", tmpDir)

	var execs []string
	if opts.exec != "" {
		execs = strings.Fields(opts.exec)
	}

	report := &models.Report{}

	for _, file := range files {
		verbose("Mutate %q", file)

		src, fset, pkg, info, err := mutesting.ParseAndTypeCheckFile(file)
		if err != nil {
			return fmt.Errorf("parse file: %w", err)
		}

		err = os.MkdirAll(filepath.Join(tmpDir, filepath.Dir(file)), 0755)
		if err != nil {
			panic(err)
		}

		tmpFile := filepath.Join(tmpDir, file)

		originalFile := fmt.Sprintf("%s.original", tmpFile)
		err = CopyFile(file, originalFile)
		if err != nil {
			panic(err)
		}
		debug("Save original into %q", originalFile)

		mutationID := 0

		if opts.match != "" {
			m, err := regexp.Compile(opts.match)
			if err != nil {
				return fmt.Errorf("match regex is not valid: %w", err)
			}

			for _, f := range astutil.Functions(src) {
				if m.MatchString(f.Name.Name) {
					mutationID = mutate(ctx, opts, mutators, mutationBlackList, mutationID, pkg, info, file, fset, src, f, tmpFile, execs, report)
				}
			}
		} else {
			_ = mutate(ctx, opts, mutators, mutationBlackList, mutationID, pkg, info, file, fset, src, src, tmpFile, execs, report)
		}
	}

	if !opts.doNotRemoveTmpFolder {
		err = os.RemoveAll(tmpDir)
		if err != nil {
			panic(err)
		}
		debug("Remove %q", tmpDir)
	}

	report.Calculate()

	if !opts.noExec {
		if !opts.silentMode {
			fmt.Printf("The mutation score is %f (%d passed, %d failed, %d duplicated, %d skipped, total is %d)\n",
				report.Stats.Msi,
				report.Stats.KilledCount,
				report.Stats.EscapedCount,
				report.Stats.DuplicatedCount,
				report.Stats.SkippedCount,
				report.Stats.TotalMutantsCount,
			)
		}
	} else {
		fmt.Println("Cannot do a mutation testing summary since no exec command was executed.")
	}

	jsonContent, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("marshal report: %w", err)
	}

	file, err := os.OpenFile(models.ReportFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("create report file: %w", err)
	}
	if file == nil {
		return errors.New("cannot create file for report")
	}

	defer func() {
		err = file.Close()
		if err != nil {
			slog.Error(fmt.Sprintf("error white report file closing: %s", err))
		}
	}()

	_, err = file.WriteString(string(jsonContent))
	if err != nil {
		return fmt.Errorf("write report: %w", err)
	}

	verbose("Save report into %q", models.ReportFileName)

	return nil
}

func mutate(
	ctx context.Context,
	opts options,
	mutators []mutatorItem,
	mutationBlackList map[string]struct{},
	mutationID int,
	pkg *types.Package,
	info *types.Info,
	originalFile string,
	fset *token.FileSet,
	src ast.Node,
	node ast.Node,
	mutatedFile string,
	execs []string,
	stats *models.Report,
) int {
	skippedLines := mutesting.Skips(fset, src.(*ast.File))

	for _, m := range mutators {
		debug("Mutator %s", m.Name)

		changed := mutesting.MutateWalk(pkg, info, fset, node, m.Mutator, skippedLines)

		for {
			_, ok := <-changed

			if !ok {
				break
			}

			originalSourceCode, err := os.ReadFile(originalFile)
			if err != nil {
				log.Fatal(err)
			}

			mutant := models.Mutant{}
			mutant.Mutator.MutatorName = m.Name
			mutant.Mutator.OriginalFilePath = originalFile
			mutant.Mutator.OriginalSourceCode = string(originalSourceCode)

			mutationFile := fmt.Sprintf("%s.%d", mutatedFile, mutationID)
			checksum, duplicate, err := saveAST(mutationBlackList, mutationFile, fset, src)
			if err != nil {
				fmt.Printf("INTERNAL ERROR %s\n", err.Error())
			} else if duplicate {
				debug("%q is a duplicate, we ignore it", mutationFile)

				stats.Stats.DuplicatedCount++
			} else {
				debug("Save mutation into %q with checksum %s", mutationFile, checksum)

				if !opts.noExec {
					execExitCode := mutateExec(ctx, opts, pkg, originalFile, src, mutationFile, execs, &mutant)

					debug("Exited with %d", execExitCode)

					mutatedSourceCode, err := os.ReadFile(mutationFile)
					if err != nil {
						log.Fatal(err)
					}
					mutant.Mutator.MutatedSourceCode = string(mutatedSourceCode)

					msg := fmt.Sprintf("%q with checksum %s", mutationFile, checksum)

					switch execExitCode {
					case 0: // Tests failed - all ok
						out := fmt.Sprintf("PASS %s\n", msg)
						if !opts.silentMode {
							fmt.Print(out)
						}

						mutant.ProcessOutput = out
						stats.Killed = append(stats.Killed, mutant)
						stats.Stats.KilledCount++
					case 1: // Tests passed
						out := fmt.Sprintf("FAIL %s\n", msg)
						if !opts.silentMode {
							fmt.Print(out)
						}

						mutant.ProcessOutput = out
						stats.Escaped = append(stats.Escaped, mutant)
						stats.Stats.EscapedCount++
					case 2: // Did not compile
						out := fmt.Sprintf("SKIP %s\n", msg)
						if !opts.silentMode {
							fmt.Print(out)
						}

						mutant.ProcessOutput = out
						stats.Stats.SkippedCount++
					default:
						out := fmt.Sprintf("UNKOWN exit code %d for %s\n", execExitCode, msg)
						if !opts.silentMode {
							fmt.Print(out)
						}

						mutant.ProcessOutput = out
						stats.Errored = append(stats.Errored, mutant)
						stats.Stats.ErrorCount++
					}
				}
			}

			changed <- true

			// Ignore original state
			<-changed
			changed <- true

			mutationID++
		}
	}

	return mutationID
}

func mutateExec(
	ctx context.Context,
	opts options,
	pkg *types.Package,
	file string,
	src ast.Node,
	mutationFile string,
	execs []string,
	mutant *models.Mutant,
) (execExitCode int) {
	if len(execs) == 0 {
		debug("Execute built-in exec command for mutation")

		diff, err := exec.Command("diff", "--label=Original", "--label=New", "-u", file, mutationFile).CombinedOutput()
		if err == nil {
			execExitCode = 0
		} else if e, ok := err.(*exec.ExitError); ok {
			execExitCode = e.Sys().(syscall.WaitStatus).ExitStatus()
		} else {
			panic(err)
		}
		if execExitCode != 0 && execExitCode != 1 {
			fmt.Printf("%s\n", diff)

			panic("Could not execute diff on mutation file")
		}

		defer func() {
			_ = os.Rename(file+".tmp", file)
		}()

		err = os.Rename(file, file+".tmp")
		if err != nil {
			panic(err)
		}
		err = CopyFile(mutationFile, file)
		if err != nil {
			panic(err)
		}

		pkgName := pkg.Path()
		if opts.testRecursive {
			pkgName += "/..."
		}

		goTestCmd := exec.Command("go", "test", "-timeout", fmt.Sprintf("%ds", opts.execTimeout), pkgName)
		goTestCmd.Env = os.Environ()

		test, err := goTestCmd.CombinedOutput()
		if err == nil {
			execExitCode = 0
		} else if e, ok := err.(*exec.ExitError); ok {
			execExitCode = e.Sys().(syscall.WaitStatus).ExitStatus()
		} else {
			panic(err)
		}

		slog.Debug(string(test))

		mutant.Diff = string(diff)

		switch execExitCode {
		case 0: // Tests passed -> FAIL
			if !opts.silentMode {
				fmt.Printf("%s\n", diff)
			}

			execExitCode = 1
		case 1: // Tests failed -> PASS
			slog.Debug(string(diff))
			execExitCode = 0
		case 2: // Did not compile -> SKIP
			slog.Info("Mutation did not compile")
			slog.Debug(string(diff))
		default: // Unknown exit code -> SKIP
			if !opts.silentMode {
				fmt.Println("Unknown exit code")
				fmt.Printf("%s\n", diff)
			}
		}

		return execExitCode
	}

	debug("Execute %q for mutation", opts.exec)

	execCommand := exec.CommandContext(ctx, execs[0], execs[1:]...)

	execCommand.Stderr = os.Stderr
	execCommand.Stdout = os.Stdout

	execCommand.Env = append(os.Environ(), []string{
		"MUTATE_CHANGED=" + mutationFile,
		fmt.Sprintf("MUTATE_DEBUG=%t", opts.debug),
		"MUTATE_ORIGINAL=" + file,
		"MUTATE_PACKAGE=" + pkg.Path(),
		fmt.Sprintf("MUTATE_TIMEOUT=%d", opts.execTimeout),
		fmt.Sprintf("MUTATE_VERBOSE=%t", opts.verbose),
	}...)
	if opts.testRecursive {
		execCommand.Env = append(execCommand.Env, "TEST_RECURSIVE=true")
	}

	err := execCommand.Start()
	if err != nil {
		panic(err)
	}

	// TODO timeout here

	err = execCommand.Wait()

	if err == nil {
		execExitCode = 0
	} else if e, ok := err.(*exec.ExitError); ok {
		execExitCode = e.Sys().(syscall.WaitStatus).ExitStatus()
	} else {
		panic(err)
	}

	return execExitCode
}

func saveAST(mutationBlackList map[string]struct{}, file string, fset *token.FileSet, node ast.Node) (string, bool, error) {
	var buf bytes.Buffer

	h := md5.New()

	err := printer.Fprint(io.MultiWriter(h, &buf), fset, node)
	if err != nil {
		return "", false, err
	}

	checksum := fmt.Sprintf("%x", h.Sum(nil))

	if _, ok := mutationBlackList[checksum]; ok {
		return checksum, true, nil
	}

	mutationBlackList[checksum] = struct{}{}

	src, err := format.Source(buf.Bytes())
	if err != nil {
		return "", false, err
	}

	err = os.WriteFile(file, src, 0666)
	if err != nil {
		return "", false, err
	}

	return checksum, false, nil
}

// CopyFile copies a file from src to dst.
//
// Code copied from "github.com/zimmski/osutil". This package fails to compile
// with alpine.
func CopyFile(src string, dst string) (err error) {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		e := s.Close()
		if err == nil {
			err = e
		}
	}()

	d, err := os.Create(dst)
	if err != nil {
		// In case the file is a symlink, we need to remove the file before we can write to it.
		if _, e := os.Lstat(dst); e == nil {
			if e := os.Remove(dst); e != nil {
				return e
			}
			d, err = os.Create(dst)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	defer func() {
		e := d.Close()
		if err == nil {
			err = e
		}
	}()

	_, err = io.Copy(d, s)
	if err != nil {
		return err
	}

	i, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, i.Mode())
}
