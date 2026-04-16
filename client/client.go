package main

import (
	"binprot/protocol"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
)

func send(conn net.Conn, msg protocol.Message) {
	if err := protocol.WriteMessage(conn, msg); err != nil {
		log.Fatalf("write message: %v", err)
	}
}

func receive(conn net.Conn) protocol.Message {
	msg, err := protocol.ReadMessage(conn)
	if err != nil {
		if errors.Is(err, io.EOF) {
			log.Fatal("server closed the connection")
		}
		log.Fatalf("read message: %v", err)
	}
	return msg
}

func printMessage(prefix string, msg protocol.Message) {
	switch msg := msg.(type) {
	case protocol.EmptyMessage:
		fmt.Printf("%s EMPTY value=%t\n", prefix, msg.Value)
	case protocol.CreateQueueMessage:
		fmt.Printf("%s CREATE_QUEUE queue=%q\n", prefix, msg.QueueName)
	case protocol.JoinQueueMessage:
		fmt.Printf("%s JOIN_QUEUE queue=%q\n", prefix, msg.QueueName)
	case protocol.PushQueueMessage:
		fmt.Printf("%s PUSH_QUEUE queue=%q body=%q\n", prefix, msg.QueueName, msg.MessageBody)
	default:
		fmt.Printf("%s unknown message type %T\n", prefix, msg)
	}
}

func main() {
	conn, err := net.Dial("tcp", "localhost:6969")
	if err != nil {
		log.Fatalf("dial server: %v", err)
	}
	defer conn.Close()

	send(conn, protocol.CreateQueueMessage{QueueName: []byte("demo")})
	printMessage("<-", receive(conn))

	send(conn, protocol.JoinQueueMessage{QueueName: []byte("demo")})
	printMessage("<-", receive(conn))

	send(conn, protocol.PushQueueMessage{QueueName: []byte("demo"), MessageBody: []byte("hello from client")})
	printMessage("<-", receive(conn))
	printMessage("<-", receive(conn))
}
