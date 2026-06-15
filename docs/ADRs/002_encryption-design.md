# ADR 002: Encryption and Key Derivation Design

**Date:** 2026-06-13

---

# Context & Threat Model

Vaultify is designed to be a secure, developer-focused secret management platform.

Vaultify's security design assumes the database is compromised or leaked completely. If an attacker steals a full database dump, they possess all secret ciphertexts, nonces, and project-specific salts. My design guarantees they cannot read any secret data without the infrastructure's `MASTER_KEY`.

## Out of Scope

We do not defend against total application server takeover.

If an attacker gains:

- Root access
- Memory-dump capabilities
- Direct access to process memory

they can steal cryptographic material directly from the application's memory space.

---

# Architectural Decision

We use a single-tier encryption engine.

Instead of generating a unique key for every individual secret (envelope encryption), we derive a single encryption key per project and use it directly with AES-256-GCM.

## Key Derivation

```text
project_key = Argon2id(MASTER_KEY, project.salt)
```

## Symmetric Encryption

```text
ciphertext, nonce = AES-256-GCM.Encrypt(
    plaintext,
    project_key
)
```

---

# Nonce Management

AES-256-GCM requires a unique 12-byte nonce for every encryption operation.

Requirements:

- Generated using Go's cryptographically secure RNG (`crypto/rand`)
- Generated fresh for every encryption operation
- Stored alongside ciphertext
- Not treated as secret material

The following values may safely be stored in plaintext within the database:

- Ciphertext
- Nonce
- Project salt

## Why Nonce Uniqueness Matters

Nonce reuse under the same encryption key is catastrophic.

If two different plaintexts are encrypted using:

- The same `project_key`
- The same nonce

an attacker can XOR the resulting ciphertexts and recover information about the original plaintexts, ultimately breaking confidentiality guarantees.

---

# Key Rotation Tradeoff

## Current Design

```text
MASTER_KEY
    ↓
Argon2id
    ↓
project_key
    ↓
AES-256-GCM
```

## Alternative Not Chosen

Many enterprise systems implement a two-layer model:

```text
KEK
    ↓
Encrypts
    ↓
DEK
    ↓
Encrypts
    ↓
Secrets
```

This allows rapid key rotation without requiring all secrets to be re-encrypted.

## Tradeoff

If a project key leaks:

- Every secret in the project must be decrypted
- Every secret must be re-encrypted

We accept this operational cost in exchange for reduced complexity and improved maintainability at our current scale.

---

# Runtime Execution Contexts

The CLI and API server intentionally use different key-management strategies.

## CLI Context (`vaultify run`)

Characteristics:

- Short-lived process
- Used for secret injection and deployment workflows
- Derives keys once during execution
- Stores keys only in volatile memory
- Keys disappear when the process exits

Execution flow:

```text
Startup
    ↓
Argon2id
    ↓
Derived Keys
    ↓
Runtime Use
    ↓
Process Exit
    ↓
Memory Destroyed
```

## API Server Context

Characteristics:

- Long-lived process
- Handles HTTP/gRPC requests
- Maintains stateless cryptographic behavior
- Never caches derived keys

Execution flow:

```text
Request
    ↓
Argon2id
    ↓
Crypto Operation
    ↓
Discard Key
```

The API server recomputes Argon2id for every cryptographic read and write operation.

---

# Performance Considerations

Argon2id is intentionally expensive in both CPU and memory usage.

Computing it on every request:

- Increases latency
- Creates availability concerns
- Introduces potential Denial of Service (DoS) amplification opportunities

To avoid introducing external key caches such as Redis:

- No shared cache layer is used
- No additional attack surface is introduced

## CLI Optimization

For `vaultify run`:

- Key derivation occurs once at startup
- Derived keys remain only in local process memory
- Hot-path cryptographic operations avoid repeated KDF execution

---

# Why Argon2id Instead of SHA-256

