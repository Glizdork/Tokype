# TitanJerboa

**TitanJerboa** (*Titan* — powerful, *Jerboa* — a fast little rodent) is the static AST optimizer for the Tokype programming language. It executes prior to runtime evaluation, rewriting and simplifying the Abstract Syntax Tree (AST) to eliminate redundant computations and reduce algorithmic complexity.

---

## Optimization Passes & Features

### 1. Loop Folding (Closed-Form Formula Replacement)
TitanJerboa analyzes `for` loops and replaces supported accumulation patterns with closed-form mathematical formulas, reducing time complexity from **$O(n)$ to $O(1)$**.

* **Static Loop Folding:** Evaluates constant-bound loops directly at optimization time.
  ```tokype
  for i = 1, 1000000: sum = sum + i end
  // Rewritten to: sum = 500000500000
  ```
* **Dynamic Loop Folding:** When bounds depend on runtime values (e.g., `input()`), TitanJerboa injects the closed-form formula directly into the AST.
  ```tokype
  n = input()
  for i = 1, n: sum = sum + i end
  // Rewritten to: sum = n * (n + 1) / 2
  ```
* **Supported Patterns:** `sum i`, `sum i*i`, `sum i*i*i`, `prod i`, `sum k`, `sum k*i`, custom step increments (`for i = 1, n, 2`), and reverse steps.

### 2. List Allocation Folding (`repeat`)
Loops that sequentially populate a list with a constant value are collapsed into a single `repeat` call ($O(n) \rightarrow O(1)$):

```tokype
arr = []
for i = 1, 1000000:
    push(arr, 42)
end
// Rewritten to: arr = repeat(42, 1000000)
```

* **Trigger Conditions:** The loop body contains strictly one `push` call targeting the same list, the pushed value is a constant, and loop bounds are static constants. (Does not trigger if the value depends on the loop counter, e.g., `push(arr, i * 2)`).
* **Adaptive Memory Representation:**
  * `count <= 100,000` → allocates a standard mutable `ListObject` (supports `push`, `pop`, indexing).
  * `count > 100,000` → returns a lazy `RepeatObject` to preserve memory while maintaining $O(1)$ indexing.

### 3. Function Optimizations
* **Recursion Memoization & Iterative Evaluation:** Recursive functions invoked with constant arguments are evaluated iteratively during the AST optimization phase:
  ```tokype
  fib(92)   // Evaluated to 7540113804746346429 in 0s
  ```
* **Pure Function Pre-Evaluation:** Side-effect-free functions called with constant arguments are resolved ahead of execution (`square(5)` $\rightarrow$ `25`).
* **Function Inlining:** Embeds small function bodies directly at call sites with argument substitution to eliminate call overhead.

### 4. Constant Folding, Algebraic Simplification & DCE
* **Constant Folding & Loop Bound Substitution:** Evaluates static expressions (`7 + 6 + 5` $\rightarrow$ `18`) and propagates known variables into `for` loop bounds.
* **Algebraic Identities:**

  | Original Expression | Simplified AST |
  | :--- | :--- |
  | `x - x` | `0` |
  | `x / x` | `1` |
  | `(x * y) / y` | `x` |
  | `x * 0` | `0` |
  | `x * 1` | `x` |
  | `x + 0` | `x` |
* **Dead Code Elimination (DCE):** Strips no-op assignments (`x = x`, `x = x + 0`), simplifies zero multiplications (`x = x * 0` $\rightarrow$ `x = 0`), and removes unreachable branches or empty functions.

---

## Optimization Pipeline (Pass Order)

TitanJerboa executes a deterministic 10-stage pass pipeline over the AST:

| Stage | Pass Name | Purpose |
| :--- | :--- | :--- |
| **1** | `resolveIncludes` | Expands `include` directives into a unified AST |
| **2** | `collectFunctions` + `collectTypes` | Indexes user-defined functions and type declarations |
| **3** | `inlineFunctionCalls` | Inlines eligible function calls at call sites |
| **4** | `constantFold` | First pass of constant expression evaluation |
| **5** | `substituteForBounds` | Propagates constant variables into `for` loop boundaries |
| **6** | `constantFold` | Second pass to fold expressions exposed after bound substitution |
| **7** | `foldDynamicLoops` | Replaces symbolic loops with closed-form AST formulas |
| **8** | `foldFunctionCalls` | Pre-evaluates pure functions and memoized recursive calls |
| **9** | `optimizeControlFlow` / `removeDeadCode` | Eliminates dead branches and redundant assignments |
| **10** | `removeEmptyFunctions` | Strips unused/empty function declarations |

---

## Benchmarks

Even as a pure tree-walking AST interpreter without compilation, Tokype outperforms AOT-compiled languages on closed-form mathematical workloads by eliminating the loop execution phase entirely:

| Benchmark | Tokype | C++ (`-O3`) | Python 3 |
| :--- | :---: | :---: | :---: |
| `sum i` ([example/benchmark.tp](example/benchmark.tp)) | **~0 sec** | **~0 sec** | ~0.5 sec |
| `fib(92)` ([example/fib.tp](example/fib.tp)) | **~0 sec** | Timeout | Timeout |
| `prod i` (`n = 20`, [example/prod.tp](example/prod.tp)) | **~0 sec** | **~0 sec** | **~0 sec** |

---

## Known Limitations

* **Type Coercion in Loop Folding:** String concatenation patterns inside supported accumulation loops may fold numerically (e.g., `"base" + i` $\rightarrow$ `"base55"`).
* **Integer Overflow:** Closed-form arithmetic relies on 64-bit signed integers (`int64`) and does not automatically promote to arbitrary-precision `BigInt`.
* **Pattern Coverage:** Complex loop bodies with nested conditionals, mutations of external state, or non-polynomial recurrences fall back to standard AST interpretation.

---

## Related Documentation

* [README.md](README.md) — project overview, installation, and CLI usage.
* [SPEC.md](SPEC.md) — full language syntax reference with examples.
