<div align="center">
  <img src="assets/logo.png" alt="Tokype Logo" width="250" />
  <h1>Tokype</h1>
  <p><strong>A simple, expressive, AST-interpreted programming language with static loop-folding optimization.</strong></p>
  <p><em>This language is held together by the creator's luck.</em></p>

  <p>
    <a href="https://github.com/Glizdork/Tokype/tags"><img src="https://img.shields.io/github/v/tag/Glizdork/Tokype" alt="Version" /></a>
    <img src="https://img.shields.io/github/go-mod/go-version/Glizdork/Tokype" alt="Go Version" />
    <a href="License"><img src="https://img.shields.io/github/license/Glizdork/Tokype" alt="License" /></a>
  </p>
</div>


---

## About Tokype

**Tokype** is an interpreted programming language written in Go, designed to combine readable, beginner-friendly syntax with high execution speed.

Unlike traditional tree-walking interpreters that blindly execute every loop iteration at runtime, Tokype passes your code through **TitanJerboa** — a built-in static AST optimizer that folds predictable loops, pre-evaluates pure functions, and turns $O(n)$ bottlenecks into $O(1)$ mathematical formulas before execution even begins.

### Key Highlights
* **TitanJerboa AST Optimizer** — automatically collapses summation/product loops, recursive functions (`fib(92)` in 0s), and list allocations (`repeat`) prior to runtime.
* **Soft Typing** — pragmatic type coercion where expressions like `"x = " + 5` work naturally without boilerplate casts.
* **Expressive Operators** — intuitive operations on strings and lists, such as `"hello" - "l"` or ` [1, 2] * 3`.
* **Modularity & Namespaces** — clean code organization via `include` directives, namespaces, and built-in standard modules.
* **Lightweight & Portable** — zero external runtime dependencies; distributed as a single standalone binary.

---

## Quick Look

```tokype
funct main():
    // Soft typing & expressive operators
    greeting = "Hello, world!!!" - "!"
    print(greeting)

    // TitanJerboa optimizes this O(n) loop into a single O(1) formula
    sum = 0
    for i = 1, 1000000:
        sum = sum + i
    end
    print("Sum = " + sum)
end

main()
```

---

## Installation

### Option 1: Pre-built Binaries (Recommended)
1. Download the archive for your OS (Linux, Windows, or macOS) from the [Releases](https://github.com) page.
2. Extract the archive and add the `tokype` binary to your system `PATH`.

> **Note:** The `License` file is included inside the release archive. Please review it before use.  
> **Windows Users:** Because release binaries are currently unsigned, Windows Defender SmartScreen may occasionally flag the executable as a false positive. You can verify file integrity against `checksums.txt` in the release assets or build directly from source.

### Option 2: Build from Source (requires Go)
```bash
go install ://github.com
```

---

## CLI Usage

| Command | Description |
| :--- | :--- |
| `tokype <file>.tp` | Execute a Tokype script |
| `tokype ast <file>.tp` | Print the parsed & optimized AST |
| `tokype version` | Display the installed Tokype version |
| `tokype help` | Show CLI usage instructions |
| `tokype license` | Display the BSD-3-Clause license |

---

## Documentation

* [**SPEC.md**](SPEC.md) — complete language specification, syntax rules, and built-in functions.
* [**TitanJerboa.md**](TitanJerboa.md) — technical breakdown of the AST optimizer passes and benchmarks.
* [**SECURITY.md**](SECURITY.md) — security model and private vulnerability reporting.
* [**CONTRIBUTING.md**](CONTRIBUTING.md) — how to contribute to the project.

---

## License

Tokype is open-source software licensed under the [BSD-3-Clause License](License).  
Created by [Glizdork](https://github.com).

**Good luck!** 
