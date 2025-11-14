package importing

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFilesOfArgs(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name   string
		args   []string
		expect []string
	}{
		{
			name: "empty",
			args: []string{},
			expect: []string{
				"go-mutesting/internal/importing/filepath.go",
				"go-mutesting/internal/importing/parse.go",
			},
		},
		{
			name:   "files",
			args:   []string{"./filepathfixtures/first.go"},
			expect: []string{"go-mutesting/internal/importing/filepathfixtures/first.go"},
		},
		{
			name: "directories",
			args: []string{"./filepathfixtures"},
			expect: []string{
				"go-mutesting/internal/importing/filepathfixtures/first.go",
				"go-mutesting/internal/importing/filepathfixtures/second.go",
				"go-mutesting/internal/importing/filepathfixtures/third.go",
			},
		},
		{
			name: "relative directories",
			args: []string{"../importing/filepathfixtures"},
			expect: []string{
				"go-mutesting/internal/importing/filepathfixtures/first.go",
				"go-mutesting/internal/importing/filepathfixtures/second.go",
				"go-mutesting/internal/importing/filepathfixtures/third.go",
			},
		},
		{
			name: "packages",
			args: []string{"github.com/leonidboykov/go-mutesting/internal/importing/filepathfixtures"},
			expect: []string{
				"go-mutesting/internal/importing/filepathfixtures/first.go",
				"go-mutesting/internal/importing/filepathfixtures/second.go",
				"go-mutesting/internal/importing/filepathfixtures/third.go",
			},
		},
		{
			name: "packages recursive",
			args: []string{"github.com/leonidboykov/go-mutesting/internal/importing/filepathfixtures/..."},
			expect: []string{
				"go-mutesting/internal/importing/filepathfixtures/first.go",
				"go-mutesting/internal/importing/filepathfixtures/second.go",
				"go-mutesting/internal/importing/filepathfixtures/third.go",
				"go-mutesting/internal/importing/filepathfixtures/secondfixturespackage/fourth.go",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, _ := FilesOfArgs(test.args, Options{})

			assert.Equal(t, test.expect, cleanupPaths(t, got), fmt.Sprintf("With args: %#v", test.args))
		})
	}
}

func TestFilesWithSkipWithoutTests(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name   string
		args   []string
		expect []string
	}{
		{
			name:   "files without tests",
			args:   []string{"./filepathfixtures/first.go"},
			expect: []string{},
		},
		{
			name:   "files with test",
			args:   []string{"./filepathfixtures/second.go"},
			expect: []string{"go-mutesting/internal/importing/filepathfixtures/second.go"},
		},
		{
			name: "directories",
			args: []string{"./filepathfixtures"},
			expect: []string{
				"go-mutesting/internal/importing/filepathfixtures/second.go",
				"go-mutesting/internal/importing/filepathfixtures/third.go",
			},
		},
		{
			name: "packages",
			args: []string{"github.com/leonidboykov/go-mutesting/internal/importing/filepathfixtures/..."},
			expect: []string{
				"go-mutesting/internal/importing/filepathfixtures/second.go",
				"go-mutesting/internal/importing/filepathfixtures/third.go",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, _ := FilesOfArgs(test.args, Options{SkipFileWithoutTest: true})

			assert.Equal(t, test.expect, cleanupPaths(t, got), fmt.Sprintf("With args: %#v", test.args))
		})
	}
}

func TestFilesWithSkipWithBuildTagsTests(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name   string
		args   []string
		expect []string
	}{
		{
			name:   "files without test file",
			args:   []string{"./filepathfixtures/first.go"},
			expect: []string{},
		},
		{
			name:   "files with build tag",
			args:   []string{"./filepathfixtures/third.go"},
			expect: []string{},
		},
		{
			name:   "files without build tag",
			args:   []string{"./filepathfixtures/second.go"},
			expect: []string{"go-mutesting/internal/importing/filepathfixtures/second.go"},
		},
		{
			name:   "directories",
			args:   []string{"./filepathfixtures"},
			expect: []string{"go-mutesting/internal/importing/filepathfixtures/second.go"},
		},
		{
			name: "packages",
			args: []string{"github.com/leonidboykov/go-mutesting/internal/importing/filepathfixtures/..."},
			expect: []string{
				"go-mutesting/internal/importing/filepathfixtures/second.go",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, _ := FilesOfArgs(test.args, Options{SkipFileWithBuildTag: true})

			assert.Equal(t, test.expect, cleanupPaths(t, got), fmt.Sprintf("With args: %#v", test.args))
		})
	}
}

func TestFilesWithExcludedDirs(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name   string
		args   []string
		expect []string
		config []string
	}{
		{
			name:   "files without excludes",
			args:   []string{"./filepathfixtures/first.go"},
			expect: []string{"go-mutesting/internal/importing/filepathfixtures/first.go"},
			config: []string(nil),
		},
		{
			name:   "files with exclude",
			args:   []string{"./filepathfixtures/second.go"},
			expect: []string{},
			config: []string{"filepathfixtures"},
		},
		{
			name: "directories with subdir exclude",
			args: []string{"./filepathfixtures/..."},
			expect: []string{
				"go-mutesting/internal/importing/filepathfixtures/first.go",
				"go-mutesting/internal/importing/filepathfixtures/second.go",
				"go-mutesting/internal/importing/filepathfixtures/third.go",
			},
			config: []string{"filepathfixtures/secondfixturespackage"},
		},
		{
			name:   "directories with exclude - recursive",
			args:   []string{"./filepathfixtures/..."},
			expect: []string{},
			config: []string{"filepathfixtures"},
		},
		{
			name:   "directories with exclude",
			args:   []string{"./filepathfixtures"},
			expect: []string{},
			config: []string{"filepathfixtures"},
		},
		{
			name: "directories without exclude",
			args: []string{"./filepathfixtures"},
			expect: []string{
				"go-mutesting/internal/importing/filepathfixtures/first.go",
				"go-mutesting/internal/importing/filepathfixtures/second.go",
				"go-mutesting/internal/importing/filepathfixtures/third.go",
			},
			config: []string(nil),
		},
		{
			name:   "packages with exclude",
			args:   []string{"github.com/leonidboykov/go-mutesting/internal/importing/filepathfixtures/..."},
			expect: []string{}, // TODO: Breaking change.
			config: []string{"filepathfixtures"},
		},
		{
			name: "packages with specific exclude",
			args: []string{"github.com/leonidboykov/go-mutesting/internal/importing/filepathfixtures/..."},
			expect: []string{
				"go-mutesting/internal/importing/filepathfixtures/first.go",
				"go-mutesting/internal/importing/filepathfixtures/second.go",
				"go-mutesting/internal/importing/filepathfixtures/third.go",
			},
			config: []string{"filepathfixtures/secondfixturespackage"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, _ := FilesOfArgs(test.args, Options{ExcludeDirs: test.config})

			assert.Equal(t, test.expect, cleanupPaths(t, got), fmt.Sprintf("With args: %#v", test.args))
		})
	}
}

// cleanupPaths is a helper to cleanup absolute paths.
func cleanupPaths(t *testing.T, paths []string) []string {
	t.Helper()
	root, err := filepath.Abs("./../../..")
	require.NoError(t, err, "project root")
	result := make([]string, 0, len(paths))
	for _, p := range paths {
		result = append(result, strings.TrimPrefix(p, root+string(filepath.Separator)))
	}
	return result
}
