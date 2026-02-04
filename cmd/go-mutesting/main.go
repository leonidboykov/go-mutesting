package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
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
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/lmittmann/tint"
	altsrc "github.com/urfave/cli-altsrc/v3"
	yamlsrc "github.com/urfave/cli-altsrc/v3/yaml"
	"github.com/urfave/cli/v3"
	"golang.org/x/tools/go/packages"

	"github.com/leonidboykov/go-mutesting"
	"github.com/leonidboykov/go-mutesting/internal/astutil"
	"github.com/leonidboykov/go-mutesting/internal/diff"
	"github.com/leonidboykov/go-mutesting/internal/execute"
	"github.com/leonidboykov/go-mutesting/internal/importing"
	"github.com/leonidboykov/go-mutesting/internal/report"
	"github.com/leonidboykov/go-mutesting/mutator"
	_ "github.com/leonidboykov/go-mutesting/mutator/arithmetic"
	_ "github.com/leonidboykov/go-mutesting/mutator/branch"
	_ "github.com/leonidboykov/go-mutesting/mutator/expression"
	_ "github.com/leonidboykov/go-mutesting/mutator/loop"
	_ "github.com/leonidboykov/go-mutesting/mutator/numbers"
	_ "github.com/leonidboykov/go-mutesting/mutator/statement"
)

const md5Len = 32

var errMutantsEscaped = errors.New("mutants escaped")

type mutatorItem struct {
	Name    string
	Mutator mutator.Mutator
}

func init() {
	if os.Getenv("FORCE_COLOR") != "" {
		color.NoColor = false
	}
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	logLevel := new(slog.LevelVar)
	logLevel.Set(slog.LevelInfo)
	slog.SetDefault(slog.New(tint.NewHandler(os.Stderr, &tint.Options{
		Level:      logLevel,
		TimeFormat: time.DateTime,
	})))
	slog.SetLogLoggerLevel(slog.LevelDebug)

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
			&cli.BoolFlag{
				Name:  "json-output",
				Usage: "output logs in json format",
				Sources: cli.NewValueSourceChain(
					yamlsrc.YAML("json_output", altsrc.NewStringPtrSourcer(&configFile)),
				),
			},
			&cli.StringFlag{
				Name:  "git-branch",
				Usage: "check only files changed against specified git branch",
			},
			&cli.BoolFlag{
				Name:  "error-on-survivals",
				Usage: "return exit code 1 if there are survived mutations",
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
				exitCodeOnSurvivals: c.Bool("error-on-survivals"),
				debug:               c.Bool("debug"),
				verbose:             c.Bool("verbose"),
				jsonOutput:          c.Bool("json_output"),
			})
		},
	}).Run(ctx, os.Args); err != nil {
		if !errors.Is(err, errMutantsEscaped) {
			slog.Error(err.Error())
		}
		os.Exit(1)
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
	jsonOutput           bool
	exitCodeOnSurvivals  bool
	debug                bool
	verbose              bool
}

func executeMutesting(ctx context.Context, opts options) error {
	rep, err := ExecuteMutesting(ctx, opts)
	if err != nil {
		return fmt.Errorf("execute mutesting: %w", err)
	}

	if !opts.noExec {
		if !opts.silentMode {
			fmt.Println(rep)
		}
	} else {
		fmt.Println("Cannot do a mutation testing summary since no exec command was executed.")
	}

	if opts.jsonOutput {
		if err := rep.WriteToFile(); err != nil {
			return fmt.Errorf("write report: %w", err)
		}
	}

	if opts.jsonOutput {
		if err := rep.WriteToFile(); err != nil {
			return fmt.Errorf("write report file: %w", err)
		}
	}

	if opts.exitCodeOnSurvivals && rep.Stats.EscapedCount > 0 {
		return errMutantsEscaped
	}

	return nil
}

