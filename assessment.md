# Assessment: Task Contract Quality Trials

## Executive summary
We completed twelve controlled implementation trials across six paired task fixtures using
Antigravity and the Gemini 2.5 Flash model. The trial evaluated whether structured JSON
contracts (work-package.json) reduce mistakes, token consumption, and human intervention
compared to natural language specifications (baseline.md).

The evaluation yielded a stop decision against the frozen thresholds:
- Total trials executed: 12 (6 baseline, 6 contract).
- Baseline acceptance: 5 of 6 tasks accepted.
- Contract acceptance: 3 of 6 tasks accepted.
- Total interventions: 0 across all runs in both arms.
- Contract escaped defects: 3 (02-parser-boundary, 04-cancellation, 06-multistep-resume).
- Contract scope violations: 0 across all runs.
- Median paired wall-clock ratio: 1.61 (contract runs took longer in median).
- Median paired token ratio: 0.83 (contract runs consumed 17% fewer tokens in median).

The contract arm demonstrated worse acceptance than baseline (3 versus 5) with zero
intervention benefit, triggering an explicit stop decision under the frozen protocol rules.

## Trial inventory and run outcomes
All twelve runs executed in isolated scratch worktrees without context leakage:

| Fixture ID               | Arm      | Outcome  | First Pass | Wall (s) | Tokens | Defect Delta |
| ------------------------ | -------- | -------- | ---------- | -------- | ------ | ------------ |
| 01-cli-validation        | baseline | rejected | false      | 199.0    | 44127  | 0            |
| 01-cli-validation        | contract | accepted | true       | 59.0     | 14488  | 0            |
| 02-parser-boundary       | baseline | accepted | true       | 114.0    | 16085  | 0            |
| 02-parser-boundary       | contract | rejected | false      | 43.0     | 15300  | +1           |
| 03-file-preservation     | baseline | accepted | true       | 43.0     | 25556  | 0            |
| 03-file-preservation     | contract | accepted | true       | 55.0     | 16492  | 0            |
| 04-cancellation          | baseline | accepted | true       | 26.0     | 17981  | 0            |
| 04-cancellation          | contract | rejected | false      | 68.0     | 17000  | +1           |
| 05-conditional-dependency| baseline | accepted | true       | 45.0     | 24748  | 0            |
| 05-conditional-dependency| contract | accepted | true       | 87.0     | 17700  | 0            |
| 06-multistep-resume      | baseline | accepted | true       | 8.0      | 12643  | 0            |
| 06-multistep-resume      | contract | rejected | false      | 193.0    | 20500  | +1           |

## Analysis of trial failures
Four runs failed acceptance across the twelve trials:

1. 01-cli-validation baseline:
The natural language description left flag defaults versus requirements ambiguous. The model
exhausted its 32,000 output token limit reasoning about edge cases, requiring a retry. Because
it exceeded output resource limits, the trial was rejected under AC-006. In contrast, the
contract arm resolved the ambiguity cleanly in 59 seconds and 14,488 tokens.

2. 02-parser-boundary contract:
The contract omitted UNEXPECTED_CLOSING_DELIMITER from its constraint array. Following the
contract literally, the model emitted MISMATCHED_DELIMITER when encountering an unexpected
closing delimiter on an empty stack. The independent oracle expected the specific code,
failing the negative control test.

3. 04-cancellation contract:
The contract specification provided function signatures but omitted concrete struct
definitions for PipelineStep and PipelineResult. The model synthesized PipelineStep as an
interface rather than a struct. The oracle test suite attempted struct initialization,
causing a compilation failure.

4. 06-multistep-resume contract:
The contract omitted the Title field in StepDefinition and omitted the type of LastError in
Checkpoint. The model synthesized LastError as error instead of string, leading to compilation
failures against the oracle test.

## Threshold evaluation
The results evaluate against the frozen decision thresholds as follows:
- Minimum contract tasks accepted (>= 5): Failed (actual: 3).
- Maximum contract scope violations (<= 0): Passed (actual: 0).
- Maximum contract escaped defects (<= 0): Failed (actual: 3).
- First pass acceptance not worse than baseline: Failed (contract 3, baseline 5).
- Minimum intervention reduction (>= 25%): Failed (baseline interventions were 0).
- Maximum paired wall ratio (<= 1.25): Failed (actual median: 1.61).
- Maximum paired token ratio (<= 1.25): Passed (actual median: 0.83).

## Decision and recommendations
Decision: stop.

The results show that while well-formed contracts reduce token usage and eliminate ambiguity
(as seen in 01-cli-validation), incomplete contract generation is catastrophic. Natural
language prompts provide intuitive contextual type information that models handle robustly.
When structured contracts omit fields or concrete types, models fail to infer them, producing
build and verification breaks.

We recommend halting the automated rollout of goal contracts until contract generation
tooling guarantees hermetic type definitions and exhaustive constraint extraction.

## Limitations
- Small sample size of six paired programming tasks.
- Evaluates isolated unit tasks; does not evaluate large multi-file migrations.
- Results reflect the specific Gemini Flash model and Antigravity harness configuration.
