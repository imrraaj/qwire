package protocol

import (
	"bytes"
	"reflect"
	"testing"
)

func TestParseProtocol(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected Protocol
		wantErr  bool
	}{
		{
			name: "Valid PUSH_QUEUE message",
			input: []byte{
				0x00,                   // Version
				0x01,                   // Payload Type (PUSH_QUEUE)
				0x03, 0x00, 0x00, 0x00, // Key Length (3)
				0x66, 0x6F, 0x6F, // Key "foo"
				0x03, 0x00, 0x00, 0x00, // Value Length (3)
				0x62, 0x61, 0x72, // Value "bar"
			},
			expected: Protocol{
				Version:     0,
				PayloadType: PUSH_QUEUE,
				KeyValuePairs: map[string]string{
					"foo": "bar",
				},
			},
			wantErr: false,
		},
		{
			name: "Valid CREATE_QUEUE message",
			input: []byte{
				0x00,                   // Version
				0x00,                   // Payload Type (CREATE_QUEUE)
				0x05, 0x00, 0x00, 0x00, // Key Length (5)
				0x71, 0x75, 0x65, 0x75, 0x65, // Key "queue"
				0x00, 0x00, 0x00, 0x00, // Value Length (0)
			},
			expected: Protocol{
				Version:     0,
				PayloadType: CREATE_QUEUE,
				KeyValuePairs: map[string]string{
					"queue": "",
				},
			},
			wantErr: false,
		},
		{
			name: "Empty key and value",
			input: []byte{
				0x00,                   // Version
				0x01,                   // Payload Type (PUSH_QUEUE)
				0x04, 0x00, 0x00, 0x00, // Key Length (4)
				0x74, 0x65, 0x73, 0x74, // Key "test"
				0x05, 0x00, 0x00, 0x00, // Value Length (5)
				0x76, 0x61, 0x6C, 0x75, 0x65, // Value "value"
			},
			expected: Protocol{
				Version:     0,
				PayloadType: PUSH_QUEUE,
				KeyValuePairs: map[string]string{
					"test": "value",
				},
			},
			wantErr: false,
		},
		{
			name: "Zero key length should error",
			input: []byte{
				0x00,                   // Version
				0x01,                   // Payload Type (PUSH_QUEUE)
				0x00, 0x00, 0x00, 0x00, // Key Length (0)
			},
			expected: Protocol{},
			wantErr:  true,
		},
		{
			name: "Key length exceeds available data",
			input: []byte{
				0x00,                   // Version
				0x01,                   // Payload Type (PUSH_QUEUE)
				0x10, 0x00, 0x00, 0x00, // Key Length (16) - too large
				0x66, 0x6F, 0x6F, // Only 3 bytes available
			},
			expected: Protocol{},
			wantErr:  true,
		},
		{
			name: "Value length exceeds available data",
			input: []byte{
				0x00,                   // Version
				0x01,                   // Payload Type (PUSH_QUEUE)
				0x03, 0x00, 0x00, 0x00, // Key Length (3)
				0x66, 0x6F, 0x6F, // Key "foo"
				0x10, 0x00, 0x00, 0x00, // Value Length (16) - too large
				0x62, 0x61, 0x72, // Only 3 bytes available
			},
			expected: Protocol{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseProtocol(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseProtocol() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseProtocol() unexpected error: %v", err)
				return
			}

			if result.Version != tt.expected.Version {
				t.Errorf("ParseProtocol() Version = %v, want %v", result.Version, tt.expected.Version)
			}

			if result.PayloadType != tt.expected.PayloadType {
				t.Errorf("ParseProtocol() PayloadType = %v, want %v", result.PayloadType, tt.expected.PayloadType)
			}

			if !reflect.DeepEqual(result.KeyValuePairs, tt.expected.KeyValuePairs) {
				t.Errorf("ParseProtocol() KeyValuePairs = %v, want %v", result.KeyValuePairs, tt.expected.KeyValuePairs)
			}
		})
	}
}

