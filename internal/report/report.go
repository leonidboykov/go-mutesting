package report

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
)

// ReportFileName File name for json report
var ReportFileName string = "report.json"

// Report Structure for mutation report
type Report struct {
	Stats     Stats    `json:"stats"`
	Escaped   []Mutant `json:"escaped"`
	Timeouted []Mutant `json:"timeouted"`
	Killed    []Mutant `json:"killed"`
	Errored   []Mutant `json:"errored"`
}

// Stats There is stats for mutations
type Stats struct {
	TotalMutantsCount    int64   `json:"totalMutantsCount"`
	KilledCount          int64   `json:"killedCount"`
	NotCoveredCount      int64   `json:"notCoveredCount"`
	EscapedCount         int64   `json:"escapedCount"`
	ErrorCount           int64   `json:"errorCount"`
	SkippedCount         int64   `json:"skippedCount"`
	TimeOutCount         int64   `json:"timeOutCount"`
	Msi                  float64 `json:"msi"`
	MutationCodeCoverage int64   `json:"mutationCodeCoverage"`
	CoveredCodeMsi       float64 `json:"coveredCodeMsi"`
	DuplicatedCount      int64   `json:"-"`
}

// Mutant report by mutant for one mutation on one file
type Mutant struct {
	Mutator       Mutator `json:"mutator"`
	Diff          string  `json:"diff"`
	ProcessOutput string  `json:"processOutput,omitempty"`
}

// Mutator mutator and changes in file
type Mutator struct {
	MutatorName        string `json:"mutatorName"`
	OriginalSourceCode string `json:"originalSourceCode"`
	MutatedSourceCode  string `json:"mutatedSourceCode"`
	OriginalFilePath   string `json:"originalFilePath"`
	OriginalStartLine  int64  `json:"originalStartLine"`
}

// Calculate calculation for final report
func (r *Report) Calculate() {
	r.Stats.Msi = r.MsiScore()
	r.Stats.TotalMutantsCount = r.TotalCount()
}

// MsiScore msi score calculation
func (r *Report) MsiScore() float64 {
	total := r.TotalCount()

	if total == 0 {
		return 0.0
	}

	return float64(r.Stats.KilledCount+r.Stats.ErrorCount+r.Stats.SkippedCount) / float64(total)
}

// TotalCount total mutations count
func (r *Report) TotalCount() int64 {
	return r.Stats.KilledCount + r.Stats.EscapedCount + r.Stats.ErrorCount + r.Stats.SkippedCount
}

// String implements [fmt.Stringer] interface.
func (r *Report) String() string {
	return fmt.Sprintf("The mutation score is %f (%d passed, %d failed, %d duplicated, %d skipped, total is %d)",
		r.Stats.Msi,
		r.Stats.KilledCount,
		r.Stats.EscapedCount,
		r.Stats.DuplicatedCount,
		r.Stats.SkippedCount,
		r.Stats.TotalMutantsCount,
	)
}

// WriteToFile writes report file to [ReportFileName].
func (r *Report) WriteToFile() error {
	file, err := os.OpenFile(ReportFileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("creaate file: %w", err)
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(r); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}

	slog.Info("save report file", slog.String("name", ReportFileName))

	return nil
}
