package main

import (
	"binprot/protocol"
	"fmt"
)

func main() {
	msg := protocol.PushQueueMessage{
		QueueName:   []byte("foo"),
		MessageBody: []byte("bar"),
	}

	buf, err := protocol.Marshal(msg)
	if err != nil {
		fmt.Println("Error marshaling protocol message:", err)
		return
	}

	parsed, err := protocol.Unmarshal(buf)
	if err != nil {
		fmt.Println("Error unmarshaling protocol message:", err)
		return
	}

	switch msg := parsed.(type) {
	case protocol.PushQueueMessage:
		fmt.Printf("Version: %d\nPayload Type: %d\nQueue Name: %s\nMessage Body: %s\n", protocol.Version, msg.Type(), msg.QueueName, msg.MessageBody)
	default:
		fmt.Printf("Version: %d\nPayload Type: %d\n", protocol.Version, msg.Type())
	}
}
