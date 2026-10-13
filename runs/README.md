# Trial Execution Runs

This directory stores sanitized trial results, logs, and patches for the pilot.

## Directory Structure

Each trial stores its artifacts in a dedicated subdirectory:
- runs/<fixture-id>-baseline/
  - result.json: structured trial result complying with TrialResult v1.
  - patch.diff: code produced by the model during the trial.
  - oracle.log: complete output log of the oracle test run.
- runs/<fixture-id>-contract/
  - result.json: structured trial result complying with TrialResult v1.
  - patch.diff: code produced by the model during the trial.
  - oracle.log: complete output log of the oracle test run.

## Index File

The index file index.json catalogs all runs and links to each result file.
The analysis tool reads index.json to verify completeness and compute outcomes.
