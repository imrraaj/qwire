# Binary Protocol Specification

## Overview

This document defines the v1 binary protocol used by this repository. A single input buffer is expected to contain exactly one complete frame in this first implementation pass.

The protocol is binary-first. Payloads may contain arbitrary bytes and are not restricted to text.

## Terminology

- `frame`: one complete wire-format unit, including header and payload
- `header`: the first 6 bytes of a frame
- `payload`: the bytes after the header, whose length is declared by `payload_length`
- `raw frame bytes`: unparsed input bytes expected to contain exactly one frame
- `payload body`: the command-specific content inside the payload
- `queue name`: the queue identifier used by queue commands
- `message body`: arbitrary binary data carried by `PUSH_QUEUE`
- `payload type`: the command discriminator byte in the header

## Protocol Structure

### Header Format

```
┌─────────────┬──────────────┬────────────────────┬─────────────┐
│   Version   │ Payload Type │   Payload Length   │   Payload   │
│   1 byte    │    1 byte    │ 4 bytes, LE uint32 │  variable   │
└─────────────┴──────────────┴────────────────────┴─────────────┘
```

### Field Descriptions

- `version`: protocol version byte, currently `0x01`
- `payload_type`: command type for the payload
- `payload_length`: number of bytes in the payload, encoded in little-endian as an unsigned 32-bit integer
- `payload`: the command-specific payload body

## Endianness

All multi-byte integers are encoded in little-endian format.

## Framing Rules

- A valid frame must contain at least the 6-byte header.
- `payload_length` must exactly match the number of bytes after the header.
- A single input must contain exactly one frame.
- Trailing bytes after a complete frame are invalid.
- Malformed or truncated input must return an error.
- Parsing must never panic.

## Payload Types

### `0x00` - `EMPTY`

`EMPTY` is a minimal status-like message.

Allowed payload encodings:

- `payload_length = 0`: boolean-like `false`
- `payload_length = 1` and payload byte `0x01`: boolean-like `true`

All other `EMPTY` payloads are invalid.

### `0x01` - `CREATE_QUEUE`

Payload body:

```
┌─────────────┐
│ Queue Name  │
│  variable   │
└─────────────┘
```

- The entire payload is the queue name.
- Queue name must be non-empty.

### `0x02` - `JOIN_QUEUE`

Payload body:

```
┌─────────────┐
│ Queue Name  │
│  variable   │
└─────────────┘
```

- The entire payload is the queue name.
- Queue name must be non-empty.

### `0x03` - `PUSH_QUEUE`

Payload body:

```
┌────────────────────┬─────────────┬──────────────┐
│ Queue Name Length  │ Queue Name  │ Message Body │
│ 4 bytes, LE uint32 │  variable   │   variable   │
└────────────────────┴─────────────┴──────────────┘
```

- `queue_name_length` is the length of the queue name in bytes.
- `queue_name` must be non-empty.
- `message_body` contains arbitrary binary data.
- `message_body` length is derived from the outer `payload_length`.

## Validation Rules

- Reject frames with unknown `payload_type` values.
- Reject queue commands with empty queue names.
- Reject `PUSH_QUEUE` payloads shorter than 4 bytes.
- Reject `PUSH_QUEUE` frames whose `queue_name_length` exceeds the available payload bytes.
- Do not accept partial payloads as empty values.

## Examples

### `EMPTY` false

Hex:

```
01 00 00 00 00 00
```

Breakdown:

- `01` - Version
- `00` - Payload Type (`EMPTY`)
- `00 00 00 00` - Payload Length (`0`)
- no payload

### `EMPTY` true

Hex:

```
01 00 01 00 00 00 01
```

Breakdown:

- `01` - Version
- `00` - Payload Type (`EMPTY`)
- `01 00 00 00` - Payload Length (`1`)
- `01` - True

### `CREATE_QUEUE` with queue name `foo`

Hex:

```
01 01 03 00 00 00 66 6F 6F
```

Breakdown:

- `01` - Version
- `01` - Payload Type (`CREATE_QUEUE`)
- `03 00 00 00` - Payload Length (`3`)
- `66 6F 6F` - Queue name `foo`

### `JOIN_QUEUE` with queue name `foo`

Hex:

```
01 02 03 00 00 00 66 6F 6F
```

Breakdown:

- `01` - Version
- `02` - Payload Type (`JOIN_QUEUE`)
- `03 00 00 00` - Payload Length (`3`)
- `66 6F 6F` - Queue name `foo`

### `PUSH_QUEUE` with queue `foo` and message body `bar`

Hex:

```
01 03 0A 00 00 00 03 00 00 00 66 6F 6F 62 61 72
```

Breakdown:

- `01` - Version
- `03` - Payload Type (`PUSH_QUEUE`)
- `0A 00 00 00` - Payload Length (`10`)
- `03 00 00 00` - Queue name length (`3`)
- `66 6F 6F` - Queue name `foo`
- `62 61 72` - Message body `bar`

## Notes For This Version

- This version assumes one complete frame is available in one input buffer.
- Streamed framing across multiple TCP reads is intentionally deferred.
- The protocol-level maximum payload size is `2^32 - 1` bytes because `payload_length` is a 32-bit unsigned integer.