func ExecuteMutesting(ctx context.Context, opts options) (*report.Report, error) {
	var (
		rep               = new(report.Report)
		mutationBlackList = map[string]struct{}{}
	)

	files, err := importing.FilesOfArgs(opts.args, opts.importingOpts)
	if err != nil {
		return nil, fmt.Errorf("load packages: %w", err)
	}
	if len(files) == 0 {
		slog.Warn("could not find any suitable Go source files")
		return nil, nil
	}

	if len(opts.blacklist) > 0 {
		for _, f := range opts.blacklist {
			c, err := os.ReadFile(f)
			if err != nil {
				return nil, fmt.Errorf("read blacklist file %q: %w", f, err)
			}

			for line := range strings.SplitSeq(string(c), "\n") {
				if line == "" {
					continue
				}

				if len(line) < md5Len {
					return nil, fmt.Errorf("%q is not a MD5 checksum", line)
				}

				// Use the first 32 chars. Everything else is considered as a comment.
				mutationBlackList[line[:md5Len]] = struct{}{}
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

		slog.Info("enable mutator", slog.String("name", name))

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
	slog.Info(fmt.Sprintf("save mutations into %q", tmpDir))

	var execs []string
	if opts.exec != "" {
		execs = strings.Fields(opts.exec)
	}

	for _, file := range files {
		slog.Info("mutate", slog.String("file", file))

		src, pkg, err := importing.ParseAndTypeCheckFile(file)
		if err != nil {
			return rep, fmt.Errorf("parse file: %w", err)
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
		log.Printf("Save original into %q", originalFile)

		mutationID := 0

		if opts.match != "" {
			m, err := regexp.Compile(opts.match)
			if err != nil {
				return rep, fmt.Errorf("match regex is not valid: %w", err)
			}

			for _, f := range astutil.Functions(src) {
				if m.MatchString(f.Name.Name) {
					mutationID = mutate(ctx, opts, mutators, mutationBlackList, mutationID, pkg, file, src, f, tmpFile, execs, rep)
				}
			}
		} else {
			_ = mutate(ctx, opts, mutators, mutationBlackList, mutationID, pkg, file, src, src, tmpFile, execs, rep)
		}
	}

	if !opts.doNotRemoveTmpFolder {
		err = os.RemoveAll(tmpDir)
		if err != nil {
			panic(err)
		}
		log.Printf("Remove %q", tmpDir)
	}

	rep.Calculate()

	return rep, nil
}

func mutate(
	ctx context.Context,
	opts options,
	mutators []mutatorItem,
	mutationBlackList map[string]struct{},
	mutationID int,
	pkg *packages.Package,
	originalFile string,
	src *ast.File,
	node ast.Node,
	mutatedFile string,
	execs []string,
	stats *report.Report,
) int {
	skippedLines := importing.Skips(pkg.Fset, src)

	for _, m := range mutators {
		log.Printf("Mutator %s", m.Name)

		mutesting.MutateWalk(pkg, node, m.Mutator, skippedLines, func() {
			originalSourceCode, err := os.ReadFile(originalFile)
			if err != nil {
				log.Fatal(err)
			}

			mutant := report.Mutant{}
			mutant.Mutator.MutatorName = m.Name
			mutant.Mutator.OriginalFilePath = originalFile
			mutant.Mutator.OriginalSourceCode = string(originalSourceCode)

			mutationFile := fmt.Sprintf("%s.%d", mutatedFile, mutationID)
			checksum, duplicate, err := saveAST(mutationBlackList, mutationFile, pkg.Fset, src)
			if err != nil {
				fmt.Printf("INTERNAL ERROR %s\n", err.Error())
			} else if duplicate {
				log.Printf("%q is a duplicate, we ignore it", mutationFile)

				stats.Stats.DuplicatedCount++
			} else {
				log.Printf("Save mutation into %q with checksum %s", mutationFile, checksum)

				if !opts.noExec {
					mutationError := mutateExec(ctx, opts, pkg.Types, originalFile, mutationFile, execs, &mutant)

					if mutationError != nil {
						slog.Info("exec mutation", slog.Any("error", mutationError))
					}

					mutatedSourceCode, err := os.ReadFile(mutationFile)
					if err != nil {
						log.Fatal(err)
					}
					mutant.Mutator.MutatedSourceCode = string(mutatedSourceCode)

					msg := fmt.Sprintf("%q with checksum %s", mutationFile, checksum)

					switch {
					case mutationError == nil: // Tests failed - all ok
						out := fmt.Sprintf("PASS %s\n", msg)
						if !opts.silentMode {
							fmt.Println(color.GreenString("✓ PASS"), msg)
						}

						mutant.ProcessOutput = out
						stats.Killed = append(stats.Killed, mutant)
						stats.Stats.KilledCount++
					case errors.Is(mutationError, execute.ErrMutationSurvived): // Tests passed
						out := fmt.Sprintf("FAIL %s\n", msg)
						if !opts.silentMode {
							fmt.Println(color.RedString("✗ FAIL"), msg)
						}

						mutant.ProcessOutput = out
						stats.Escaped = append(stats.Escaped, mutant)
						stats.Stats.EscapedCount++
					case errors.Is(mutationError, execute.ErrCompilationError),
						errors.Is(mutationError, context.DeadlineExceeded): // Did not compile
						out := fmt.Sprintf("SKIP %s\n", msg)
						if !opts.silentMode {
							fmt.Println("~ SKIP", msg)
						}

						mutant.ProcessOutput = out
						stats.Stats.SkippedCount++
					case errors.Is(mutationError, context.Canceled): // Cancel
						slog.Warn("cancel signal received, exiting now")
						os.Exit(1)
					default:
						out := fmt.Sprintf("UNKOWN exit code for %s: %s\n", msg, mutationError)
						if !opts.silentMode {
							fmt.Print(out)
						}

						mutant.ProcessOutput = out
						stats.Errored = append(stats.Errored, mutant)
						stats.Stats.ErrorCount++
					}
				}
			}
			mutationID++
		}, func() {})
	}

	return mutationID
}

func mutateExec(
	ctx context.Context,
	opts options,
	pkg *types.Package,
	file string,
	mutationFile string,
	execs []string,
	mutant *report.Mutant,
) error {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(opts.execTimeout)*time.Second)
	defer cancel()

	if len(execs) == 0 {
		log.Printf("Execute built-in exec command for mutation")

		diffStr, err := diff.CompareFiles(file, mutationFile, mutant.Mutator.MutatorName)
		if err != nil {
			panic(err) // TODO: Do not panic on every error.
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

		err = execute.GoTest(ctx, pkg.Path(), opts.testRecursive)

		mutant.Diff = string(diffStr)

		switch err {
		case execute.ErrMutationSurvived: // Tests passed -> FAIL
			if !opts.silentMode {
				fmt.Println(diffStr)
			}
		case nil: // Tests failed -> PASS
			if opts.debug {
				fmt.Println(diffStr)
			}
		case execute.ErrCompilationError: // Did not compile -> SKIP
			slog.Info("Mutation did not compile")
			if opts.debug {
				fmt.Println(diffStr)
			}

		default: // Unknown exit code -> SKIP
			if !opts.silentMode {
				fmt.Println(diffStr)
			}
		}

		return err
	}

	log.Printf("Execute %q for mutation", opts.exec)

	if err := execute.Custom(ctx, execs, execute.CustomMutationOptions{
		Changed:       mutationFile,
		Debug:         opts.debug,
		Original:      file,
		Package:       pkg.Path(),
		Timeout:       opts.execTimeout,
		Verbose:       opts.verbose,
		TestRecursive: opts.testRecursive,
	}); err != nil {
		return fmt.Errorf("execute custom command %q: %w", execs[0], err)
	}

	return nil
}

func saveAST(mutationBlackList map[string]struct{}, file string, fset *token.FileSet, node ast.Node) (string, bool, error) {
	var buf bytes.Buffer

	h := md5.New()

	err := printer.Fprint(io.MultiWriter(h, &buf), fset, node)
	if err != nil {
		return "", false, err
	}

	checksum := hex.EncodeToString(h.Sum(nil))

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
