package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"qwire/protocol"
	"sync"
)

type Server struct {
	Port   int
	Host   string
	mu     sync.Mutex
	queues map[string]*queue
}

type queue struct {
	messages    [][]byte
	subscribers map[*clientSession]struct{}
}

type clientSession struct {
	conn   net.Conn
	server *Server
	mu     sync.Mutex
	joined map[string]struct{}
}

func NewServer(host string, port int) *Server {
	return &Server{
		Port:   port,
		Host:   host,
		queues: make(map[string]*queue),
	}
}

func (s *Server) Serve() {
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Host, s.Port))
	if err != nil {
		log.Fatalf("could not initialize the server: %v", err)
	}
	defer ln.Close()

	if err := s.ServeListener(ln); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

func (s *Server) ServeListener(ln net.Listener) error {
	log.Printf("listening on %s\n", ln.Addr())
	for {
		conn, err := ln.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			return err
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	session := &clientSession{
		conn:   conn,
		server: s,
		joined: make(map[string]struct{}),
	}
	defer s.removeSession(session)

	for {
		msg, err := protocol.ReadMessage(conn)
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				return
			}
			log.Printf("invalid message from %s: %v", conn.RemoteAddr(), err)
			if writeErr := session.send(protocol.EmptyMessage{Value: false}); writeErr != nil {
				log.Printf("error writing failure response: %v", writeErr)
				return
			}
			continue
		}

		if err := s.handleMessage(session, msg); err != nil {
			log.Printf("error handling message from %s: %v", conn.RemoteAddr(), err)
			if writeErr := session.send(protocol.EmptyMessage{Value: false}); writeErr != nil {
				log.Printf("error writing failure response: %v", writeErr)
				return
			}
			continue
		}

		if err := session.send(protocol.EmptyMessage{Value: true}); err != nil {
			log.Printf("error writing success response: %v", err)
			return
		}
	}
}

func (s *Server) handleMessage(session *clientSession, msg protocol.Message) error {
	switch msg := msg.(type) {
	case protocol.EmptyMessage:
		return nil
	case protocol.CreateQueueMessage:
		return s.createQueue(msg.QueueName)
	case protocol.JoinQueueMessage:
		return s.joinQueue(session, msg.QueueName)
	case protocol.PushQueueMessage:
		return s.pushQueue(msg)
	default:
		return fmt.Errorf("unsupported message type %T", msg)
	}
}

func (s *Server) createQueue(queueName []byte) error {
	key := string(queueName)

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.queues[key]; exists {
		return fmt.Errorf("queue %q already exists", queueName)
	}

	s.queues[key] = &queue{subscribers: make(map[*clientSession]struct{})}
	return nil
}

func (s *Server) joinQueue(session *clientSession, queueName []byte) error {
	key := string(queueName)

	s.mu.Lock()
	defer s.mu.Unlock()

	q, exists := s.queues[key]
	if !exists {
		return fmt.Errorf("queue %q does not exist", queueName)
	}

	q.subscribers[session] = struct{}{}
	session.joined[key] = struct{}{}
	return nil
}

func (s *Server) pushQueue(msg protocol.PushQueueMessage) error {
	key := string(msg.QueueName)

	s.mu.Lock()
	q, exists := s.queues[key]
	if !exists {
		s.mu.Unlock()
		return fmt.Errorf("queue %q does not exist", msg.QueueName)
	}

	q.messages = append(q.messages, append([]byte(nil), msg.MessageBody...))
	subscribers := make([]*clientSession, 0, len(q.subscribers))
	for subscriber := range q.subscribers {
		subscribers = append(subscribers, subscriber)
	}
	s.mu.Unlock()

	for _, subscriber := range subscribers {
		if err := subscriber.send(msg); err != nil {
			log.Printf("error delivering queue message to %s: %v", subscriber.conn.RemoteAddr(), err)
		}
	}

	return nil
}

func (s *Server) removeSession(session *clientSession) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for queueName := range session.joined {
		q, exists := s.queues[queueName]
		if !exists {
			continue
		}
		delete(q.subscribers, session)
	}
}

func (s *clientSession) send(msg protocol.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return protocol.WriteMessage(s.conn, msg)
}

func main() {
	server := NewServer("localhost", 6969)
	server.Serve()
}
