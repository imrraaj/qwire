package main

import (
	"binprot/protocol"
	"net"
	"testing"
	"time"
)

func TestServerCreateJoinPushFlow(t *testing.T) {
	server, addr, shutdown := startTestServer(t)
	defer shutdown()

	_ = server

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("dial server: %v", err)
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("set deadline: %v", err)
	}

	if err := protocol.WriteMessage(conn, protocol.CreateQueueMessage{QueueName: []byte("demo")}); err != nil {
		t.Fatalf("write CREATE_QUEUE: %v", err)
	}
	assertEmptyResponse(t, conn, true)

	if err := protocol.WriteMessage(conn, protocol.JoinQueueMessage{QueueName: []byte("demo")}); err != nil {
		t.Fatalf("write JOIN_QUEUE: %v", err)
	}
	assertEmptyResponse(t, conn, true)

	push := protocol.PushQueueMessage{QueueName: []byte("demo"), MessageBody: []byte("hello")}
	if err := protocol.WriteMessage(conn, push); err != nil {
		t.Fatalf("write PUSH_QUEUE: %v", err)
	}
	assertPushMessage(t, conn, push)
	assertEmptyResponse(t, conn, true)
}

func TestServerRejectsMalformedPayload(t *testing.T) {
	_, addr, shutdown := startTestServer(t)
	defer shutdown()

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("dial server: %v", err)
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("set deadline: %v", err)
	}

	invalidEmptyTrue := []byte{0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00}
	if _, err := conn.Write(invalidEmptyTrue); err != nil {
		t.Fatalf("write invalid frame: %v", err)
	}

	assertEmptyResponse(t, conn, false)
}

func startTestServer(t *testing.T) (*Server, string, func()) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	server := NewServer("127.0.0.1", 0)
	serverDone := make(chan error, 1)
	go func() {
		serverDone <- server.ServeListener(listener)
	}()

	shutdown := func() {
		_ = listener.Close()
		select {
		case err := <-serverDone:
			if err != nil {
				t.Fatalf("server exited with error: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("timed out waiting for server shutdown")
		}
	}

	return server, listener.Addr().String(), shutdown
}

func assertEmptyResponse(t *testing.T, conn net.Conn, want bool) {
	t.Helper()

	msg, err := protocol.ReadMessage(conn)
	if err != nil {
		t.Fatalf("read EMPTY response: %v", err)
	}

	empty, ok := msg.(protocol.EmptyMessage)
	if !ok {
		t.Fatalf("response type = %T, want protocol.EmptyMessage", msg)
	}

	if empty.Value != want {
		t.Fatalf("EMPTY value = %t, want %t", empty.Value, want)
	}
}

func assertPushMessage(t *testing.T, conn net.Conn, want protocol.PushQueueMessage) {
	t.Helper()

	msg, err := protocol.ReadMessage(conn)
	if err != nil {
		t.Fatalf("read PUSH_QUEUE broadcast: %v", err)
	}

	push, ok := msg.(protocol.PushQueueMessage)
	if !ok {
		t.Fatalf("response type = %T, want protocol.PushQueueMessage", msg)
	}

	if string(push.QueueName) != string(want.QueueName) {
		t.Fatalf("queue name = %q, want %q", push.QueueName, want.QueueName)
	}

	if string(push.MessageBody) != string(want.MessageBody) {
		t.Fatalf("message body = %q, want %q", push.MessageBody, want.MessageBody)
	}
}
