package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

type LimitsConfig struct {
	WallClockTimeoutMinutes int `json:"wall_clock_timeout_minutes"`
	MaxOutputTokens         int `json:"max_output_tokens"`
}

type DecisionThresholds struct {
	MinContractAccepted         int     `json:"min_contract_accepted"`
	MaxContractScopeViolations  int     `json:"max_contract_scope_violations"`
	MaxContractEscapedDefects   int     `json:"max_contract_escaped_defects"`
	MinInterventionReductionPct float64 `json:"min_intervention_reduction_pct"`
	MaxPairedWallRatio          float64 `json:"max_paired_wall_ratio"`
	MaxPairedTokenRatio         float64 `json:"max_paired_token_ratio"`
}

type FixtureManifest struct {
	ID                   string `json:"id"`
	SourceCommit         string `json:"source_commit"`
	TaskKind             string `json:"task_kind"`
	OraclePath           string `json:"oracle_path"`
	OracleDigest         string `json:"oracle_digest"`
	BaselinePacketPath   string `json:"baseline_packet_path"`
	BaselinePacketDigest string `json:"baseline_packet_digest"`
	ContractPacketPath   string `json:"contract_packet_path"`
	ContractPacketDigest string `json:"contract_packet_digest"`
}

type PilotManifest struct {
	SchemaVersion      string             `json:"schema_version"`
	BaselineCommit     string             `json:"baseline_commit"`
	CandidateCommit    string             `json:"candidate_commit"`
	Model              string             `json:"model"`
	Harness            string             `json:"harness"`
	Limits             LimitsConfig       `json:"limits"`
	RandomizedOrder    []string           `json:"randomized_order"`
	DecisionThresholds DecisionThresholds `json:"decision_thresholds"`
	Fixtures           []FixtureManifest  `json:"fixtures"`
}

type TrialResult struct {
	SchemaVersion   string   `json:"schema_version"`
	FixtureID       string   `json:"fixture_id"`
	Arm             string   `json:"arm"`
	ModelVersion    string   `json:"model_version"`
	HarnessVersion  string   `json:"harness_version"`
	StartCommit     string   `json:"start_commit"`
	FinalCommit     string   `json:"final_commit"`
	PacketDigest    string   `json:"packet_digest"`
	OracleDigest    string   `json:"oracle_digest"`
	FirstPass       bool     `json:"first_pass"`
	Accepted        bool     `json:"accepted"`
	Interventions   int      `json:"interventions"`
	ScopeViolations int      `json:"scope_violations"`
	EscapedDefects  int      `json:"escaped_defects"`
	WallSeconds     float64  `json:"wall_seconds"`
	InputTokens     int      `json:"input_tokens"`
	OutputTokens    int      `json:"output_tokens"`
	Cost            float64  `json:"cost"`
	CostUnit        string   `json:"cost_unit"`
	EvidencePaths   []string `json:"evidence_paths"`
	Outcome         string   `json:"outcome"`
}

type RunIndex struct {
	SchemaVersion   string        `json:"schema_version"`
	TotalTrials     int           `json:"total_trials"`
	CompletedTrials int           `json:"completed_trials"`
	Runs            []TrialResult `json:"runs"`
}

type SummaryStats struct {
	TotalRuns       int     `json:"total_runs"`
	AcceptedRuns    int     `json:"accepted_runs"`
	FirstPassRuns   int     `json:"first_pass_runs"`
	Interventions   int     `json:"interventions"`
	ScopeViolations int     `json:"scope_violations"`
	EscapedDefects  int     `json:"escaped_defects"`
	TotalWallSecs   float64 `json:"total_wall_seconds"`
	TotalTokens     int     `json:"total_tokens"`
}

