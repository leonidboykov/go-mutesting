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
	"slices"
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
			suite, err := newSuite(options{
				args:                 c.Args().Slice(),
				disabledMutators:     c.StringSlice("disable"),
				blacklist:            c.StringSlice("blacklist"),
				match:                c.String("match"),
				silentMode:           c.Bool("silent-mode"),
				doNotRemoveTmpFolder: c.Bool("do-not-remove-tmp-folder"),
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
			if err != nil {
				return fmt.Errorf("prepare mutation framework: %w", err)
			}
			return suite.executeMutesting(ctx)
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
	disabledMutators     []string
	blacklist            []string
	match                string
	silentMode           bool
	testRecursive        bool
	doNotRemoveTmpFolder bool
	noExec               bool
	execTimeout          uint
	jsonOutput           bool
	exitCodeOnSurvivals  bool
	debug                bool
	verbose              bool
}

// suite allows to execute mutations.
type suite struct {
	opts      options
	checksums map[string]struct{}
	mutators  []mutatorItem
}

// newSuite creates a new [suite].
func newSuite(opts options) (*suite, error) {
	checksums, err := loadChecksums(opts.blacklist)
	if err != nil {
		return nil, fmt.Errorf("load checksums: %w", err)
	}
	mutators, err := loadMutators(opts.disabledMutators)
	if err != nil {
		return nil, fmt.Errorf("load mutators: %w", err)
	}
	return &suite{
		opts:      opts,
		checksums: checksums,
		mutators:  mutators,
	}, nil
}

// loadChecksums loads files with blacklisted md5 checksums.
func loadChecksums(files []string) (map[string]struct{}, error) {
	checksums := make(map[string]struct{}, len(files))
	for _, f := range files {
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
			checksums[line[:md5Len]] = struct{}{}
		}
	}
	return checksums, nil
}

func loadMutators(disabledMutators []string) ([]mutatorItem, error) {
	var mutators []mutatorItem
	for _, name := range mutator.List() {
		if slices.ContainsFunc(disabledMutators, func(d string) bool {
			ok, _ := filepath.Match(d, name)
			return ok
		}) {
			continue
		}

		slog.Info("enable mutator", slog.String("name", name))

		m, err := mutator.New(name)
		if err != nil {
			return nil, fmt.Errorf("create mutator %q: %w", name, err)
		}
		mutators = append(mutators, mutatorItem{
			Name:    name,
			Mutator: m,
		})
	}
	return mutators, nil
}

func (s *suite) executeMutesting(ctx context.Context) error {
	rep, err := s.ExecuteMutesting(ctx)
	if err != nil {
		return fmt.Errorf("execute mutesting: %w", err)
	}

	if !s.opts.noExec {
		if !s.opts.silentMode {
			fmt.Println(rep)
		}
	} else {
		fmt.Println("Cannot do a mutation testing summary since no exec command was executed.")
	}

	if s.opts.jsonOutput {
		if err := rep.WriteToFile(); err != nil {
			return fmt.Errorf("write report file: %w", err)
		}
	}

	if s.opts.exitCodeOnSurvivals && rep.Stats.EscapedCount > 0 {
		return errMutantsEscaped
	}

	return nil
}

func (s *suite) ExecuteMutesting(ctx context.Context) (*report.Report, error) {
	var rep = new(report.Report)

	files, err := importing.FilesOfArgs(s.opts.args, s.opts.importingOpts)
	if err != nil {
		return nil, fmt.Errorf("load packages: %w", err)
	}
	if len(files) == 0 {
		slog.Warn("could not find any suitable Go source files")
		return rep, nil
	}

	tmpDir, err := os.MkdirTemp("", "go-mutesting-")
	if err != nil {
		return nil, fmt.Errorf("create temp directory: %w", err)
	}
	slog.Info("save mutations", slog.String("dir", tmpDir))

	for _, file := range files {
		slog.Info("mutate", slog.String("file", file))

		src, pkg, err := importing.ParseAndTypeCheckFile(ctx, file)
		if err != nil {
			return rep, fmt.Errorf("parse file: %w", err)
		}

		if err := os.MkdirAll(filepath.Join(tmpDir, filepath.Dir(file)), 0755); err != nil {
			return nil, fmt.Errorf("copy files in temp directory: %w", err)
		}

		var mutationID int

		if s.opts.match != "" {
			m, err := regexp.Compile(s.opts.match)
			if err != nil {
				return nil, fmt.Errorf("match regex is not valid: %w", err)
			}

			for _, f := range astutil.Functions(src) {
				if m.MatchString(f.Name.Name) {
					mutationID = s.mutate(ctx, mutationID, pkg, file, src, f, tmpDir, rep)
				}
			}
		} else {
			_ = s.mutate(ctx, mutationID, pkg, file, src, src, tmpDir, rep)
		}
	}

	if !s.opts.doNotRemoveTmpFolder {
		if err := os.RemoveAll(tmpDir); err != nil {
			return nil, fmt.Errorf("remove temp directory: %w", err)
		}
		slog.Info("remove temp directory", slog.String("dir", tmpDir))
	}

	rep.Calculate()

	return rep, nil
}

