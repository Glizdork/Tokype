# Tokype programming language

**Version**: 0.0.3

Creator: Glizdork.

The license is located in the same folder as the executable file. It is recommended to review the license before use.

## What is Tokype?

Tokype is a simple, fast programming language with:

- **Soft typing** — `"x = " + 5` works
- **TitanJerboa** — AST optimizer that replaces loops with formulas
- **Unique operators** — `"hello" - "l"`, `[1, 2] * 3`
- **Include** + namespaces
- **Null, float, unary minus**

## Quick start

```tokype
funct main():
    print("Hello, world!")
end

main()
```

## CLI commands

| Command | Description |
|---------|-------------|
| `tokype <file>.tp` | Run script |
| `tokype ast <file>.tp` | Show AST |
| `tokype version` | Show version |
| `tokype help` | Show help |
| `tokype license` | Show license |

## Documentation

- [**SPEC.md**](SPEC.md) — full syntax reference with examples
- [**TitanJerboa.md**](TitanJerboa.md) — optimizer description

Good luck!