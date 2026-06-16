# CLI Architecture & Secrets Injection

### 1. OS Execution Model

When `vaultify run -- node server.js` executes, it relies on the POSIX fork-exec system architecture. First, the OS duplicates the vaultify process via a fork system call, creating a child process with a separate memory space. The child immediately executes an execve system call, which tells the kernel to wipe the child's memory and replace it entirely with the node executable.

Crucially, the `execve` syscall is the exact boundary where secrets cross over: it accepts an array of environment pointers, allowing the kernel to write the secrets directly into the top of the newly allocated memory stack for the Node process. Meanwhile, the vaultify parent process executes a `waitpid()` syscall. The OS scheduler puts vaultify to sleep, consuming zero CPU cycles, until it receives a `SIGCHLD` signal indicating the Node process has terminated.

---

### 2. The Environment Variable Inheritance Model
To inject secrets, the CLI constructs the child's environment using `proc.Env = append(os.Environ(), secrets...)`. `os.Environ()` captures the baseline OS context of the terminal, such as `$PATH`, `$HOME`, and `$PWD`. The injected secrets are appended to this baseline to form the final environment array passed to `execve`.

If `os.Environ()` is omitted and only the secrets are passed, the child process boots into a vacuum. Without `$PATH`, the OS cannot resolve the node binary to execute it. Even if an absolute path is provided, modern applications will crash on startup because they rely on variables like `$HOME` to locate configurations or `$PWD` to resolve relative file paths. We must inherit the host environment to guarantee application stability.

---

### 3. Secrets Slice Zeroing Problem

In the specification, the code loops over the secrets slice and assigns an empty string (`secretEnvVars[i] = ""`). Because strings in Go are immutable, read-only byte arrays, this assignment does not overwrite the physical RAM containing the plaintext. It simply abandons the old memory pointer, leaving the secret intact on the heap until the garbage collector eventually overwrites it.

We explicitly document this limitation instead of hiding it because security theater is dangerous. In a garbage-collected language like Go, deterministic cryptographic wiping of immutable types is practically impossible without leveraging the `unsafe` package to manipulate internal string slice headers. Documenting this GC lag and the parent-child memory overlap acknowledges the operational reality: while secrets never touch the disk, they do linger in the parent process's RAM slightly longer than the exact moment of injection.

---

### 4. Config File & Threat Model
The vaultify login command generates a configuration file at `~/.vaultify/config` containing the host, the raw API token (vt_...), and the default_project. Because this file stores a highly privileged secret in plaintext, it must be created with strict 0600 POSIX file permissions, ensuring that only the current user (the owner) can read or write to it.

The threat model here focuses on local privilege boundaries. 0600 permissions successfully defend against lateral movement by other non-root users logged into the same host. However, it does not protect against malware executing under the context of the same user, nor does it protect against a root/kernel-level compromise. If an attacker gains shell execution as the developer, the token is considered compromised.

---

### 5. Memory Lifecycle of a Secret
The lifecycle begins when the CLI receives the TLS-encrypted API response. The Go runtime decrypts the payload into a mutable []byte buffer, which is unmarshaled into structured structs. To satisfy the `exec.Cmd.Env` signature, these structs are mapped into a newly allocated `[]string` slice (e.g., ["API_KEY=vt_123"]). Up to this point, the secrets exist strictly in the parent's heap memory.

During `execve`, the OS kernel reads the `[]string` slice and copies the bytes into the protected memory stack of the child process. Because the kernel copies rather than consumes this data, there is a brief concurrency window between `proc.Start()` executing and the parent's subsequent zeroing attempt where both the parent and child processes hold the plaintext secrets in memory simultaneously. After `proc.Start()` fires, the CLI attempts to zero the `[]string` slice in the parent process. Finally, the unreferenced strings are marked for garbage collection by the Go runtime, sitting in RAM until the non-deterministic GC sweep reclaims the memory pages.

---

### 6. Command Flow for `vaultify run`

1. **Config Read**: The CLI reads `~/.vaultify/config` to retrieve the API token and project context.

2. **API Call**: A GET request is dispatched to the Vaultify server, authenticating via the token to fetch the environment's secrets.

3. **Env Construction**: The CLI loops through the API response, formatting KEY=VALUE strings, and appends them to `os.Environ()`.

4. **Exec & Start**: The CLI configures `exec.CommandContext` and fires `proc.Start()`, triggering the OS-level fork-exec injection.

5. **Zeroing**: The CLI executes a best-effort reference clearing of the Go secrets slice.

6. **Wait & Exit**: The parent process blocks via `proc.Wait()`. When the child process finishes, the CLI captures the child's exit code and terminates itself using `os.Exit()`.

---

