package main

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

func serialize(m Message) []byte {
	buf := make([]byte, 0)
	buf = append(buf, byte(len(m.sender)))
	for _, a := range m.sender {
		buf = append(buf, byte(a))
	}

	buf = append(buf, byte(len(m.receiver)))
	for _, a := range m.receiver {
		buf = append(buf, byte(a))
	}
	buf = append(buf, byte(len(m.message)))
	for _, a := range m.message {
		buf = append(buf, byte(a))
	}
	fmt.Println(buf)
	return buf
}

type Message struct {
	id       time.Time
	sender   string
	receiver string
	message  string
}

func NewMessage(sender, receiver, message string) Message {
	return Message{
		sender:   sender,
		receiver: receiver,
		message:  message,
		id:       time.Now(),
	}
}

func main() {
	conn, err := net.Dial("tcp", "localhost:6969")
	if err != nil {
		log.Fatalf("ERROR: %s\n", err)
	}

	go (func(conn net.Conn) {
		for i := 0; i < 10; i++ {
			_, err := conn.Write(serialize(NewMessage("raj", "priya", "Hello, World\n")))
			if err != nil {
				log.Fatalf("ERROR: %s\n", err)
			}
		}
	})(conn)
	var wg sync.WaitGroup
	wg.Add(1)
	go (func(conn net.Conn) {
		for {
			buffer := make([]byte, 64*1024)
			_, err := conn.Read(buffer)

			if err != nil {
				log.Printf("error reading from connection: %s\n", err)
				return
			}
			fmt.Printf("%s\n", buffer)
			fmt.Println()
		}
	})(conn)
	wg.Wait()
}
