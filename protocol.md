# Binary Protocol Specification

## Overview

This document describes a simple binary protocol designed for use over TCP/SSL connections. The protocol supports variable-length payloads with different payload types for different message formats.

## Protocol Structure

### Header Format

```
┌─────────────┬─────────────┬─────────────┬─────────────┐
│   Version   │ Payload Type│ Payload Len │   Payload   │
│   1 byte    │   1 byte    │   4 bytes   │  variable   │
└─────────────┴─────────────┴─────────────┴─────────────┘
```

### Field Descriptions

- **Version** (1 byte): Protocol version, currently `0x00`
- **Payload Type** (1 byte): Type of payload data (see Payload Types section)
- **Payload Length** (4 bytes): Length of payload in bytes (little-endian)
- **Payload** (variable): Actual message data, format depends on payload type

## Endianness

All multi-byte integers are encoded in **little-endian** format (least significant byte first).

## Payload Types

### 0x00 - Single String
Payload contains a single string without length prefix.

### 0x01 - Key-Value Pair (PUSH_QUEUE)
Payload format:
```
┌─────────────┬─────────────┬─────────────┬─────────────┐
│  Key Length │     Key     │ Value Length│    Value    │
│   4 bytes   │  variable   │   4 bytes   │  variable   │
└─────────────┴─────────────┴─────────────┴─────────────┘
```

- **Key Length** (4 bytes): Length of key string (little-endian)
- **Key** (variable): Queue name as UTF-8 string
- **Value Length** (4 bytes): Length of value string (little-endian)  
- **Value** (variable): Message data as UTF-8 string

## Example Message

**PUSH_QUEUE message with queue="foo", message="bar"**

### Binary Representation (Hexadecimal)
```
00 01 08 00 00 00 03 00 00 00 66 6F 6F 03 00 00 00 62 61 72
```

### Breakdown
- `00` - Version (0)
- `01` - Payload Type (PUSH_QUEUE)
- `08 00 00 00` - Payload Length (8 bytes, little-endian)
- `03 00 00 00` - Key Length (3 bytes, little-endian)
- `66 6F 6F` - Key "foo" (UTF-8)
- `03 00 00 00` - Value Length (3 bytes, little-endian)
- `62 61 72` - Value "bar" (UTF-8)
