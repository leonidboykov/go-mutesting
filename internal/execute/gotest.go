package execute

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"

	"github.com/leonidboykov/go-mutesting/internal/diff"
	"github.com/leonidboykov/go-mutesting/internal/report"
)

var (
	// ErrMutationSurvived means that mutation go tests passed after mutation is applied.
	ErrMutationSurvived = errors.New("mutation survived")

	// ErrCompilationError means that mutation is pointless and leads to broken code.
	ErrCompilationError = errors.New("compilation error")
)

type GoTestOptions struct {
	Changed       string
	Original      string
	PackagePath   string
	Debug         bool
	SilentMode    bool
	TestRecursive bool
}

func GoTest(ctx context.Context, mutant *report.Mutant, opts GoTestOptions) error {
	diffStr, err := diff.CompareFiles(opts.Original, opts.Changed, mutant.Mutator.MutatorName)
	if err != nil {
		panic(err) // TODO: Do not panic on every error.
	}

	defer func() {
		_ = os.Rename(opts.Original+".tmp", opts.Original)
	}()

	err = os.Rename(opts.Original, opts.Original+".tmp")
	if err != nil {
		panic(err)
	}
	err = CopyFile(opts.Changed, opts.Original)
	if err != nil {
		panic(err)
	}

	err = runGoTest(ctx, opts.PackagePath, opts.TestRecursive)

	mutant.Diff = string(diffStr)

	switch err {
	case ErrMutationSurvived: // Tests passed -> FAIL
		if !opts.SilentMode {
			fmt.Println(diffStr)
		}
	case nil: // Tests failed -> PASS
		if opts.Debug {
			fmt.Println(diffStr)
		}
	case ErrCompilationError: // Did not compile -> SKIP
		slog.Info("Mutation did not compile")
		if opts.Debug {
			fmt.Println(diffStr)
		}

	default: // Unknown exit code -> SKIP
		if !opts.SilentMode {
			fmt.Println(diffStr)
		}
	}

	return err
}

// GoTest executes default go test command and returns is mutation was "killed", i.e. tests failed.
func runGoTest(ctx context.Context, pkgName string, recursive bool) error {
	if recursive {
		pkgName += "/..."
	}

	// The use of flag `-count=1` prevents from using testcache.
	cmd := exec.CommandContext(ctx, "go", "test", "-count", "1", pkgName)
	cmd.Env = os.Environ() // Is is necessary?

	output, err := cmd.CombinedOutput()

	if slog.Default().Enabled(ctx, slog.LevelDebug) {
		fmt.Fprintln(os.Stderr, string(output))
	}

	if err == nil {
		// No errors, mutaton survived.
		return ErrMutationSurvived
	}

	// Checking error from context is easier that deal with exotic exit codes.
	if err := ctx.Err(); err != nil {
		return err
	}

	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		switch exitError.ExitCode() {
		case 1:
			// Test failed and mutation is killed.
			return nil
		case 2:
			// Most probably a compilation error.
			return ErrCompilationError
		}
	}

	// Unknown error.
	return err
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
