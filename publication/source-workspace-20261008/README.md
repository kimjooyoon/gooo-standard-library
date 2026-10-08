# Adopting source-derived package metadata

The manifest change and pre-run plan were frozen at
`dacd7d284ea6742b20441b60420799d23725b8c3`. Gooo functions, preview choices,
input fixtures and model weights are unchanged from the 0.1.0 library.
The compiler reads package names and imports from those source files.

## Local candidate regression

The full paired observer used clean Gooo 0.6.11 candidate
`60b1d7cb274f44531aa2501b00a29aed94b3958f`, Go 1.27.1 and the unchanged
workbench graph QAT model identified in PLAN.md.

All 17 named functions matched their 79 function-input cases and saved replay.
The single importing consumer was observed at budgets 1, 2, 4 and 8 under fixed
and model ordering. This makes 25 function/mode/budget runs, not 25 programs.

The existing exact-number comparison read all 25 summaries and 52 raw records
from the original retained observations and this new execution. Delivered inputs,
actual and expected values, finite counts, generated-code hashes and replay
flags matched. Timings and platform-dependent executable hashes were excluded.

| Budget | Fixed: input / field matches | Model: input / field matches |
| --- | --- | --- |
| 1 | 1/10 · 14/30 | 1/10 · 14/30 |
| 2 | 1/10 · 20/30 | 3/10 · 17/30 |
| 4 | 3/10 · 23/30 | 4/10 · 24/30 |
| 8 | 10/10 · 30/30 | 10/10 · 30/30 |

Both orderings still require eight candidates to satisfy the selection examples.
Earlier partial results are retained. This compiler-adoption regression provides
no new training or unseen-program transfer result.

The preview command also constructed fixed and model-guided programs for title
`한글 예제`, budget 20. Both returned `{available:true, bytes:13, title:"한글 예제"}`.
Replaying the model-built program with an empty title and budget 1 returned
`{available:false, bytes:1, title:"untitled"}` and made zero new predictions.

## Retained files

`summary.json`, `compiler-build.json` and `comparison.txt` retain the local
identities and result. `local-observations.tar.gz` contains the full paired
observer outputs and three direct CLI uses, with no compiler or model binaries.
The original baseline is already in `publication/initial/evidence.tar.gz`.
Use the unchanged `publication/ci-20261008/compare.go` to compare the extracted
baseline's `final` directory with the new archive's `paired` directory.

The candidate observation precedes public 0.6.11 adoption. Its public archive,
CI pin and subsequent Linux reproduction are recorded separately when available.
