# Security Policy

## Supported Versions

Tokype is in early active development. Security updates are currently applied only to the latest released version.

| Version   | Supported          |
| --------- | ------------------ |
| `0.0.3`   | :white_check_mark: |
| `< 0.0.3` | :x:                |

## Security Model & Scope

Tokype is an AST-based interpreter written in Go

* **Execution Privileges:** In version `0.0.3`, scripts run with the same privileges as the host process without sandboxing. Executing untrusted third-party scripts is **not safe by design**, and host access by a script is not considered a vulnerability in this version.
* **In Scope:** Unhandled Go runtime panics, parser/AST crashes, or infinite hangs triggered during static AST optimization and loop folding when processing malformed or valid code.

## Reporting a Vulnerability

**Please do not open public GitHub Issues for security vulnerabilities.**

To keep communication private and organized, please submit all reports exclusively through GitHub's built-in **Private Vulnerability Reporting** feature:

1. Navigate to the **Security** tab of this repository.
2. Click **Advisories** in the left sidebar (or **Report a vulnerability**).
3. Fill out the private form with the details of the issue.

### What to include in your report:
* Tokype version (`0.0.3` or commit hash)
* A minimal code snippet reproducing the crash or optimizer hang.
* Expected vs. actual behavior.

## Response Timeline

Tokype is maintained by an independent developer:
* **Initial response:** Within **3–7 days**.
* **Disclosure:** Please refrain from publicly disclosing the issue until a fix is released.
