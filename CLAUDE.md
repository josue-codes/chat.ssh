# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

- Build: `go build -o chat.ssh`
- Run: `./chat.ssh -addr :2222 -hostkey host_key` (host key is generated on first run if missing)
- Vet: `go vet ./...`
- Sync deps: `go mod tidy`

There is no test suite yet. The local sandbox has no `ssh` client installed; to exercise the server end-to-end, write a short Go program that dials it with `golang.org/x/crypto/ssh` (use `ssh.InsecureIgnoreHostKey()` and any password — auth is open by design).

## Architecture

The server is a single-room ephemeral chat over SSH. Four files, one process, all state in memory.

**Hub / client fan-out (`hub.go`).** A `Hub` holds a mutex-protected `map[*Client]struct{}`. Each connected session owns a `Client` with a buffered `Out` channel (size 64). `Hub.broadcast` snapshots the client list under the lock, then sends to each `Out` with `select { case ... default: }` — **slow clients drop messages rather than blocking the hub**. Join/leave methods broadcast a system message (empty `From`) carrying the current online count.

**Session lifecycle (`session.go`).** `handleSession` requires a PTY, wraps the SSH channel in a `golang.org/x/term.Terminal`, prompts for a chat username (1–20 chars, alnum/`-`/`_`), then registers a `Client` with the hub. It spawns a **writer goroutine** that drains `client.Out` into the terminal while the main goroutine does `ReadLine` in a loop. `term.Terminal` is the linchpin here: its internal mutex lets concurrent `Write`s from the writer goroutine interleave with `ReadLine` on the main goroutine without trampling the user's in-progress input line. Cleanup order matters — on exit, `hub.Leave` runs (deferred) so the hub stops sending, then `close(client.Out)` lets the writer goroutine drain and exit, then we wait on `writerDone`.

**SSH server (`server.go`).** Uses `github.com/gliderlabs/ssh`. Both `PasswordHandler` and `PublicKeyHandler` unconditionally return true — the SSH-level user is ignored; identity in this app is the chat username typed at the prompt. On first start, an ed25519 host key is generated and written to disk; subsequent starts reuse it.

## Design constraints to preserve

- **Ephemeral.** No persistence, no history replay, no logging of message content. Don't add a database, file log, or message buffer for late joiners unless the user asks.
- **Single global room.** No DMs, threads, or channels.
- **Open auth.** Don't add SSH-level authentication; the chat username prompt is the only identity.
- **PTY required.** Non-PTY sessions are rejected; the line editor depends on it.
