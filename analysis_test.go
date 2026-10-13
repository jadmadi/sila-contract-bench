package main

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"
)

func createTestManifest(t *testing.T, dir string) string {
	t.Helper()
	manifest := PilotManifest{
		SchemaVersion:   "sila/pilot-manifest/v1",
		BaselineCommit:  "test-commit",
		CandidateCommit: "test-commit",
		Model:           "test-model",
		Harness:         "test-harness",
		Limits: LimitsConfig{
			WallClockTimeoutMinutes: 30,
			MaxOutputTokens:         32000,
		},
		RandomizedOrder: []string{"fix-1", "fix-2"},
		DecisionThresholds: DecisionThresholds{
			MinContractAccepted:         1,
			MaxContractScopeViolations:  0,
			MaxContractEscapedDefects:   0,
			MinInterventionReductionPct: 25.0,
			MaxPairedWallRatio:          1.25,
			MaxPairedTokenRatio:         1.25,
		},
		Fixtures: []FixtureManifest{
			{
				ID:                   "fix-1",
				SourceCommit:         "test-commit",
				TaskKind:             "test-task",
				OraclePath:           "oracle.go",
				OracleDigest:         "oracle-digest-1",
				BaselinePacketPath:   "baseline.md",
				BaselinePacketDigest: "base-digest-1",
				ContractPacketPath:   "work-package.json",
				ContractPacketDigest: "contract-digest-1",
			},
			{
				ID:                   "fix-2",
				SourceCommit:         "test-commit",
				TaskKind:             "test-task-2",
				OraclePath:           "oracle.go",
				OracleDigest:         "oracle-digest-2",
				BaselinePacketPath:   "baseline.md",
				BaselinePacketDigest: "base-digest-2",
				ContractPacketPath:   "work-package.json",
				ContractPacketDigest: "contract-digest-2",
			},
		},
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatalf("marshaling test manifest: %v", err)
	}

	path := filepath.Join(dir, "manifest.json")
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("writing test manifest: %v", err)
	}
	return path
}

