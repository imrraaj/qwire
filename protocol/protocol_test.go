package protocol

import (
	"bytes"
	"reflect"
	"testing"
)

func TestMarshal(t *testing.T) {
	tests := []struct {
		name     string
		input    Message
		expected []byte
		wantErr  bool
	}{
		{
			name:     "EMPTY false",
			input:    EmptyMessage{Value: false},
			expected: []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:     "EMPTY true",
			input:    EmptyMessage{Value: true},
			expected: []byte{0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01},
		},
		{
			name:     "CREATE_QUEUE",
			input:    CreateQueueMessage{QueueName: []byte("foo")},
			expected: []byte{0x01, 0x01, 0x03, 0x00, 0x00, 0x00, 0x66, 0x6F, 0x6F},
		},
		{
			name:     "JOIN_QUEUE",
			input:    JoinQueueMessage{QueueName: []byte("foo")},
			expected: []byte{0x01, 0x02, 0x03, 0x00, 0x00, 0x00, 0x66, 0x6F, 0x6F},
		},
		{
			name:     "PUSH_QUEUE",
			input:    PushQueueMessage{QueueName: []byte("foo"), MessageBody: []byte("bar")},
			expected: []byte{0x01, 0x03, 0x0A, 0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x66, 0x6F, 0x6F, 0x62, 0x61, 0x72},
		},
		{
			name:    "CREATE_QUEUE empty queue name",
			input:   CreateQueueMessage{QueueName: nil},
			wantErr: true,
		},
		{
			name:    "JOIN_QUEUE empty queue name",
			input:   JoinQueueMessage{QueueName: nil},
			wantErr: true,
		},
		{
			name:    "PUSH_QUEUE empty queue name",
			input:   PushQueueMessage{QueueName: nil, MessageBody: []byte("bar")},
			wantErr: true,
		},
		{
			name:    "nil message",
			input:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Marshal(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("Marshal() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Marshal() unexpected error: %v", err)
			}

			if !bytes.Equal(result, tt.expected) {
				t.Fatalf("Marshal() = %x, want %x", result, tt.expected)
			}
		})
	}
}

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected Message
		wantErr  bool
	}{
		{
			name:     "EMPTY false",
			input:    []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00},
			expected: EmptyMessage{Value: false},
		},
		{
			name:     "EMPTY true",
			input:    []byte{0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01},
			expected: EmptyMessage{Value: true},
		},
		{
			name:     "CREATE_QUEUE",
			input:    []byte{0x01, 0x01, 0x03, 0x00, 0x00, 0x00, 0x66, 0x6F, 0x6F},
			expected: CreateQueueMessage{QueueName: []byte("foo")},
		},
		{
			name:     "JOIN_QUEUE",
			input:    []byte{0x01, 0x02, 0x03, 0x00, 0x00, 0x00, 0x66, 0x6F, 0x6F},
			expected: JoinQueueMessage{QueueName: []byte("foo")},
		},
		{
			name:     "PUSH_QUEUE",
			input:    []byte{0x01, 0x03, 0x0A, 0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x66, 0x6F, 0x6F, 0x62, 0x61, 0x72},
			expected: PushQueueMessage{QueueName: []byte("foo"), MessageBody: []byte("bar")},
		},
		{
			name:    "frame too short",
			input:   []byte{0x01, 0x03, 0x0A},
			wantErr: true,
		},
		{
			name:    "unsupported version",
			input:   []byte{0x02, 0x00, 0x00, 0x00, 0x00, 0x00},
			wantErr: true,
		},
		{
			name:    "unknown payload type",
			input:   []byte{0x01, 0x7F, 0x00, 0x00, 0x00, 0x00},
			wantErr: true,
		},
		{
			name:    "truncated payload",
			input:   []byte{0x01, 0x01, 0x03, 0x00, 0x00, 0x00, 0x66, 0x6F},
			wantErr: true,
		},
		{
			name:    "trailing bytes",
			input:   []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF},
			wantErr: true,
		},
		{
			name:    "invalid EMPTY payload byte",
			input:   []byte{0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00},
			wantErr: true,
		},
		{
			name:    "invalid EMPTY payload length",
			input:   []byte{0x01, 0x00, 0x02, 0x00, 0x00, 0x00, 0x01, 0x01},
			wantErr: true,
		},
		{
			name:    "CREATE_QUEUE empty queue name",
			input:   []byte{0x01, 0x01, 0x00, 0x00, 0x00, 0x00},
			wantErr: true,
		},
		{
			name:    "JOIN_QUEUE empty queue name",
			input:   []byte{0x01, 0x02, 0x00, 0x00, 0x00, 0x00},
			wantErr: true,
		},
		{
			name:    "PUSH_QUEUE payload too short",
			input:   []byte{0x01, 0x03, 0x03, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00},
			wantErr: true,
		},
		{
			name:    "PUSH_QUEUE zero queue name length",
			input:   []byte{0x01, 0x03, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			wantErr: true,
		},
		{
			name:    "PUSH_QUEUE queue name length exceeds payload",
			input:   []byte{0x01, 0x03, 0x07, 0x00, 0x00, 0x00, 0x05, 0x00, 0x00, 0x00, 0x66, 0x6F, 0x6F},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Unmarshal(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("Unmarshal() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unmarshal() unexpected error: %v", err)
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Fatalf("Unmarshal() = %#v, want %#v", result, tt.expected)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		original Message
	}{
		{name: "EMPTY false", original: EmptyMessage{Value: false}},
		{name: "EMPTY true", original: EmptyMessage{Value: true}},
		{name: "CREATE_QUEUE", original: CreateQueueMessage{QueueName: []byte("queue")}},
		{name: "JOIN_QUEUE", original: JoinQueueMessage{QueueName: []byte("queue")}},
		{name: "PUSH_QUEUE", original: PushQueueMessage{QueueName: []byte("queue"), MessageBody: []byte{0x00, 0xFF, 0x01}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serialized, err := Marshal(tt.original)
			if err != nil {
				t.Fatalf("Marshal() error: %v", err)
			}

			parsed, err := Unmarshal(serialized)
			if err != nil {
				t.Fatalf("Unmarshal() error: %v", err)
			}

			if !reflect.DeepEqual(parsed, tt.original) {
				t.Fatalf("round trip = %#v, want %#v", parsed, tt.original)
			}
		})
	}
}

func TestPayloadTypeConstants(t *testing.T) {
	if EMPTY != 0 {
		t.Fatalf("EMPTY = %v, want 0", EMPTY)
	}

	if CREATE_QUEUE != 1 {
		t.Fatalf("CREATE_QUEUE = %v, want 1", CREATE_QUEUE)
	}

	if JOIN_QUEUE != 2 {
		t.Fatalf("JOIN_QUEUE = %v, want 2", JOIN_QUEUE)
	}

	if PUSH_QUEUE != 3 {
		t.Fatalf("PUSH_QUEUE = %v, want 3", PUSH_QUEUE)
	}
}

func BenchmarkUnmarshal(b *testing.B) {
	frame := []byte{0x01, 0x03, 0x0A, 0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x66, 0x6F, 0x6F, 0x62, 0x61, 0x72}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Unmarshal(frame)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshal(b *testing.B) {
	msg := PushQueueMessage{QueueName: []byte("foo"), MessageBody: []byte("bar")}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Marshal(msg)
		if err != nil {
			b.Fatal(err)
		}
	}
}
