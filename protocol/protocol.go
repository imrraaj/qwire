package protocol

import (
	"encoding/binary"
	"fmt"
)

type PayloadType uint8

const (
	CREATE_QUEUE PayloadType = iota
	PUSH_QUEUE   PayloadType = iota
)

func takeBytes(data *[]byte, n int) []byte {
	if n > len(*data) {
		result := *data
		*data = (*data)[:0]
		return result
	}
	result := (*data)[:n]
	*data = (*data)[n:]
	return result
}

func takeUint32(slice *[]byte) uint32 {
	bytes := takeBytes(slice, 4)
	if len(bytes) < 4 {
		return 0
	}
	return binary.LittleEndian.Uint32(bytes)
}

type Protocol struct {
	Version       uint8
	PayloadType   PayloadType
	Payload       []byte
	KeyValuePairs map[string]string
}

func ParseProtocol(msg []byte) (Protocol, error) {
	var p Protocol
	p.Version = msg[0]
	p.PayloadType = PayloadType(msg[1])
	p.Payload = msg[2:]
	p.KeyValuePairs = make(map[string]string)

	keyLength := takeUint32(&p.Payload)
	if keyLength == 0 {
		return Protocol{}, fmt.Errorf("Key length is zero")
	}
	if len(p.Payload) < int(keyLength) {
		return Protocol{}, fmt.Errorf("Payload too short for key length %d", keyLength)
	}
	keyBytes := takeBytes(&p.Payload, int(keyLength))
	if len(keyBytes) == 0 {
		return Protocol{}, fmt.Errorf("Key length is non-zero but no key bytes found")
	}
	valueLength := takeUint32(&p.Payload)
	if valueLength == 0 {
		p.KeyValuePairs[string(keyBytes)] = ""
		return p, nil
	}
	if len(p.Payload) < int(valueLength) {
		return Protocol{}, fmt.Errorf("Payload too short for value length %d", valueLength)
	}
	valueBytes := takeBytes(&p.Payload, int(valueLength))
	if len(valueBytes) == 0 {
		return Protocol{}, fmt.Errorf("Value length is non-zero but no value bytes found")
	}
	p.KeyValuePairs[string(keyBytes)] = string(valueBytes)
	return p, nil
}

func StringifyProtocol(p Protocol) ([]byte, error) {
	var payloadSize int
	for k, v := range p.KeyValuePairs {
		if len(k) != 0 { // Only include non-empty keys
			payloadSize += 4 + len(k) // 4 bytes for length of key + key
		}
		if len(v) != 0 { // Only include non-empty values
			payloadSize += 4 + len(v) // 4 bytes for length of value + value
		}
	}

	buf := make([]byte, 0, 6+payloadSize) // 6 bytes for header + payload
	buf = append(buf, byte(p.Version))
	buf = append(buf, uint8(p.PayloadType))

	for k, v := range p.KeyValuePairs {
		key := []byte(k)
		value := []byte(v)

		if len(key) != 0 {
			keyLenBuf := make([]byte, 4)
			binary.LittleEndian.PutUint32(keyLenBuf, uint32(len(key)))
			buf = append(buf, keyLenBuf...)
			buf = append(buf, key...)
		}
		if len(value) != 0 {
			valueLenBuf := make([]byte, 4)
			binary.LittleEndian.PutUint32(valueLenBuf, uint32(len(value)))
			buf = append(buf, valueLenBuf...)
			buf = append(buf, value...)
		}
	}
	return buf, nil
}
