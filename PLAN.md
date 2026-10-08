# Initial library and consumer evaluation

Freeze this plan and the source/fixtures before invoking the model.

The library has 17 Gooo activities in three packages: eight integer helpers,
three Boolean helpers and six Text helpers. Calls within the library and calls
from an importing consumer use the compiler's normal package path. Text uses
the current 1,024-byte UTF-8 boundary; lengths are bytes. Reversed integer bounds
are normalized and abs(int64 minimum) saturates at int64 maximum.

Supply explicit expected values for every library activity, including empty
Text, Korean and emoji, int64 boundaries, Boolean combinations and reversed bounds.
Execute and replay each entry through the released 0.6.10 compiler. Count named
functions, input tuples, output expectations and repeat executions separately.

The preview consumer imports all three packages. It assembles three result fields
from two choices each. Five source selection cases and separate execution inputs
exercise title fallback, availability and a byte budget clamped to 0..80. A byte
budget is a count, not a Unicode slicing operation. Compare fixed ordering with
the unchanged public graph QAT model, at budgets 1, 2, 4 and 8. Preserve incomplete
outputs, finite field counts, chosen candidates and original inference timings.
Saved replay must preserve generated code and actual outputs with zero new model
calls. Input-only use must keep its correctness score unobserved.

This is one importing consumer and a small public library. It is not a new
training study or a calibrated probability of arbitrary intent completion.
Model files stay external; no model download is necessary for fixed ordering.
Any unsupported compiler behavior discovered here is recorded and repaired at
its source rather than replaced by a Go implementation of the library function.
