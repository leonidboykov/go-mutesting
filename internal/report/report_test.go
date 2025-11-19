package report

import (
	"testing"

	"pgregory.net/rapid"
)

func TestReport_Calculate(t *testing.T) {
	t.Parallel()

	rapid.Check(t, func(t *rapid.T) {
		report := Report{
			Stats: rapid.Custom(statsGenerator).Draw(t, "stats"),
		}
		report.Calculate()
	})
	// rapid does not support explicit examples.
	t.Run("zero case", func(t *testing.T) {
		(&Report{}).Calculate()
	})
}

func statsGenerator(t *rapid.T) Stats {
	return Stats{
		KilledCount:  rapid.Int64().Draw(t, "killedCount"),
		EscapedCount: rapid.Int64().Draw(t, "escapedCount"),
		ErrorCount:   rapid.Int64().Draw(t, "errorCount"),
		SkippedCount: rapid.Int64().Draw(t, "skippedCount"),
	}
}
