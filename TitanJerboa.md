# TitanJerboa

**TitanJerboa** is an AST optimizer for the Tokype language. The name: **Titan** (powerful) + **Jerboa** (a fast little rodent). It runs before execution: it rewrites the AST so the program computes faster.

## What it does

**Constant folding** — evaluates expressions at optimization time:
```tokype
x = 7 + 6 + 5   // x = 18
```

**Formulas instead of loops** — replaces loops with mathematical formulas (O(n) O(1)):
```tokype
for i = 1, 1000000: sum = sum + i end   // sum = 500000500000
```
Supports: `sum i`, `sum i*i`, `sum i*i*i`, `prod i`, `sum k`, `sum k*i`, steps (`for i = 1, n, 2`), reverse steps.

**Dynamic formulas** — if the bounds are unknown (`input()`), the loop is replaced with a formula in the AST:
```tokype
n = input()
for i = 1, n: sum = sum + i end   // sum = n * (n + 1) / 2
```

**Recursion memoization** — evaluates recursive functions with a constant argument iteratively:
```tokype
fib(92)   // 7540113804746346429, 0 sec (instead of forever)
```

**Pure function evaluation** — functions without side effects and with constant arguments are computed in advance:
```tokype
x = square(5)   // x = 25
```

**Dead code elimination**:
```tokype
x = x + 0   // removed
x = x       // removed
x = x * 0   // x = 0
```

**Algebraic rules**:
```tokype
x - x        // 0
x / x        // 1
(x * y) / y  // x
x * 0        // 0
x * 1        // x
x + 0        // x
```

**Function inlining** — embeds the function body at the call site with argument substitution.

**Loop bound substitution** — replaces known variables in `for` bounds with constants.

## Folding into `repeat`

TitanJerboa detects loops that just fill a list with a constant and replaces them with a single `repeat` call:

```tokype
arr = []
for i = 1, 1000000:
    push(arr, 42)
end
// → arr = repeat(42, 1000000)
```

Complexity drops from O(n) to O(1).

**Triggers when:** the loop body is exactly one `push`, the first argument is the same list, the second argument is a constant, and the loop bounds are constants.

**Does not trigger** when the value depends on the counter (`push(arr, i * 2)`), when the body has more than one statement, or when the bounds are variables.

**What `repeat` returns:**
- small `count` (≤ 100000) → real list (`ListObject`) — supports `push`, `pop`, indexing
- large `count` → lazy `RepeatObject` — saves memory, indexing still works

## Pass order

1. `resolveIncludes` — expand `include`
2. `collectFunctions` + `collectTypes` — gather functions and types
3. `inlineFunctionCalls` — inlining
4. `constantFold` — constant folding
5. `substituteForBounds` — bound substitution
6. `constantFold` — second fold
7. `foldDynamicLoops` — dynamic formulas
8. `foldFunctionCalls` — pure function and recursion evaluation
9. `optimizeControlFlow` / `removeDeadCode` — cleanup
10. `removeEmptyFunctions` — remove empty functions

## Results

| Test | Tokype | C++ (-O3) | Python |
|------|--------|---|--------|
| `sum i (example/benchmark.tp)` | **0 sec** | **0 sec** | 0.5 sec |
| `fib(92) (example/fib.tp)` | **0 sec** | forever | forever |
| `prod i (example/prod.tp)` (20) | **0 sec** | **0 sec** | **0 sec** |

For math loops Tokype is faster than compiled languages — because it replaces the loop with a formula instead of running it.

## Limitations

- String concatenation in a loop folds as a number (`"base" + i` → `"base55"`)
- `int64` overflow
- Not all loop patterns are supported

## Documentation

- [**README.md**](README.md) — overview and quick start
- [**SPEC.md**](SPEC.md) — — full syntax reference with examples