func TestStringifyProtocol(t *testing.T) {
	tests := []struct {
		name     string
		input    Protocol
		expected []byte
		wantErr  bool
	}{
		{
			name: "Valid PUSH_QUEUE message",
			input: Protocol{
				Version:     0,
				PayloadType: PUSH_QUEUE,
				KeyValuePairs: map[string]string{
					"foo": "bar",
				},
			},
			expected: []byte{
				0x00,                   // Version
				0x01,                   // Payload Type (PUSH_QUEUE)
				0x03, 0x00, 0x00, 0x00, // Key Length (3)
				0x66, 0x6F, 0x6F, // Key "foo"
				0x03, 0x00, 0x00, 0x00, // Value Length (3)
				0x62, 0x61, 0x72, // Value "bar"
			},
			wantErr: false,
		},
		{
			name: "Valid CREATE_QUEUE message with empty value",
			input: Protocol{
				Version:     0,
				PayloadType: CREATE_QUEUE,
				KeyValuePairs: map[string]string{
					"queue": "",
				},
			},
			expected: []byte{
				0x00,                   // Version
				0x00,                   // Payload Type (CREATE_QUEUE)
				0x05, 0x00, 0x00, 0x00, // Key Length (5)
				0x71, 0x75, 0x65, 0x75, 0x65, // Key "queue"
			},
			wantErr: false,
		},
		{
			name: "Empty key and non-empty value",
			input: Protocol{
				Version:     0,
				PayloadType: PUSH_QUEUE,
				KeyValuePairs: map[string]string{
					"": "value",
				},
			},
			expected: []byte{
				0x00,                   // Version
				0x01,                   // Payload Type (PUSH_QUEUE)
				0x05, 0x00, 0x00, 0x00, // Value Length (5)
				0x76, 0x61, 0x6C, 0x75, 0x65, // Value "value"
			},
			wantErr: false,
		},
		{
			name: "Both empty key and value",
			input: Protocol{
				Version:     0,
				PayloadType: CREATE_QUEUE,
				KeyValuePairs: map[string]string{
					"": "",
				},
			},
			expected: []byte{
				0x00, // Version
				0x00, // Payload Type (CREATE_QUEUE)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := StringifyProtocol(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("StringifyProtocol() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("StringifyProtocol() unexpected error: %v", err)
				return
			}

			if !bytes.Equal(result, tt.expected) {
				t.Errorf("StringifyProtocol() = %v, want %v", result, tt.expected)
				t.Errorf("StringifyProtocol() hex = %x, want %x", result, tt.expected)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		original Protocol
	}{
		{
			name: "PUSH_QUEUE round trip",
			original: Protocol{
				Version:     0,
				PayloadType: PUSH_QUEUE,
				KeyValuePairs: map[string]string{
					"myqueue": "hello world",
				},
			},
		},
		{
			name: "CREATE_QUEUE round trip",
			original: Protocol{
				Version:     0,
				PayloadType: CREATE_QUEUE,
				KeyValuePairs: map[string]string{
					"newqueue": "",
				},
			},
		},
		{
			name: "Unicode characters",
			original: Protocol{
				Version:     0,
				PayloadType: PUSH_QUEUE,
				KeyValuePairs: map[string]string{
					"测试": "🚀 unicode test",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Stringify the protocol
			serialized, err := StringifyProtocol(tt.original)
			if err != nil {
				t.Fatalf("StringifyProtocol() error: %v", err)
			}

			// Parse it back
			parsed, err := ParseProtocol(serialized)
			if err != nil {
				t.Fatalf("ParseProtocol() error: %v", err)
			}

			// Compare
			if parsed.Version != tt.original.Version {
				t.Errorf("Round trip Version = %v, want %v", parsed.Version, tt.original.Version)
			}

			if parsed.PayloadType != tt.original.PayloadType {
				t.Errorf("Round trip PayloadType = %v, want %v", parsed.PayloadType, tt.original.PayloadType)
			}

			if !reflect.DeepEqual(parsed.KeyValuePairs, tt.original.KeyValuePairs) {
				t.Errorf("Round trip KeyValuePairs = %v, want %v", parsed.KeyValuePairs, tt.original.KeyValuePairs)
			}
		})
	}
}

func TestPayloadTypeConstants(t *testing.T) {
	if CREATE_QUEUE != 0 {
		t.Errorf("CREATE_QUEUE = %v, want 0", CREATE_QUEUE)
	}

	if PUSH_QUEUE != 1 {
		t.Errorf("PUSH_QUEUE = %v, want 2", PUSH_QUEUE)
	}
}

func TestTakeBytes(t *testing.T) {
	tests := []struct {
		name      string
		input     []byte
		takeSize  int
		expected  []byte
		remaining []byte
	}{
		{
			name:      "Normal take",
			input:     []byte{1, 2, 3, 4, 5},
			takeSize:  3,
			expected:  []byte{1, 2, 3},
			remaining: []byte{4, 5},
		},
		{
			name:      "Take all",
			input:     []byte{1, 2, 3},
			takeSize:  3,
			expected:  []byte{1, 2, 3},
			remaining: []byte{},
		},
		{
			name:      "Take more than available",
			input:     []byte{1, 2},
			takeSize:  5,
			expected:  []byte{1, 2},
			remaining: []byte{},
		},
		{
			name:      "Take from empty slice",
			input:     []byte{},
			takeSize:  3,
			expected:  []byte{},
			remaining: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]byte, len(tt.input))
			copy(data, tt.input)

			result := takeBytes(&data, tt.takeSize)

			if !bytes.Equal(result, tt.expected) {
				t.Errorf("takeBytes() = %v, want %v", result, tt.expected)
			}

			if !bytes.Equal(data, tt.remaining) {
				t.Errorf("remaining data = %v, want %v", data, tt.remaining)
			}
		})
	}
}

