# chat.ssh

A tiny ephemeral chat server you connect to over SSH. One global room, no
accounts, no database, no history. When the process stops, everything is gone.

## Run

```sh
go build -o chat.ssh
./chat.ssh -addr :2222
```

A host key is generated on first run and written to `./host_key`.

Flags:

- `-addr` — listen address (default `:2222`)
- `-hostkey` — path to the SSH host key (default `host_key`, created if missing)

## Connect

```sh
ssh -p 2222 anything@<host>
```

Authentication is open — any username and any password/key are accepted; the
SSH-level user is ignored. You'll be prompted for a chat username (1–20
letters, digits, `-` or `_`) and then dropped into the room.

Each line you type is broadcast. Type `/quit` or close the connection to leave.

## Notes

- Joins and leaves are announced with the current online count.
- All state lives in memory; there is no persistence.
