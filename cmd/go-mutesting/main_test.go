package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/leonidboykov/go-mutesting/internal/importing"
	"github.com/leonidboykov/go-mutesting/internal/report"
)

func TestExecuteMutesting(t *testing.T) {
	tt := []struct {
		name          string
		root          string
		opts          options
		expectedStats report.Stats
		expectedErr   string
	}{
		{
			name:          "simple",
			root:          "../../example",
			opts:          options{execTimeout: 10},
			expectedErr:   "",
			expectedStats: report.Stats{Msi: 0.573770, KilledCount: 35, EscapedCount: 26, DuplicatedCount: 7, SkippedCount: 0, TotalMutantsCount: 61},
		},
		{
			name:          "recursive",
			root:          "../../example",
			opts:          options{args: []string{"./..."}, debug: true, execTimeout: 10},
			expectedErr:   "",
			expectedStats: report.Stats{Msi: 0.600000, KilledCount: 39, EscapedCount: 26, DuplicatedCount: 7, SkippedCount: 0, TotalMutantsCount: 65},
		},
		{
			name:          "from other directory",
			root:          "../..",
			opts:          options{args: []string{"github.com/leonidboykov/go-mutesting/example"}, debug: true, execTimeout: 10},
			expectedStats: report.Stats{Msi: 0.573770, KilledCount: 35, EscapedCount: 26, DuplicatedCount: 7, SkippedCount: 0, TotalMutantsCount: 61},
			expectedErr:   "",
		},
		{
			name:          "match",
			root:          "../../example",
			opts:          options{args: []string{"./..."}, debug: true, execTimeout: 10, exec: "../scripts/exec/test-mutated-package.sh", match: "baz"},
			expectedStats: report.Stats{Msi: 0.500000, KilledCount: 4, EscapedCount: 4, DuplicatedCount: 0, SkippedCount: 0, TotalMutantsCount: 8},
			expectedErr:   "",
		},
		{
			name: "skip without tests",
			root: "../../example",
			opts: options{args: []string{}, debug: true, execTimeout: 10, importingOpts: importing.Options{
				SkipFileWithoutTest:  true,
				SkipFileWithBuildTag: true,
			}},
			expectedStats: report.Stats{Msi: 0.583333, KilledCount: 35, EscapedCount: 25, DuplicatedCount: 7, SkippedCount: 0, TotalMutantsCount: 60},
			expectedErr:   "",
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			saveCwd, err := os.Getwd()
			require.NoError(t, err)
			assert.NoError(t, os.Chdir(tc.root))
			t.Cleanup(func() { os.Chdir(saveCwd) })

			rep, err := ExecuteMutesting(t.Context(), tc.opts)

			assert.InDelta(t, tc.expectedStats.Msi, rep.Stats.Msi, 0.000001)
			assert.Equal(t, tc.expectedStats.KilledCount, rep.Stats.KilledCount)
			assert.Equal(t, tc.expectedStats.EscapedCount, rep.Stats.EscapedCount)
			assert.Equal(t, tc.expectedStats.DuplicatedCount, rep.Stats.DuplicatedCount)
			assert.Equal(t, tc.expectedStats.SkippedCount, rep.Stats.SkippedCount)
			assert.Equal(t, tc.expectedStats.TotalMutantsCount, rep.Stats.TotalMutantsCount)
			if tc.expectedErr == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tc.expectedErr)
			}
		})
	}
}
