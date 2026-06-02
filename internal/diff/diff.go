// Package diff wraps utilities for diff calculation.
package diff

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/pmezard/go-difflib/difflib"
)

// CompareStrings computes a diff for two files.
func CompareStrings(original, mutated, mutatorName string) (string, error) {
	diff, err := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        difflib.SplitLines(original),
		B:        difflib.SplitLines(mutated),
		FromFile: "Original",
		ToFile:   fmt.Sprintf("Mutation: %s", mutatorName),
		Context:  3,
	})
	if err != nil {
		return "", fmt.Errorf("get diff: %w", err)
	}

	return colorDiff(diff), nil
}

func colorDiff(diff string) string {
	var (
		deletedLinesColor = color.New(color.FgRed)
		newLinesColor     = color.New(color.FgGreen)
		diffLinesColor    = color.New(color.FgHiMagenta)
	)

	colorDiff := new(strings.Builder)
	for line := range strings.SplitSeq(diff, "\n") {
		if len(line) > 0 {
			switch line[0] {
			case '-':
				deletedLinesColor.Fprintln(colorDiff, line)
			case '+':
				newLinesColor.Fprintln(colorDiff, line)
			case '@':
				diffLinesColor.Fprintln(colorDiff, line)
			default:
				fmt.Fprintln(colorDiff, line)
			}
		} else {
			fmt.Fprintln(colorDiff, line)
		}
	}

	return colorDiff.String()
}
