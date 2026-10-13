# Sila Contract Bench

An empirical benchmark evaluating structured JSON task contracts versus plain Markdown specifications for autonomous AI coding agents.

## Overview

Many AI agent frameworks assume structured JSON schemas reduce ambiguity, save tokens, and prevent errors compared to natural language Markdown specifications.

We evaluated this hypothesis across 12 controlled implementation trials using Antigravity and the Gemini 2.5 Flash model across six paired programming tasks.

The results yielded an explicit stop decision against frozen engineering thresholds:
- Token efficiency: JSON contracts reduced token consumption by 17% in median.
- Acceptance rate: Task acceptance fell from 5 of 6 (83.3%) on Markdown to 3 of 6 (50.0%) on JSON.
- Escaped defects: 3 in the JSON contract arm, 0 in the Markdown baseline arm.
- Wall-clock latency: JSON contract runs took 61% longer in median due to compilation retries.
- Decision: STOP. Automated contract rollout was halted until extraction tooling guarantees exhaustive type definitions.

## Results Matrix

| Fixture ID | Arm | Outcome | First Pass | Wall (s) | Tokens | Defect Delta |
| --- | --- | --- | --- | --- | --- | --- |
| 01-cli-validation | baseline | rejected | false | 199.0 | 44127 | 0 |
| 01-cli-validation | contract | accepted | true | 59.0 | 14488 | 0 |
| 02-parser-boundary | baseline | accepted | true | 114.0 | 16085 | 0 |
| 02-parser-boundary | contract | rejected | false | 43.0 | 15300 | +1 |
| 03-file-preservation | baseline | accepted | true | 43.0 | 25556 | 0 |
| 03-file-preservation | contract | accepted | true | 55.0 | 16492 | 0 |
| 04-cancellation | baseline | accepted | true | 26.0 | 17981 | 0 |
| 04-cancellation | contract | rejected | false | 68.0 | 17000 | +1 |
| 05-conditional-dependency | baseline | accepted | true | 45.0 | 24748 | 0 |
| 05-conditional-dependency | contract | accepted | true | 87.0 | 17700 | 0 |
| 06-multistep-resume | baseline | accepted | true | 8.0 | 12643 | 0 |
| 06-multistep-resume | contract | rejected | false | 193.0 | 20500 | +1 |

## Why JSON Failed: The Redundancy Gap

The trials exposed a fundamental contrast between natural language and strict schemas:

1. Natural language has defensive redundancy:
Prose, examples, and signatures reinforce each other. If a requirements table misses a field, a bullet point or code snippet often provides the missing context.

2. Strict schemas are completely brittle:
When an automated schema extractor extracts only direct function signatures and omits auxiliary structs or error constants, the prompt has zero surrounding context. The model follows the schema literally, is forced to guess without guidance, and synthesizes broken types.

### Specific Failure Breakdowns

- Task 02-parser-boundary (Contract Arm Failed):
  The JSON contract omitted UNEXPECTED_CLOSING_DELIMITER from its constraint array. Following the schema literally, the model emitted MISMATCHED_DELIMITER on empty stacks. The independent test oracle expected the specific code, failing verification.

- Task 04-cancellation (Contract Arm Failed):
  The JSON contract extracted the function signature but omitted concrete struct definitions for PipelineStep and PipelineResult. The model synthesized PipelineStep as a Go interface instead of a struct. The oracle test initialized the type as a struct, causing a build break.

- Task 06-multistep-resume (Contract Arm Failed):
  The JSON schema omitted the Title field in StepDefinition and omitted the type of LastError in Checkpoint. The model synthesized LastError as the Go error interface instead of string, failing compilation against the oracle.

- Task 01-cli-validation (Baseline Arm Failed):
  Conversely, when natural language is genuinely ambiguous about default flags versus mandatory flags, models can burn tokens in extensive reasoning loops. Here, the well-formed JSON contract won decisively (59 seconds vs 199 seconds).

## Repository Structure

```
├── fixtures/                      # Six paired programming task fixtures
│   ├── 01-cli-validation/         # CLI parser with numeric bounds
│   ├── 02-parser-boundary/        # Delimited token extractor with quote escaping
│   ├── 03-file-preservation/      # Atomic file updates with permission preservation
│   ├── 04-cancellation/           # Pipeline context cancellation and cleanup
│   ├── 05-conditional-dependency/ # Topological DAG task resolution and cycle rejection
│   └── 06-multistep-resume/       # Sequential workflow with checkpoint resume
├── runs/                          # Raw evidence from all 12 model trials
│   ├── index.json                 # Master manifest of runs and recorded metrics
│   └── <run-id>/                  # Individual patch.diff, oracle.log, and result.json
├── manifest.json                  # Cryptographic SHA-256 digests of all fixtures and oracles
├── protocol.md                    # Frozen experimental protocol and trial runbook
├── assessment.md                  # Comprehensive empirical evaluation report
├── analysis.go                    # Automated decision analyzer in pure Go
├── analysis_test.go               # Integrity and oracle verification tests
└── LICENSE                        # GNU Affero General Public License v3.0
```

## Running the Test Oracles

All fixtures use pure Go with zero third-party dependencies.

To run all six test oracles and negative control tests locally:

```bash
go test -v ./fixtures/...
```

To run the integrity tests on the analysis engine:

```bash
go test -v .
```

To recompute the full statistical decision report against frozen thresholds:

```bash
go run . --manifest manifest.json --runs runs/index.json --output analysis.json
```

## Benchmarking New Models and Schemas

We welcome contributions, alternative schemas, and benchmark runs from other models (Claude 3.7 Sonnet, OpenAI o3, DeepSeek, etc.).

1. Follow the runbook in protocol.md.
2. Maintain clean directory isolation: run each trial in an empty scratch directory without context leakage.
3. Present only baseline.md (Markdown arm) or work-package.json (Contract arm) to the model. Do not show the oracle tests.
4. Verify the candidate implementation against the corresponding oracle_test.go.
5. Record the resulting diff, token usage, wall-clock duration, and test output into runs/<id>/.
6. Submit a pull request with your results.

## License

GNU Affero General Public License v3.0 (AGPL-3.0). See LICENSE for details.
