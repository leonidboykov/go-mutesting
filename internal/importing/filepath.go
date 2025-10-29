package importing

/*

This file holds lots of code of the golint project https://github.com/golang/lint and some code of a pull request of mine https://github.com/golang/lint/pull/76
This is just temporary until I have time to clean up this code and make a more general solution for go-commands as I stated here https://github.com/kisielk/errcheck/issues/45#issuecomment-57732642

so TODO and FIXME. Heck I also give you a WORKAROUND.

*/

import (
	"fmt"
	"iter"
	"log"
	"log/slog"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
	"golang.org/x/tools/go/packages"
)

type Options struct {
	SkipFileWithoutTest  bool
	SkipFileWithBuildTag bool
	GitMainBranch        string
	ExcludeDirs          []string
}

func FilesOfArgs(args []string, opts Options) ([]string, error) {
	if len(args) == 0 {
		args = []string{"."}
	}
	pkgs, err := packages.Load(&packages.Config{
		Mode:  packages.LoadFiles,
		Tests: false,
	}, args...)
	if err != nil {
		return nil, fmt.Errorf("load packages: %w", err)
	}

	var gitChangedFiles []string
	if opts.GitMainBranch != "" {
		var err error
		gitChangedFiles, err = getChangedFilesFromGit(opts.GitMainBranch)
		if err != nil {
			return nil, fmt.Errorf("get git changed files: %w", err)
		}
	}

	var files []string
	for _, p := range pkgs {
		iter := skipExcludedDirs(
			removeDuplicates(slices.Values(p.GoFiles)),
			opts.ExcludeDirs,
		)
		if opts.GitMainBranch != "" {
			iter = skipUnchangedFiles(iter, gitChangedFiles)
		}
		if opts.SkipFileWithoutTest || opts.SkipFileWithBuildTag {
			iter = skipFilesWithoutTests(iter)
			if opts.SkipFileWithBuildTag {
				iter = skipFilesWithBuildTag(iter)
			}
		}
		files = append(files, slices.Collect(iter)...)
	}
	return files, nil
}

func removeDuplicates(files iter.Seq[string]) iter.Seq[string] {
	m := make(map[string]struct{})
	return func(yield func(string) bool) {
		for filename := range files {
			if _, ok := m[filename]; ok {
				continue
			}
			m[filename] = struct{}{}
			if !yield(filename) {
				return
			}
		}
	}
}

func skipExcludedDirs(files iter.Seq[string], excludedDirs []string) iter.Seq[string] {
	return func(yield func(string) bool) {
	OUTER_LOOP:
		for filename := range files {
			for _, dir := range excludedDirs {
				// TODO: Replace with [strings.HasPrefix] in case of errors.
				if strings.Contains(filename, dir) {
					continue OUTER_LOOP
				}
			}
			if !yield(filename) {
				return
			}
		}
	}
}

func skipFilesWithoutTests(files iter.Seq[string]) iter.Seq[string] {
	const extLen = len(".go")
	return func(yield func(string) bool) {
		for filename := range files {
			nameSize := len(filename)
			if nameSize <= 3 {
				continue
			}
			testFileName := filename[:nameSize-extLen] + "_test.go"
			if !exists(testFileName) {
				continue
			}
			if !yield(filename) {
				return
			}
		}
	}
}

func skipUnchangedFiles(files iter.Seq[string], changedFiles []string) iter.Seq[string] {
	return func(yield func(string) bool) {
		for filename := range files {
			changed := false
			for _, changedFile := range changedFiles {
				if strings.HasSuffix(filename, changedFile) {
					changed = true
					break
				}
			}
			if !changed {
				continue
			}
			if !yield(filename) {
				return
			}
		}
	}
}

func skipFilesWithBuildTag(files iter.Seq[string]) iter.Seq[string] {
	const extLen = len(".go")
	re := regexp.MustCompile(`\+build (.*)(\s+)package`)
	return func(yield func(string) bool) {
		for filename := range files {
			nameSize := len(filename)
			if nameSize <= 3 {
				continue
			}
			testFileName := filename[:nameSize-extLen] + "_test.go"
			if regexpSearchInFile(testFileName, re) {
				continue
			}
			if !yield(filename) {
				return
			}
		}
	}
}

func exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func regexpSearchInFile(file string, re *regexp.Regexp) bool {
	contents, err := os.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}

	return re.MatchString(string(contents))
}

func getChangedFilesFromGit(mainBranch string) ([]string, error) {
	repo, err := git.PlainOpen("./")
	if err != nil {
		return nil, fmt.Errorf("open git repository: %w", err)
	}

	currRef, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("get current ref: %w", err)
	}
	slog.Debug("current ref", slog.String("short", currRef.Name().Short()))
	currCommit, err := repo.CommitObject(currRef.Hash())
	if err != nil {
		return nil, fmt.Errorf("get current commit: %w", err)
	}
	currTree, err := currCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("get current tree: %w", err)
	}

	mainRef, err := repo.Reference(plumbing.NewBranchReferenceName(mainBranch), false)
	if err != nil {
		return nil, fmt.Errorf("get main ref: %w", err)
	}
	mainCommit, err := repo.CommitObject(mainRef.Hash())
	if err != nil {
		return nil, fmt.Errorf("get main commit: %w", err)
	}
	mainTree, err := mainCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("get main tree: %w", err)
	}

	changes, err := object.DiffTree(currTree, mainTree)
	if err != nil {
		return nil, fmt.Errorf("get changes: %w", err)
	}

	var changedFiles []string
	for _, change := range changes {
		if change.To.Name != "" {
			changedFiles = append(changedFiles, change.To.Name)
		}
	}

	return changedFiles, nil
}
