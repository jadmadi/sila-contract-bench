# Pilot Fixtures

This directory contains six paired programming tasks for the contract pilot.

## Overview

Each fixture folder contains four core files:
- baseline.md: Natural language task specification.
- work-package.json: Structured machine-readable task contract.
- Implementation source file (.go): Target code under test.
- oracle_test.go: Test suite verifying acceptance and negative controls.

## Fixture List

- 01-cli-validation: CLI option parser with timeout and concurrency flags.
  Negative controls reject non-positive timeout values and out-of-range bounds.
- 02-parser-boundary: Delimited token extractor with quote-aware escaping.
  Negative controls reject unclosed quote sequences and trailing escape marks.
- 03-file-preservation: Atomic file update with fallback and permission preservation.
  Negative controls reject empty payload writes and preserve existing files on failure.
- 04-cancellation: Multi-stage pipeline with context propagation and early cancellation.
  Negative controls verify that cancelled context halts work without leaked stages.
- 05-conditional-dependency: Task DAG execution planner with topological resolution.
  Negative controls reject cyclical dependency graphs and self-referential tasks.
- 06-multistep-resume: Multi-step task checkpoint and resume workflow.
  Negative controls halt work immediately on error and skip previously completed steps.

## Verification

Run all oracles with this command from the repository root:
go test -v ./goals/goal-contract-pilot/fixtures/...
