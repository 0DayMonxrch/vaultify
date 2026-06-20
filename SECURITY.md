# Security Policy

## Supported Versions

Vaultify strictly maintains security patches for the latest major version.

| Version | Supported          |
| ------- | ------------------ |
| 2.x.x   | :white_check_mark: |
| < 2.0   | :x:                |

## Reporting a Vulnerability

Security is a core focus for Vaultify. If you discover a vulnerability, please do **not** open a public issue on GitHub. 

Instead, please responsibly disclose it by emailing me directly at **zdmonarch.tech@gmail.com**.

I will acknowledge receipt of your vulnerability report within 48 hours and provide regular updates about the remediation progress. If the vulnerability is confirmed, a patch will be issued, and you will be credited in the release notes.

## Out of Scope Vulnerabilities

As detailed in our [Encryption Design Architecture](docs/ADRs/002_encryption-design.md), Vaultify explicitly accepts certain risks as out-of-scope for the application layer. The following scenarios are **not** considered valid vulnerabilities for this project:

- **Total Host Compromise:** Attacks requiring root or administrative access to the underlying server.
- **Process Memory Dumping:** Extraction of secrets via debugging tools (`ptrace`), `/proc/[pid]/mem`, or core dumps from a compromised host environment.
- **Local Context Hijacking:** Reading the `~/.vaultify/config` file from a malicious process executing under the exact same user account on the client machine.
- **Ciphertext Recovery:** Extracting ciphertext from a database dump (decryption is impossible without the infrastructure-level `MASTER_KEY` environment variable).
