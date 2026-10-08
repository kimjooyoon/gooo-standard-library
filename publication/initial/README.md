# Initial reusable-library observation

Date: 2026-10-08. Local platform: macOS arm64. Compiler: published Gooo
0.6.10-dev, clean source `eb0dc4704705147f9ba944db3df2ba5e3225cbff`, Go 1.27.1,
decision runtime v0.2.26-experimental. `compiler-build.json` retains the identity.

## What was frozen and run

- Library, consumer and plan: commit `2598be3`.
- Explicit library and consumer evaluation fixtures: commit `dbeaab4`.
- First paired model execution: runner commit `d129e62`.
- Final observer, including input and model-call checks: commit `75833a0`.

Library and selection/evaluation sources were unchanged across these runs.
`first-fixed-summary.json`, `first-paired-summary.json` and the final `summary.json`
retain each stage. Canonical comparison of the two paired summaries found the
same generated hashes, actual-case counts, field counts and saved-replay outcomes.
Timing changes between observations are retained.

The optional graph QAT bundle is from the ecosystem workbench at source
`19485bb64278ab8ac219e042f92af677ff3f5051`:

- metadata SHA-256: `3c68205a660695103712e2115ce90accd298b406dced8ad5ad5c9b84bab0a202`
- weights SHA-256: `76f68845a03ed8c8bc261a57c919e96dcdd352d3c35bcdaeb3d9b51a88975c6f`

It was trained previously on filename, division and retry programs. No model
training was performed for this library or consumer. The model is not included
in the evidence archive; its metadata and weight digests are in each model receipt.

## Counts and results

All 17 library entry points matched their 79 explicit function-input rows and
their saved replays. These are rows across named functions; a value used for two
functions is not claimed to be two globally distinct inputs.

One importing preview consumer has five source selection examples, three binary
field choices, and ten distinct evaluation input tuples disjoint from those five.
Both fixed ordering and the model required all eight candidate attempts to satisfy
the selection examples. Each complete program matched 10/10 evaluation cases and
30/30 fields. At smaller budgets the partial results differ; see the root README
table and all actual values in the archive. No incomplete run was discarded.

Every saved replay preserved the generated program and the actual outputs.
Each model construction used one prediction; native execution and saved replay
used zero new model calls. The full-budget programs were also replayed with three
input-only requests, retaining `OBSERVED`, a 0/0 correctness score and actual values.
Those requests carry no supplied expected answers.

This is a small library plus one consumer with fixed wording and choice layout.
The observations support this import-and-call path and finite behavior. They do
not establish broad program transfer or a probability of arbitrary intent completion.
Model timing and tensor storage are recorded per run; whole-process memory and
host CPU increase were not measured. No general speed advantage was demonstrated.

## Observer corrections

Two early observer runs stopped on metadata assumptions: package execution does
not emit the standalone body's `generated_now` flag, and fixed library bodies
omit chooser metadata when they contain no assembly. The observer now uses the
explicit package replay record and checks model metadata only where construction
requires it. The two original responses are retained under `failures/`; unit tests
cover missing counts, wrong identities, changed inputs/expectations and absent
required model observations. These corrections changed the Go observer, not Gooo
function behavior, source choices or model weights.

`evidence.tar.gz` contains the first paired and final raw receipts, source snapshots,
input files and generated programs, without compiler binaries. Extract into a new
directory. `SHA256SUMS` covers the retained publication files. See the root README
for the verification command; each execution uses a new output directory.