A common question is:

> Why not simply hash the master key using SHA-256?

## SHA-256 Characteristics

- Extremely fast
- Requires negligible memory
- Can be executed billions of times per second on GPUs

If a weak master key is chosen, brute-force attacks become economically feasible.

## Argon2id Characteristics

Argon2id introduces a memory bottleneck.

Benefits:

- Requires physical RAM per guess
- Reduces GPU efficiency
- Slows large-scale brute-force attacks
- Provides stronger resistance against weak key configurations

---

# Argon2id Parameter Differentiation

Vaultify uses `golang.org/x/crypto/argon2` in two separate contexts.

## User Authentication

Purpose:

- Protect low-entropy human passwords

Configuration goals:

- High memory usage (64 MB+)
- Multiple iterations
- Approximately 300 ms verification cost

## Infrastructure Key Derivation

Purpose:

- Protect high-entropy infrastructure secrets

Configuration goals:

- Lower memory footprint
- Fast startup
- Minimal operational overhead

Target range:

- 8–16 MB memory
- Single iteration

---

# Fixed Production Parameters

To ensure predictable behavior across all deployments, Vaultify uses a fixed, non-configurable Argon2id configuration.

| Parameter | Value |
|------------|--------|
| Time Cost (`t`) | 1 |
| Memory Cost (`m`) | 16384 KB (16 MB) |
| Parallelism (`p`) | 2 |
| Key Length | 32 bytes |

## Rationale

### Time Cost = 1

- Minimum memory pass
- Prevents CPU starvation
- Keeps API derivations fast

### Memory Cost = 16 MB

- Introduces hardware resistance
- Limits GPU efficiency
- Avoids container-level out-of-memory risks during concurrency spikes

### Parallelism = 2

- Matches baseline container assumptions
- Uses multiple execution channels
- Avoids CPU scheduler oversaturation

### Key Length = 32 Bytes

Produces a key suitable for direct AES-256 usage:

```go
aes.NewCipher(projectKey)
```

No additional:

- Hashing
- Expansion
- Truncation

is required.

---

# Blast Radius Matrix

| Scenario | Outcome |
|-----------|----------|
| Database Leak Only | Attacker obtains ciphertexts, salts, and nonces but cannot decrypt without `MASTER_KEY`. **Defended.** |
| Environment Variable Leak Only | Attacker obtains `MASTER_KEY` but lacks ciphertext data. **Defended.** |
| Total Compromise | Attacker obtains both database contents and `MASTER_KEY`. **Complete compromise.** |

---

# Explicitly Accepted Risks

Vaultify is a storage-hardened platform and does not attempt to defend against host-level compromise.

## Process Memory Dumping

Examples:

- `ptrace`
- `/proc/[pid]/mem`
- Debugging utilities
- Memory scraping

Impact:

- Recovery of `MASTER_KEY`
- Recovery of active `project_key` values

## Environment Variable Exposure

Examples:

- `/proc/[pid]/environ`
- Container configuration leaks
- Container orchestration compromise

Impact:

- Immediate disclosure of `MASTER_KEY`

## Justification

These threats belong to:

- Operating system security
- Infrastructure hardening
- Container security
- Runtime isolation

Application-level cryptography cannot provide meaningful guarantees once the underlying host is untrusted.

---

# Decision Summary

Vaultify adopts a single-tier encryption architecture:

1. Derive one project-specific encryption key using Argon2id and a project salt.
2. Encrypt secrets directly using AES-256-GCM.
3. Generate a fresh 12-byte nonce for every encryption operation.
4. Store ciphertext, nonce, and salt in the database.
5. Keep the `MASTER_KEY` external to the database.
6. Accept project-wide re-encryption costs during key rotation in exchange for reduced complexity.
7. Use fixed Argon2id parameters optimized for infrastructure-secret protection.
8. Explicitly treat host compromise and process-memory extraction as out-of-scope threats.