type PairedComparison struct {
	FixtureID            string   `json:"fixture_id"`
	BaselineAccepted     bool     `json:"baseline_accepted"`
	ContractAccepted     bool     `json:"contract_accepted"`
	BaselineIntervene    int      `json:"baseline_interventions"`
	ContractIntervene    int      `json:"contract_interventions"`
	WallRatio            *float64 `json:"wall_ratio"`
	TokenRatio           *float64 `json:"token_ratio"`
	ScopeViolationDelta  int      `json:"scope_violation_delta"`
	EscapedDefectDelta   int      `json:"escaped_defect_delta"`
}

type InterventionDelta struct {
	BaselineTotal int      `json:"baseline_total"`
	ContractTotal int      `json:"contract_total"`
	AbsoluteDelta int      `json:"absolute_delta"`
	ReductionPct  *float64 `json:"reduction_pct"`
	Reason        string   `json:"reason,omitempty"`
}

type MedianRatios struct {
	WallRatio  *float64 `json:"wall_ratio"`
	TokenRatio *float64 `json:"token_ratio"`
}

type ThresholdChecks struct {
	ContractAcceptedMet      bool `json:"contract_accepted_met"`
	ScopeViolationsZero      bool `json:"scope_violations_zero"`
	EscapedDefectsZero       bool `json:"escaped_defects_zero"`
	FirstPassNotWorse        bool `json:"first_pass_not_worse"`
	InterventionReductionMet bool `json:"intervention_reduction_met"`
	WallRatioMet             bool `json:"wall_ratio_met"`
	TokenRatioMet            bool `json:"token_ratio_met"`
}

type AnalysisResult struct {
	SchemaVersion       string             `json:"schema_version"`
	Status              string             `json:"status"`
	TotalPairs          int                `json:"total_pairs"`
	TotalRuns           int                `json:"total_runs"`
	BaselineSummary     SummaryStats       `json:"baseline_summary"`
	ContractSummary     SummaryStats       `json:"contract_summary"`
	PairedComparisons   []PairedComparison `json:"paired_comparisons"`
	MedianRatios        MedianRatios       `json:"median_ratios"`
	InterventionDelta   InterventionDelta  `json:"intervention_delta"`
	ThresholdChecks     ThresholdChecks    `json:"threshold_checks"`
	PreliminaryDecision string             `json:"preliminary_decision"`
	Limitations         []string           `json:"limitations"`
}

