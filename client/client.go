package main

import (
	"binprot/protocol"
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

const demoQueueName = "demo"

func receive(conn net.Conn) (protocol.Message, error) {
	msg, err := protocol.ReadMessage(conn)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, errors.New("server closed the connection")
		}
		return nil, err
	}
	return msg, nil
}

func printMessage(prefix string, msg protocol.Message) {
	switch msg := msg.(type) {
	case protocol.EmptyMessage:
		fmt.Printf("\n%s STATUS\n", prefix)
		fmt.Printf("  ok: %t\n", msg.Value)
	case protocol.CreateQueueMessage:
		fmt.Printf("\n%s CREATE_QUEUE\n", prefix)
		fmt.Printf("  queue: %q\n", msg.QueueName)
	case protocol.JoinQueueMessage:
		fmt.Printf("\n%s JOIN_QUEUE\n", prefix)
		fmt.Printf("  queue: %q\n", msg.QueueName)
	case protocol.PushQueueMessage:
		fmt.Printf("\n%s PUSH_QUEUE\n", prefix)
		fmt.Printf("  queue: %q\n", msg.QueueName)
		fmt.Printf("  body : %q\n", msg.MessageBody)
	default:
		fmt.Printf("\n%s UNKNOWN\n", prefix)
		fmt.Printf("  type: %T\n", msg)
	}
}

func printInstructions() {
	fmt.Println()
	fmt.Println("Connected to the queue demo.")
	fmt.Printf("This client is subscribed to queue %q.\n", demoQueueName)
	fmt.Println("Anything you type here will be pushed to that queue.")
	fmt.Println("If you run this client in another terminal and type something there, it will also appear here because both clients are listening on the same queue.")
	fmt.Println("Press Enter after each message to send it.")
}

func send(conn net.Conn, msg protocol.Message) error {
	if err := protocol.WriteMessage(conn, msg); err != nil {
		return fmt.Errorf("write message: %w", err)
	}
	return nil
}

func expectStatus(conn net.Conn) (bool, error) {
	msg, err := receive(conn)
	if err != nil {
		return false, err
	}

	empty, ok := msg.(protocol.EmptyMessage)
	if !ok {
		return false, fmt.Errorf("expected EMPTY status response, got %T", msg)
	}

	return empty.Value, nil
}

func receiveLoop(conn net.Conn) {
	for {
		msg, err := receive(conn)
		if err != nil {
			log.Fatalf("read message: %v", err)
		}
		printMessage("<-", msg)
	}
}

func main() {
	conn, err := net.Dial("tcp", "localhost:6969")
	if err != nil {
		log.Fatalf("dial server: %v", err)
	}
	defer conn.Close()

	if err := send(conn, protocol.CreateQueueMessage{QueueName: []byte(demoQueueName)}); err != nil {
		log.Fatalf("create queue request: %v", err)
	}

	created, err := expectStatus(conn)
	if err != nil {
		log.Fatalf("create queue response: %v", err)
	}
	if created {
		fmt.Printf("Queue %q created.\n", demoQueueName)
	} else {
		fmt.Printf("Queue %q already exists or could not be created. Continuing with join.\n", demoQueueName)
	}

	if err := send(conn, protocol.JoinQueueMessage{QueueName: []byte(demoQueueName)}); err != nil {
		log.Fatalf("join queue request: %v", err)
	}

	joined, err := expectStatus(conn)
	if err != nil {
		log.Fatalf("join queue response: %v", err)
	}
	if !joined {
		log.Fatalf("failed to join queue %q", demoQueueName)
	}

	go receiveLoop(conn)

	printInstructions()
	if err := send(conn, protocol.PushQueueMessage{
		QueueName:   []byte(demoQueueName),
		MessageBody: []byte("hello from subscribed client"),
	}); err != nil {
		log.Fatalf("initial push queue request: %v", err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		if err := send(conn, protocol.PushQueueMessage{
			QueueName:   []byte(demoQueueName),
			MessageBody: []byte(line),
		}); err != nil {
			log.Fatalf("push queue request: %v", err)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("read stdin: %v", err)
	}
}