func TestPilotAnalysis(t *testing.T) {
	tempDir := t.TempDir()
	manifestPath := createTestManifest(t, tempDir)

	t.Run("SyntheticUnitFixture_ProceedCase", func(t *testing.T) {
		runs := RunIndex{
			SchemaVersion:   "sila/pilot-run-index/v1",
			TotalTrials:     4,
			CompletedTrials: 4,
			Runs: []TrialResult{
				{
					FixtureID:     "fix-1",
					Arm:           "baseline",
					OracleDigest:  "oracle-digest-1",
					PacketDigest:  "base-digest-1",
					FirstPass:     false,
					Accepted:      true,
					Interventions: 4,
					WallSeconds:   100.0,
					InputTokens:   500,
					OutputTokens:  500,
				},
				{
					FixtureID:     "fix-1",
					Arm:           "contract",
					OracleDigest:  "oracle-digest-1",
					PacketDigest:  "contract-digest-1",
					FirstPass:     true,
					Accepted:      true,
					Interventions: 1,
					WallSeconds:   90.0,
					InputTokens:   550,
					OutputTokens:  450,
				},
				{
					FixtureID:     "fix-2",
					Arm:           "baseline",
					OracleDigest:  "oracle-digest-2",
					PacketDigest:  "base-digest-2",
					FirstPass:     true,
					Accepted:      true,
					Interventions: 2,
					WallSeconds:   80.0,
					InputTokens:   400,
					OutputTokens:  400,
				},
				{
					FixtureID:     "fix-2",
					Arm:           "contract",
					OracleDigest:  "oracle-digest-2",
					PacketDigest:  "contract-digest-2",
					FirstPass:     true,
					Accepted:      true,
					Interventions: 0,
					WallSeconds:   70.0,
					InputTokens:   420,
					OutputTokens:  380,
				},
			},
		}

		runsPath := filepath.Join(tempDir, "runs-valid.json")
		runsData, _ := json.Marshal(runs)
		_ = os.WriteFile(runsPath, runsData, 0644)

		res, err := RunAnalysis(manifestPath, runsPath)
		if err != nil {
			t.Fatalf("unexpected error in analysis: %v", err)
		}

		if res.PreliminaryDecision != "proceed" {
			t.Fatalf("expected decision proceed, got: %s", res.PreliminaryDecision)
		}
		if res.ContractSummary.AcceptedRuns != 2 {
			t.Fatalf("expected 2 accepted contract runs, got: %d", res.ContractSummary.AcceptedRuns)
		}
		if res.InterventionDelta.BaselineTotal != 6 {
			t.Fatalf("expected 6 baseline interventions, got: %d", res.InterventionDelta.BaselineTotal)
		}
		if res.InterventionDelta.ContractTotal != 1 {
			t.Fatalf("expected 1 contract intervention, got: %d", res.InterventionDelta.ContractTotal)
		}
		if res.InterventionDelta.AbsoluteDelta != -5 {
			t.Fatalf("expected -5 delta, got: %d", res.InterventionDelta.AbsoluteDelta)
		}
		expectedPct := (5.0 / 6.0) * 100.0
		if res.InterventionDelta.ReductionPct == nil || math.Abs(*res.InterventionDelta.ReductionPct-expectedPct) > 0.0001 {
			t.Fatalf("expected reduction pct %f, got: %v", expectedPct, *res.InterventionDelta.ReductionPct)
		}
	})

	t.Run("SyntheticUnitFixture_ZeroBaselineInterventions", func(t *testing.T) {
		runs := RunIndex{
			SchemaVersion: "sila/pilot-run-index/v1",
			Runs: []TrialResult{
				{
					FixtureID:     "fix-1",
					Arm:           "baseline",
					OracleDigest:  "oracle-digest-1",
					PacketDigest:  "base-digest-1",
					FirstPass:     false,
					Accepted:      true,
					Interventions: 0,
					WallSeconds:   50.0,
					InputTokens:   200,
					OutputTokens:  200,
				},
				{
					FixtureID:     "fix-1",
					Arm:           "contract",
					OracleDigest:  "oracle-digest-1",
					PacketDigest:  "contract-digest-1",
					FirstPass:     true,
					Accepted:      true,
					Interventions: 0,
					WallSeconds:   50.0,
					InputTokens:   200,
					OutputTokens:  200,
				},
				{
					FixtureID:     "fix-2",
					Arm:           "baseline",
					OracleDigest:  "oracle-digest-2",
					PacketDigest:  "base-digest-2",
					FirstPass:     true,
					Accepted:      true,
					Interventions: 0,
					WallSeconds:   50.0,
					InputTokens:   200,
					OutputTokens:  200,
				},
				{
					FixtureID:     "fix-2",
					Arm:           "contract",
					OracleDigest:  "oracle-digest-2",
					PacketDigest:  "contract-digest-2",
					FirstPass:     true,
					Accepted:      true,
					Interventions: 0,
					WallSeconds:   50.0,
					InputTokens:   200,
					OutputTokens:  200,
				},
			},
		}

		runsPath := filepath.Join(tempDir, "runs-zero-interventions.json")
		runsData, _ := json.Marshal(runs)
		_ = os.WriteFile(runsPath, runsData, 0644)

		res, err := RunAnalysis(manifestPath, runsPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.InterventionDelta.ReductionPct != nil {
			t.Fatalf("expected nil reduction pct for zero baseline, got: %v", res.InterventionDelta.ReductionPct)
		}
		if res.InterventionDelta.Reason != "zero_denominator" {
			t.Fatalf("expected reason zero_denominator, got: %s", res.InterventionDelta.Reason)
		}
	})

	t.Run("MissingPair_BlocksProceed", func(t *testing.T) {
		// Only 3 runs, missing contract for fix-2
		runs := RunIndex{
			SchemaVersion: "sila/pilot-run-index/v1",
			Runs: []TrialResult{
				{
					FixtureID:    "fix-1",
					Arm:          "baseline",
					OracleDigest: "oracle-digest-1",
					PacketDigest: "base-digest-1",
				},
				{
					FixtureID:    "fix-1",
					Arm:          "contract",
					OracleDigest: "oracle-digest-1",
					PacketDigest: "contract-digest-1",
				},
				{
					FixtureID:    "fix-2",
					Arm:          "baseline",
					OracleDigest: "oracle-digest-2",
					PacketDigest: "base-digest-2",
				},
			},
		}

		runsPath := filepath.Join(tempDir, "runs-missing.json")
		runsData, _ := json.Marshal(runs)
		_ = os.WriteFile(runsPath, runsData, 0644)

		_, err := RunAnalysis(manifestPath, runsPath)
		if err == nil {
			t.Fatalf("expected error due to missing pair, got nil")
		}
	})

	t.Run("OracleDigestMismatch_BlocksProceed", func(t *testing.T) {
		runs := RunIndex{
			SchemaVersion: "sila/pilot-run-index/v1",
			Runs: []TrialResult{
				{
					FixtureID:    "fix-1",
					Arm:          "baseline",
					OracleDigest: "wrong-oracle-digest",
					PacketDigest: "base-digest-1",
				},
			},
		}

		runsPath := filepath.Join(tempDir, "runs-mismatch.json")
		runsData, _ := json.Marshal(runs)
		_ = os.WriteFile(runsPath, runsData, 0644)

		_, err := RunAnalysis(manifestPath, runsPath)
		if err == nil {
			t.Fatalf("expected error due to oracle digest mismatch, got nil")
		}
	})
}
