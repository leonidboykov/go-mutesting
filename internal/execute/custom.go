package execute

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

type CustomMutationOptions struct {
	Changed       string
	Original      string
	Package       string
	Debug         bool
	Verbose       bool
	Timeout       uint
	TestRecursive bool
}

// Custom allows to execute custom command.
func Custom(ctx context.Context, execArgs []string, opts CustomMutationOptions) error {
	cmd := exec.CommandContext(ctx, execArgs[0], execArgs[1:]...)
	cmd.Env = append(os.Environ(), []string{
		"MUTATE_CHANGED=" + opts.Changed,
		"MUTATE_ORIGINAL=" + opts.Original,
		"MUTATE_PACKAGE=" + opts.Package,
		"MUTATE_DEBUG=" + strconv.FormatBool(opts.Debug),
		"MUTATE_VERBOSE=" + strconv.FormatBool(opts.Verbose),
		"MUTATE_TIMEOUT=" + strconv.FormatUint(uint64(opts.Timeout), 10),
		"TEST_RECURSIVE=" + strconv.FormatBool(opts.TestRecursive),
	}...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start cmd: %w", err)
	}

	err := cmd.Wait()

	if err == nil {
		// Mutation was killed.
		return nil
	}

	// Checking error from context is easier that deal with exotic exit codes.
	if err := ctx.Err(); err != nil {
		return err
	}

	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		switch exitError.ExitCode() {
		case 1:
			return ErrMutationSurvived
		case 2:
			return ErrCompilationError
		}
	}

	// Unknown error.
	return fmt.Errorf("wait cmd: %w", err)
}
