package execute

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
)

var (
	// ErrMutationSurvived means that mutation go tests passed after mutation is applied.
	ErrMutationSurvived = errors.New("mutation survived")

	// ErrCompilationError means that mutation is pointless and leads to broken code.
	ErrCompilationError = errors.New("compilation error")
)

// Default executes default go test command and returns is mutation was "killed", i.e. tests failed.
func GoTest(ctx context.Context, pkgName string, recursive bool) error {
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
