# Pilot Protocol: Task Contract Quality Trials

This document specifies the operational rules for the contract evaluation trials.

## Purpose

The pilot measures whether structured task contracts reduce implementation mistakes,
scope creep, and human intervention compared to natural language task descriptions.
The trials evaluate six paired tasks across two experimental arms:
- Baseline arm: plain Markdown task specification (baseline.md).
- Contract arm: structured JSON work package (work-package.json).

## Core Guardrails and Constraints

Every operator and automated agent must follow these strict operational rules:
1. No synthetic receipts: all trial evidence must come from genuine model runs.
   Simulated transcripts, fabricated token counts, and fake exit codes are forbidden.
2. Complete workspace isolation: each trial must run in a dedicated disposable
   directory. No artifacts, code, or context from prior trials may leak into a run.
3. Blind execution: the model under test must only receive the task specification
   (baseline.md or work-package.json) and interface signatures. The test oracle and
   reference solution must never be shown to the model during implementation.
4. Zero unrecorded intervention: if an operator provides any guidance, prompt
   correction, or clarification, record an intervention with its exact cause.
5. Resource bounds: each trial has a strict limit of 30 minutes wall time and 32,000
   output tokens. Exceeding limits marks the trial as failed with limits exceeded.
6. Fail-closed verification: acceptance is decided only by the independent oracle
   test suite running against the generated code.

## Trial Inventory and Execution Order

Trials execute in seeded randomized order across the six task fixtures:
1. Task 03-file-preservation (atomic file write with fallback)
2. Task 01-cli-validation (flag validation and numeric bounds)
3. Task 05-conditional-dependency (task DAG resolution and cycle rejection)
4. Task 02-parser-boundary (token delimiter parsing and quote escape)
5. Task 06-multistep-resume (sequential workflow with checkpoint resume)
6. Task 04-cancellation (pipeline context cancellation and resource cleanup)

For each task, execute the baseline arm first, then the contract arm, in separate
fresh sessions or clean disposable directories.

## Per-Trial Execution Runbook

For each task fixture (for example, 01-cli-validation) and arm (baseline or contract):

Step 1: Set up disposable workspace
Create a clean directory outside the main code tree:
mkdir -p scratch/trials/<fixture-id>-<arm>
Copy only the package name declaration into the scratch directory.

Step 2: Dispatch task prompt to model
Feed the task specification to the model:
- For baseline arm: provide the text of baseline.md.
- For contract arm: provide the JSON content of work-package.json.

Step 3: Capture generated code and metadata
Save the generated implementation into the designated Go source file.
Record wall clock duration, token usage, and any interventions required.

Step 4: Execute independent oracle verification
Copy the independent oracle_test.go from the fixture into the scratch directory.
Run the oracle suite:
go test -v ./...
Capture full stdout, stderr, and exit status.

Step 5: Record trial result
Create the run record directory:
mkdir -p goals/goal-contract-pilot/runs/<fixture-id>-<arm>
Save the execution result as result.json using the schema below.

## Result Schema

Save each trial outcome to runs/<fixture-id>-<arm>/result.json:
{
  "schema_version": "sila/trial-result/v1",
  "fixture_id": "<fixture-id>",
  "arm": "baseline|contract",
  "model_version": "<model-identifier>",
  "harness_version": "antigravity-1.0",
  "start_commit": "652add091e1bb066398528277cda0a8308eb4b91",
  "final_commit": "<commit-hash-or-uncommitted>",
  "packet_digest": "<sha256-of-baseline-or-contract-packet>",
  "oracle_digest": "<sha256-of-oracle-test>",
  "first_pass": true,
  "accepted": true,
  "interventions": 0,
  "scope_violations": 0,
  "escaped_defects": 0,
  "wall_seconds": 45,
  "input_tokens": 1200,
  "output_tokens": 850,
  "cost": 0.0,
  "cost_unit": "USD",
  "evidence_paths": [
    "goals/goal-contract-pilot/runs/<fixture-id>-<arm>/oracle.log"
  ],
  "outcome": "accepted|rejected|timeout|blocked"
}

## Run Index Maintenance

Maintain all trial references in goals/goal-contract-pilot/runs/index.json:
{
  "schema_version": "sila/pilot-run-index/v1",
  "total_trials": 12,
  "completed_trials": 0,
  "runs": []
}

## Automated Analysis

Once all twelve runs complete, compute statistics using the analysis tool:
go run ./goals/goal-contract-pilot/analysis.go \
  --manifest goals/goal-contract-pilot/pilot-manifest.json \
  --runs goals/goal-contract-pilot/runs/index.json \
  --output goals/goal-contract-pilot/analysis.json