func TestTakeUint32(t *testing.T) {
	tests := []struct {
		name      string
		input     []byte
		expected  uint32
		remaining []byte
	}{
		{
			name:      "Normal uint32",
			input:     []byte{0x08, 0x00, 0x00, 0x00, 0x01, 0x02},
			expected:  8,
			remaining: []byte{0x01, 0x02},
		},
		{
			name:      "Exactly 4 bytes",
			input:     []byte{0xFF, 0xFF, 0xFF, 0xFF},
			expected:  0xFFFFFFFF,
			remaining: []byte{},
		},
		{
			name:      "Less than 4 bytes",
			input:     []byte{0x01, 0x02},
			expected:  0,
			remaining: []byte{},
		},
		{
			name:      "Empty slice",
			input:     []byte{},
			expected:  0,
			remaining: []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make([]byte, len(tt.input))
			copy(data, tt.input)

			result := takeUint32(&data)

			if result != tt.expected {
				t.Errorf("takeUint32() = %v, want %v", result, tt.expected)
			}

			if !bytes.Equal(data, tt.remaining) {
				t.Errorf("remaining data = %v, want %v", data, tt.remaining)
			}
		})
	}
}

// Benchmark tests
func BenchmarkParseProtocol(b *testing.B) {
	msg := []byte{
		0x00,                   // Version
		0x01,                   // Payload Type (PUSH_QUEUE)
		0x03, 0x00, 0x00, 0x00, // Key Length (3)
		0x66, 0x6F, 0x6F, // Key "foo"
		0x03, 0x00, 0x00, 0x00, // Value Length (3)
		0x62, 0x61, 0x72, // Value "bar"
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParseProtocol(msg)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStringifyProtocol(b *testing.B) {
	p := Protocol{
		Version:     0,
		PayloadType: PUSH_QUEUE,
		KeyValuePairs: map[string]string{
			"foo": "bar",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := StringifyProtocol(p)
		if err != nil {
			b.Fatal(err)
		}
	}
}
