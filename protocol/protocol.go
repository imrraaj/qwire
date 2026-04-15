package protocol

import (
	"encoding/binary"
	"errors"
	"fmt"
)

const (
	Version    uint8 = 0x01
	HeaderSize       = 6
	trueByte   byte  = 0x01
)

type PayloadType uint8

const (
	EMPTY PayloadType = iota
	CREATE_QUEUE
	JOIN_QUEUE
	PUSH_QUEUE
)

type Message interface {
	Type() PayloadType
}

type EmptyMessage struct {
	Value bool
}

func (EmptyMessage) Type() PayloadType {
	return EMPTY
}

type CreateQueueMessage struct {
	QueueName []byte
}

func (CreateQueueMessage) Type() PayloadType {
	return CREATE_QUEUE
}

type JoinQueueMessage struct {
	QueueName []byte
}

func (JoinQueueMessage) Type() PayloadType {
	return JOIN_QUEUE
}

type PushQueueMessage struct {
	QueueName   []byte
	MessageBody []byte
}

func (PushQueueMessage) Type() PayloadType {
	return PUSH_QUEUE
}

func Marshal(msg Message) ([]byte, error) {
	if msg == nil {
		return nil, errors.New("message is nil")
	}

	payload, err := marshalPayload(msg)
	if err != nil {
		return nil, err
	}

	if len(payload) > int(^uint32(0)) {
		return nil, fmt.Errorf("payload too large: %d", len(payload))
	}

	frame := make([]byte, HeaderSize+len(payload))
	frame[0] = Version
	frame[1] = byte(msg.Type())
	binary.LittleEndian.PutUint32(frame[2:HeaderSize], uint32(len(payload)))
	copy(frame[HeaderSize:], payload)
	return frame, nil
}

func Unmarshal(raw []byte) (Message, error) {
	if len(raw) < HeaderSize {
		return nil, fmt.Errorf("frame too short: got %d bytes, need at least %d", len(raw), HeaderSize)
	}

	version := raw[0]
	if version != Version {
		return nil, fmt.Errorf("unsupported version: %d", version)
	}

	payloadType := PayloadType(raw[1])
	payloadLength := binary.LittleEndian.Uint32(raw[2:HeaderSize])
	actualPayloadLength := len(raw) - HeaderSize
	if payloadLength != uint32(actualPayloadLength) {
		return nil, fmt.Errorf("payload length mismatch: declared %d, actual %d", payloadLength, actualPayloadLength)
	}

	payload := raw[HeaderSize:]
	return unmarshalPayload(payloadType, payload)
}

func marshalPayload(msg Message) ([]byte, error) {
	switch msg := msg.(type) {
	case EmptyMessage:
		if msg.Value {
			return []byte{trueByte}, nil
		}
		return nil, nil
	case CreateQueueMessage:
		return marshalQueueNamePayload(msg.QueueName)
	case JoinQueueMessage:
		return marshalQueueNamePayload(msg.QueueName)
	case PushQueueMessage:
		return marshalPushQueuePayload(msg)
	default:
		return nil, fmt.Errorf("unsupported message type %T", msg)
	}
}

func unmarshalPayload(payloadType PayloadType, payload []byte) (Message, error) {
	switch payloadType {
	case EMPTY:
		return unmarshalEmptyMessage(payload)
	case CREATE_QUEUE:
		return unmarshalCreateQueueMessage(payload)
	case JOIN_QUEUE:
		return unmarshalJoinQueueMessage(payload)
	case PUSH_QUEUE:
		return unmarshalPushQueueMessage(payload)
	default:
		return nil, fmt.Errorf("unknown payload type: %d", payloadType)
	}
}

func marshalQueueNamePayload(queueName []byte) ([]byte, error) {
	if len(queueName) == 0 {
		return nil, errors.New("queue name must not be empty")
	}
	return cloneBytes(queueName), nil
}

func marshalPushQueuePayload(msg PushQueueMessage) ([]byte, error) {
	if len(msg.QueueName) == 0 {
		return nil, errors.New("queue name must not be empty")
	}

	if len(msg.QueueName) > int(^uint32(0))-4-len(msg.MessageBody) {
		return nil, fmt.Errorf("payload too large: queue name %d bytes, message body %d bytes", len(msg.QueueName), len(msg.MessageBody))
	}

	payload := make([]byte, 4+len(msg.QueueName)+len(msg.MessageBody))
	binary.LittleEndian.PutUint32(payload[:4], uint32(len(msg.QueueName)))
	copy(payload[4:], msg.QueueName)
	copy(payload[4+len(msg.QueueName):], msg.MessageBody)
	return payload, nil
}

func unmarshalEmptyMessage(payload []byte) (Message, error) {
	switch len(payload) {
	case 0:
		return EmptyMessage{Value: false}, nil
	case 1:
		if payload[0] != trueByte {
			return nil, fmt.Errorf("invalid EMPTY payload byte: %d", payload[0])
		}
		return EmptyMessage{Value: true}, nil
	default:
		return nil, fmt.Errorf("invalid EMPTY payload length: %d", len(payload))
	}
}

func unmarshalCreateQueueMessage(payload []byte) (Message, error) {
	if len(payload) == 0 {
		return nil, errors.New("queue name must not be empty")
	}
	return CreateQueueMessage{QueueName: cloneBytes(payload)}, nil
}

func unmarshalJoinQueueMessage(payload []byte) (Message, error) {
	if len(payload) == 0 {
		return nil, errors.New("queue name must not be empty")
	}
	return JoinQueueMessage{QueueName: cloneBytes(payload)}, nil
}

func unmarshalPushQueueMessage(payload []byte) (Message, error) {
	if len(payload) < 4 {
		return nil, fmt.Errorf("PUSH_QUEUE payload too short: got %d bytes", len(payload))
	}

	queueNameLength := binary.LittleEndian.Uint32(payload[:4])
	if queueNameLength == 0 {
		return nil, errors.New("queue name must not be empty")
	}

	if queueNameLength > uint32(len(payload)-4) {
		return nil, fmt.Errorf("queue name length exceeds payload: declared %d, available %d", queueNameLength, len(payload)-4)
	}

	queueNameEnd := 4 + int(queueNameLength)
	return PushQueueMessage{
		QueueName:   cloneBytes(payload[4:queueNameEnd]),
		MessageBody: cloneBytes(payload[queueNameEnd:]),
	}, nil
}

func cloneBytes(src []byte) []byte {
	if len(src) == 0 {
		return nil
	}
	return append([]byte(nil), src...)
}
