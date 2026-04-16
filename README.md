# binprot

`binprot` is a small Go project for experimenting with a binary application protocol and a TCP client/server that speak it end to end.

The current v1 work focuses on making these parts agree with each other:

- `protocol.md`: wire-format specification
- `protocol/`: encoding, decoding, and stream I/O helpers
- `server/`: in-memory queue server using protocol messages
- `client/`: demo client using protocol messages

## Current Protocol

Each frame has a fixed 6-byte header followed by a variable-length payload:

```text
+---------+--------------+----------------+---------+
| version | payload_type | payload_length | payload |
| 1 byte  |   1 byte     | 4 bytes, LE    |   ...   |
+---------+--------------+----------------+---------+
```

- protocol version: `0x01`
- payload length: little-endian unsigned 32-bit integer
- one input buffer is expected to contain exactly one complete frame in this implementation pass

Supported payload types:

- `0x00` `EMPTY`
- `0x01` `CREATE_QUEUE`
- `0x02` `JOIN_QUEUE`
- `0x03` `PUSH_QUEUE`

`EMPTY` is used as a minimal boolean-like status message:

- zero-length payload means `false`
- one-byte payload with value `0x01` means `true`

The full protocol reference lives in `protocol.md`.

## Repository Layout

- `main.go`: simple local marshal/unmarshal demo
- `protocol/`: typed protocol messages plus `Marshal`, `Unmarshal`, `ReadMessage`, and `WriteMessage`
- `server/`: TCP server with in-memory queues and subscriber delivery
- `client/`: TCP client demo for create/join/push flows
- `v1-findings.md`: implementation notes and design decisions captured during the restart work

## Requirements

- Go `1.24.4` or compatible toolchain

## Quick Start

Run the root protocol demo:

```bash
make run
```

Run the server in one terminal:

```bash
make run-server
```

Run the client in another terminal:

```bash
make run-client
```

The demo client will:

1. create queue `demo`
2. join queue `demo`
3. push a message into `demo`

The server replies with `EMPTY true` for successful commands and broadcasts `PUSH_QUEUE` messages to subscribed clients.

## Development

Common commands:

```bash
make fmt
make test
make test-race
make test-cover
make vet
make build
```

Built binaries are written to `build/`.

## Testing

The repository includes:

- protocol unit tests for exact frame encoding/decoding
- malformed and truncated frame tests
- TCP integration tests for create/join/push behavior

Run everything with:

```bash
make test
make test-race
make vet
```

## Current Limitations

- streamed frame assembly across multiple TCP reads is still deferred
- there is no persistence; queue state is in memory only
- response semantics are intentionally minimal and currently use `EMPTY` as success/failure status

## Notes

- stale generated artifacts from earlier prototype work were removed from the repository root
- queue names are handled as raw `[]byte` in the protocol layer; the server converts them to strings for in-memory storage