func fileSHA256(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func RunAnalysis(manifestPath, runsPath string) (*AnalysisResult, error) {
	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("reading manifest: %w", err)
	}

	var manifest PilotManifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return nil, fmt.Errorf("parsing manifest: %w", err)
	}

	runsBytes, err := os.ReadFile(runsPath)
	if err != nil {
		return nil, fmt.Errorf("reading runs index: %w", err)
	}

	var runIndex RunIndex
	if err := json.Unmarshal(runsBytes, &runIndex); err != nil {
		return nil, fmt.Errorf("parsing runs index: %w", err)
	}

	fixtureMap := make(map[string]FixtureManifest)
	for _, fix := range manifest.Fixtures {
		fixtureMap[fix.ID] = fix
	}

	baselineRuns := make(map[string]TrialResult)
	contractRuns := make(map[string]TrialResult)

	for _, run := range runIndex.Runs {
		fix, ok := fixtureMap[run.FixtureID]
		if !ok {
			return nil, fmt.Errorf("run references unknown fixture: %s", run.FixtureID)
		}

		// Verify oracle digest matches
		if run.OracleDigest != fix.OracleDigest {
			return nil, fmt.Errorf("fixture %s: oracle digest mismatch: expected %s, got %s",
				run.FixtureID, fix.OracleDigest, run.OracleDigest)
		}

		// Verify packet digest matches arm
		if run.Arm == "baseline" {
			if run.PacketDigest != fix.BaselinePacketDigest {
				return nil, fmt.Errorf("fixture %s: baseline packet digest mismatch: expected %s, got %s",
					run.FixtureID, fix.BaselinePacketDigest, run.PacketDigest)
			}
			baselineRuns[run.FixtureID] = run
		} else if run.Arm == "contract" {
			if run.PacketDigest != fix.ContractPacketDigest {
				return nil, fmt.Errorf("fixture %s: contract packet digest mismatch: expected %s, got %s",
					run.FixtureID, fix.ContractPacketDigest, run.PacketDigest)
			}
			contractRuns[run.FixtureID] = run
		} else {
			return nil, fmt.Errorf("unknown arm: %s", run.Arm)
		}
	}

	// Verify all fixtures have both baseline and contract runs
	for _, fix := range manifest.Fixtures {
		if _, ok := baselineRuns[fix.ID]; !ok {
			return nil, fmt.Errorf("missing baseline run for fixture %s", fix.ID)
		}
		if _, ok := contractRuns[fix.ID]; !ok {
			return nil, fmt.Errorf("missing contract run for fixture %s", fix.ID)
		}
	}

	var baseSummary, contSummary SummaryStats
	var paired []PairedComparison
	var wallRatios []float64
	var tokenRatios []float64

	for _, fix := range manifest.Fixtures {
		b := baselineRuns[fix.ID]
		c := contractRuns[fix.ID]

		// Update baseline summary
		baseSummary.TotalRuns++
		if b.Accepted {
			baseSummary.AcceptedRuns++
		}
		if b.FirstPass {
			baseSummary.FirstPassRuns++
		}
		baseSummary.Interventions += b.Interventions
		baseSummary.ScopeViolations += b.ScopeViolations
		baseSummary.EscapedDefects += b.EscapedDefects
		baseSummary.TotalWallSecs += b.WallSeconds
		baseSummary.TotalTokens += b.InputTokens + b.OutputTokens

		// Update contract summary
		contSummary.TotalRuns++
		if c.Accepted {
			contSummary.AcceptedRuns++
		}
		if c.FirstPass {
			contSummary.FirstPassRuns++
		}
		contSummary.Interventions += c.Interventions
		contSummary.ScopeViolations += c.ScopeViolations
		contSummary.EscapedDefects += c.EscapedDefects
		contSummary.TotalWallSecs += c.WallSeconds
		contSummary.TotalTokens += c.InputTokens + c.OutputTokens

		// Paired metrics
		comp := PairedComparison{
			FixtureID:           fix.ID,
			BaselineAccepted:    b.Accepted,
			ContractAccepted:    c.Accepted,
			BaselineIntervene:   b.Interventions,
			ContractIntervene:   c.Interventions,
			ScopeViolationDelta: c.ScopeViolations - b.ScopeViolations,
			EscapedDefectDelta:  c.EscapedDefects - b.EscapedDefects,
		}

		if b.WallSeconds > 0 {
			wr := c.WallSeconds / b.WallSeconds
			comp.WallRatio = &wr
			wallRatios = append(wallRatios, wr)
		}

		bTokens := b.InputTokens + b.OutputTokens
		cTokens := c.InputTokens + c.OutputTokens
		if bTokens > 0 {
			tr := float64(cTokens) / float64(bTokens)
			comp.TokenRatio = &tr
			tokenRatios = append(tokenRatios, tr)
		}

		paired = append(paired, comp)
	}

	// Calculate median ratios
	median := func(vals []float64) *float64 {
		if len(vals) == 0 {
			return nil
		}
		sorted := make([]float64, len(vals))
		copy(sorted, vals)
		sort.Float64s(sorted)
		n := len(sorted)
		var m float64
		if n%2 == 1 {
			m = sorted[n/2]
		} else {
			m = (sorted[n/2-1] + sorted[n/2]) / 2.0
		}
		return &m
	}

	medRatios := MedianRatios{
		WallRatio:  median(wallRatios),
		TokenRatio: median(tokenRatios),
	}

	// Calculate intervention delta
	delta := contSummary.Interventions - baseSummary.Interventions
	interDelta := InterventionDelta{
		BaselineTotal: baseSummary.Interventions,
		ContractTotal: contSummary.Interventions,
		AbsoluteDelta: delta,
	}

	if baseSummary.Interventions == 0 {
		interDelta.Reason = "zero_denominator"
	} else {
		pct := (float64(baseSummary.Interventions-contSummary.Interventions) / float64(baseSummary.Interventions)) * 100.0
		interDelta.ReductionPct = &pct
	}

	// Evaluate against thresholds
	thresh := manifest.DecisionThresholds
	checks := ThresholdChecks{
		ContractAcceptedMet: contSummary.AcceptedRuns >= thresh.MinContractAccepted,
		ScopeViolationsZero: contSummary.ScopeViolations <= thresh.MaxContractScopeViolations,
		EscapedDefectsZero:  contSummary.EscapedDefects <= thresh.MaxContractEscapedDefects,
		FirstPassNotWorse:   contSummary.FirstPassRuns >= baseSummary.FirstPassRuns,
	}

	if baseSummary.Interventions > 0 {
		checks.InterventionReductionMet = interDelta.ReductionPct != nil && *interDelta.ReductionPct >= thresh.MinInterventionReductionPct
	} else {
		// Zero baseline interventions requirement: zero contract interventions and strictly better first-pass
		checks.InterventionReductionMet = contSummary.Interventions == 0 && contSummary.FirstPassRuns > baseSummary.FirstPassRuns
	}

	if medRatios.WallRatio != nil {
		checks.WallRatioMet = *medRatios.WallRatio <= thresh.MaxPairedWallRatio
	}
	if medRatios.TokenRatio != nil {
		checks.TokenRatioMet = *medRatios.TokenRatio <= thresh.MaxPairedTokenRatio
	}

	// Preliminary decision derivation
	allPassed := checks.ContractAcceptedMet &&
		checks.ScopeViolationsZero &&
		checks.EscapedDefectsZero &&
		checks.FirstPassNotWorse &&
		checks.InterventionReductionMet &&
		checks.WallRatioMet &&
		checks.TokenRatioMet

	var decision string
	if allPassed {
		decision = "proceed"
	} else if contSummary.ScopeViolations > 0 || contSummary.AcceptedRuns < baseSummary.AcceptedRuns {
		decision = "stop"
	} else {
		decision = "revise"
	}

	result := &AnalysisResult{
		SchemaVersion:       "sila/pilot-analysis/v1",
		Status:              "complete",
		TotalPairs:          len(manifest.Fixtures),
		TotalRuns:           len(manifest.Fixtures) * 2,
		BaselineSummary:     baseSummary,
		ContractSummary:     contSummary,
		PairedComparisons:   paired,
		MedianRatios:        medRatios,
		InterventionDelta:   interDelta,
		ThresholdChecks:     checks,
		PreliminaryDecision: decision,
		Limitations: []string{
			"Small-sample pilot across six paired programming tasks.",
			"Evaluates structural contract presentation; does not extrapolate to general codebase changes.",
			"Results apply to the evaluated model and harness configuration only.",
		},
	}

	return result, nil
}

func main() {
	manifestPath := flag.String("manifest", "", "Path to pilot manifest JSON")
	runsPath := flag.String("runs", "", "Path to runs index JSON")
	outputPath := flag.String("output", "", "Path to output analysis JSON")
	flag.Parse()

	if *manifestPath == "" || *runsPath == "" || *outputPath == "" {
		fmt.Fprintf(os.Stderr, "Usage: analysis --manifest <path> --runs <path> --output <path>\n")
		os.Exit(1)
	}

	result, err := RunAnalysis(*manifestPath, *runsPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running analysis: %v\n", err)
		os.Exit(2)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error serializing analysis: %v\n", err)
		os.Exit(3)
	}

	if err := os.MkdirAll(filepath.Dir(*outputPath), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(4)
	}

	if err := os.WriteFile(*outputPath, append(data, '\n'), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
		os.Exit(5)
	}

	fmt.Printf("Analysis complete. Decision: %s. Output written to %s\n",
		result.PreliminaryDecision, *outputPath)
}