func (s *suite) mutate(
	ctx context.Context,
	mutationID int,
	pkg *packages.Package,
	originalFile string,
	src *ast.File,
	node ast.Node,
	tempDir string,
	stats *report.Report,
) int {
	skippedLines := importing.Skips(pkg.Fset, src)

	originalSourceCode, err := os.ReadFile(originalFile)
	if err != nil {
		log.Fatal(err)
	}

	for _, m := range s.mutators {
		log.Printf("Mutator %s", m.Name)

		mutesting.MutateWalk(pkg, node, m.Mutator, skippedLines, func() {
			mutant := report.Mutant{}
			mutant.Mutator.MutatorName = m.Name
			mutant.Mutator.OriginalFilePath = originalFile
			mutant.Mutator.OriginalSourceCode = string(originalSourceCode)

			mutationFile := filepath.Join(tempDir, fmt.Sprintf("%s.%d", originalFile, mutationID))
			checksum, duplicate, err := s.saveAST(mutationFile, pkg.Fset, src)
			if err != nil {
				slog.Error("save ast", slog.String("file", mutationFile), slog.Any("error", err))
			} else if duplicate {
				log.Printf("%q is a duplicate, we ignore it", mutationFile)

				stats.Stats.DuplicatedCount++
			} else {
				log.Printf("Save mutation into %q with checksum %s", mutationFile, checksum)

				if !s.opts.noExec {
					mutationError := s.mutateExec(ctx, pkg.Types, originalFile, mutationFile, &mutant)

					if mutationError != nil {
						slog.Info("exec mutation", slog.Any("error", mutationError))
					}

					mutatedSourceCode, err := os.ReadFile(mutationFile)
					if err != nil {
						log.Fatal(err)
					}
					mutant.Mutator.MutatedSourceCode = string(mutatedSourceCode)

					msg := fmt.Sprintf("%q #%d with checksum %s", originalFile, mutationID, checksum)

					switch {
					case mutationError == nil: // Tests failed - all ok
						out := fmt.Sprintf("PASS %s\n", msg)
						if !s.opts.silentMode {
							fmt.Println(color.GreenString("✓ PASS"), msg)
						}

						mutant.ProcessOutput = out
						stats.Killed = append(stats.Killed, mutant)
						stats.Stats.KilledCount++
					case errors.Is(mutationError, execute.ErrMutationSurvived): // Tests passed
						out := fmt.Sprintf("FAIL %s\n", msg)
						if !s.opts.silentMode {
							fmt.Println(color.RedString("✗ FAIL"), msg)
						}

						mutant.ProcessOutput = out
						stats.Escaped = append(stats.Escaped, mutant)
						stats.Stats.EscapedCount++
					case errors.Is(mutationError, execute.ErrCompilationError),
						errors.Is(mutationError, context.DeadlineExceeded): // Did not compile
						out := fmt.Sprintf("SKIP %s\n", msg)
						if !s.opts.silentMode {
							fmt.Println("~ SKIP", msg)
						}

						mutant.ProcessOutput = out
						stats.Stats.SkippedCount++
					case errors.Is(mutationError, context.Canceled): // Cancel
						slog.Warn("cancel signal received, exiting now")
						os.Exit(1)
					default:
						out := fmt.Sprintf("UNKOWN exit code for %s: %s\n", msg, mutationError)
						if !s.opts.silentMode {
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

func (s *suite) mutateExec(
	ctx context.Context,
	pkg *types.Package,
	file string,
	mutationFile string,
	mutant *report.Mutant,
) error {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(s.opts.execTimeout)*time.Second)
	defer cancel()

	log.Printf("Execute built-in exec command for mutation")

	return execute.GoTest(ctx, mutant, execute.GoTestOptions{
		Changed:       mutationFile,
		Original:      file,
		PackagePath:   pkg.Path(),
		Debug:         s.opts.debug,
		SilentMode:    s.opts.silentMode,
		TestRecursive: s.opts.testRecursive,
	})
}

func (s *suite) saveAST(file string, fset *token.FileSet, node ast.Node) (string, bool, error) {
	var buf bytes.Buffer

	h := md5.New()

	err := printer.Fprint(io.MultiWriter(h, &buf), fset, node)
	if err != nil {
		return "", false, err
	}

	checksum := hex.EncodeToString(h.Sum(nil))

	if _, ok := s.checksums[checksum]; ok {
		return checksum, true, nil
	}

	s.checksums[checksum] = struct{}{}

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
