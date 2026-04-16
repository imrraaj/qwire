package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
)

func ReadMessage(r io.Reader) (Message, error) {
	header := make([]byte, HeaderSize)
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, err
	}

	payloadLength := binary.LittleEndian.Uint32(header[2:HeaderSize])
	payload := make([]byte, payloadLength)
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, err
	}

	frame := append(header, payload...)
	msg, err := Unmarshal(frame)
	if err != nil {
		return nil, fmt.Errorf("unmarshal message: %w", err)
	}
	return msg, nil
}

func WriteMessage(w io.Writer, msg Message) error {
	frame, err := Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	if _, err := w.Write(frame); err != nil {
		return err
	}
	return nil
}
