package main

import(
	"net"
	"fmt"
	"log"
)

type Server struct {
	Port      int
	Host      string
}

func NewServer(host string, port int) *Server {
	return &Server{
		Port:      port,
		Host:      host,
	}
}

func (s *Server) Serve() {
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Host, s.Port))
	if err != nil {
		log.Fatalf("Could not initialize the server: %s", err)
	}

	log.Printf("Listening on port %d...\n", s.Port)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Could not accept the connection...")
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {

	defer conn.Close()
	for {
		buffer := make([]byte, 64*1024)
		n, err := conn.Read(buffer)

		if err != nil {
			log.Printf("error reading from connection: %s\n", err)
			return
		}

		stringBuffer := string(buffer[:n])
		conn.Write([]byte(stringBuffer))
	}
}


func main() {
	server := NewServer("localhost", 6969)
	server.Serve()
}
