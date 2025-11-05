package main

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"

	"github.com/leonidboykov/go-mutesting"
	"github.com/leonidboykov/go-mutesting/internal/importing"
	"github.com/leonidboykov/go-mutesting/mutator"
)

var listFilesCommand = &cli.Command{
	Name:  "list-files",
	Usage: "List found files",
	Action: func(ctx context.Context, c *cli.Command) error {
		files, err := importing.FilesOfArgs(c.Args().Slice(), importing.Options{
			SkipFileWithoutTest:  c.Bool("skip-without-test"),
			SkipFileWithBuildTag: c.Bool("skip-with-build-tags"),
			GitMainBranch:        c.String("git-branch"),
			ExcludeDirs:          c.StringArgs("exclude-dirs"),
		})
		if err != nil {
			return fmt.Errorf("import files: %w", err)
		}
		for _, file := range files {
			fmt.Println(file)
		}
		return nil
	},
}

var listMutatorsCommand = &cli.Command{
	Name:  "list-mutators",
	Usage: "List all available mutators (including disabled)",
	Action: func(ctx context.Context, c *cli.Command) error {
		for _, name := range mutator.List() {
			fmt.Println(name)
		}
		return nil
	},
}

var printASTCommand = &cli.Command{
	Name:  "print-ast",
	Usage: "Print the ASTs of all given files and exit",
	Action: func(ctx context.Context, c *cli.Command) error {
		files, err := importing.FilesOfArgs(c.Args().Slice(), importing.Options{
			SkipFileWithoutTest:  c.Bool("skip-without-test"),
			SkipFileWithBuildTag: c.Bool("skip-with-build-tags"),
			GitMainBranch:        c.String("git-branch"),
			ExcludeDirs:          c.StringArgs("exclude-dirs"),
		})
		if err != nil {
			return fmt.Errorf("import files: %w", err)
		}
		for _, file := range files {
			fmt.Println(file)
			src, _, err := importing.ParseFile(file)
			if err != nil {
				return fmt.Errorf("parse file %q: %v", file, err)
			}

			mutesting.PrintWalk(src)

			fmt.Println()
		}
		return nil
	},
}
