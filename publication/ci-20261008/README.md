# Linux reproduction and direct use

Date: 2026-10-08. The public repository's first
[CI run 37766633713](https://github.com/kimjooyoon/gooo-standard-library/actions/runs/37766633713)
passed at source `d8fc7613e5f2352d0c8b14c00f59ef47f307fa1d`.
It downloaded the published Linux amd64 Gooo 0.6.10 compiler and checked its
archive hash, embedded version, clean source identity and Go 1.27.1 toolchain.
The same frozen Workbench model was used without additional training.

`linux-ci.zip` is the original `library-native-observations` artifact, ID
11545115086, SHA-256
`b94cc84d363157fbd0391a1d007119bf8ace9b5210260ffc7ae202601e8dd8b7`.
`run.json` and `artifact.json` retain the source-bound GitHub observations.

## Compared with the local observation

The 25 runs comprise 17 named library functions and one importing consumer at
four budgets under two candidate orderings. They are not 25 independent programs.
All library expectations matched (79 function-input rows). The consumer's ten
distinct evaluation tuples are disjoint from its five source selection examples.
Both full-budget constructions matched 10/10 cases and 30/30 fields. The partial
case/field counts in the root README were reproduced without changing Gooo source,
fixtures or weights. Neither ordering reduced the eight attempts needed for full
selection completion in this consumer.

`compare.go` compares the summaries and the 52 raw execution/replay/input-only
records with exact JSON numbers. It checks delivered inputs, actual and expected
values, finite counts, generated-code hashes and saved-replay flags. It deliberately
does not compare timings or platform-dependent executable hashes. Both the
first/final local comparison and final-local/Linux comparison passed; their
complete results are retained here.

To repeat from the repository root after extracting the initial local archive
and the Linux ZIP into two new directories:

```sh
go run ./publication/ci-20261008/compare.go /path/to/local/final /path/to/linux/ci
```

## Direct user inputs

The README's direct compiler commands were run from this repository's root using
the installed macOS arm64 public compiler. All three actual records matched their
saved replay; the result remained `OBSERVED` with no correctness expectations.

The subsequently added `cmd/preview` forwards one title and integer budget to the
same Gooo workspace and prints its returned record. It does not implement the
library's conditions or calculations in Go. Local fixed and model construction
on `("한글 예제", 20)` returned `{available:true, title:"한글 예제", bytes:13}`.
Reusing the model-built program on `("", 1)` returned
`{available:false, title:"untitled", bytes:1}` with zero new inference.
`local-usage.json` and `local-usage.tar.gz` retain those outputs, inputs and full
receipts. These usage checks are separate from the ten-case frozen evaluation.

This original CI run predates the short preview command. Its workflow now also
runs the fixed/model command and a different-input saved replay; the later CI
run is linked from the repository's release notes once completed. Observer unit
and race checks, static checks and the three actual local command invocations
passed before publication.

No whole-process CPU or memory comparison was performed. Per-call model timing
and native phase observations remain in the raw records. `SHA256SUMS` covers the
retained files, including the original ZIP and the local usage archive.
