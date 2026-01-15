package message

import (
	"testing"
)

func TestDiscoverMarshal(t *testing.T) {
	tests := []struct {
		name     string
		discover *Discover
		expected string
	}{
		{
			name:     "empty list",
			discover: &Discover{List: []string{}},
			expected: "DISCOVER,\n",
		},
		{
			name:     "single address",
			discover: &Discover{List: []string{"127.0.0.1:1378"}},
			expected: "DISCOVER,127.0.0.1:1378\n",
		},
		{
			name:     "multiple addresses",
			discover: &Discover{List: []string{"127.0.0.1:1378", "192.168.1.1:1379"}},
			expected: "DISCOVER,127.0.0.1:1378,192.168.1.1:1379\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.discover.Marshal()
			if result != tt.expected {
				t.Errorf("Marshal() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetMarshal(t *testing.T) {
	get := &Get{Name: "test.pdf"}
	expected := "Get,test.pdf\n"

	result := get.Marshal()
	if result != expected {
		t.Errorf("Marshal() = %q, want %q", result, expected)
	}
}

func TestFileMarshal(t *testing.T) {
	file := &File{Method: 1, TCPPort: 33680}
	expected := "File,1,33680\n"

	result := file.Marshal()
	if result != expected {
		t.Errorf("Marshal() = %q, want %q", result, expected)
	}
}

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectType  string
		expectError bool
	}{
		{
			name:        "discover message",
			input:       "DISCOVER,127.0.0.1:1378,192.168.1.1:1379",
			expectType:  "Discover",
			expectError: false,
		},
		{
			name:        "get message",
			input:       "Get,resume.pdf",
			expectType:  "Get",
			expectError: false,
		},
		{
			name:        "file message",
			input:       "File,1,33680",
			expectType:  "File",
			expectError: false,
		},
		{
			name:        "empty message",
			input:       "",
			expectType:  "",
			expectError: true,
		},
		{
			name:        "unknown message type",
			input:       "Unknown,data",
			expectType:  "",
			expectError: true,
		},
		{
			name:        "malformed get message",
			input:       "Get",
			expectType:  "",
			expectError: true,
		},
		{
			name:        "malformed file message",
			input:       "File,1",
			expectType:  "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Unmarshal(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Unmarshal() expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unmarshal() unexpected error: %v", err)
				return
			}

			switch tt.expectType {
			case "Discover":
				if _, ok := result.(*Discover); !ok {
					t.Errorf("Unmarshal() expected *Discover, got %T", result)
				}
			case "Get":
				if _, ok := result.(*Get); !ok {
					t.Errorf("Unmarshal() expected *Get, got %T", result)
				}
			case "File":
				if _, ok := result.(*File); !ok {
					t.Errorf("Unmarshal() expected *File, got %T", result)
				}
			}
		})
	}
}

func TestUnmarshalDiscover(t *testing.T) {
	input := "DISCOVER,127.0.0.1:1378,192.168.1.1:1379\n"
	result, err := Unmarshal(input)
	if err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}

	discover, ok := result.(*Discover)
	if !ok {
		t.Fatalf("Expected *Discover, got %T", result)
	}

	if len(discover.List) != 2 {
		t.Errorf("Expected 2 addresses, got %d", len(discover.List))
	}

	if discover.List[0] != "127.0.0.1:1378" {
		t.Errorf("First address = %q, want %q", discover.List[0], "127.0.0.1:1378")
	}
}

func TestUnmarshalGet(t *testing.T) {
	input := "Get,resume.pdf"
	result, err := Unmarshal(input)
	if err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}

	get, ok := result.(*Get)
	if !ok {
		t.Fatalf("Expected *Get, got %T", result)
	}

	if get.Name != "resume.pdf" {
		t.Errorf("Name = %q, want %q", get.Name, "resume.pdf")
	}
}

func TestUnmarshalFile(t *testing.T) {
	input := "File,1,33680"
	result, err := Unmarshal(input)
	if err != nil {
		t.Fatalf("Unmarshal() error: %v", err)
	}

	file, ok := result.(*File)
	if !ok {
		t.Fatalf("Expected *File, got %T", result)
	}

	if file.Method != 1 {
		t.Errorf("Method = %d, want %d", file.Method, 1)
	}
	if file.TCPPort != 33680 {
		t.Errorf("TCPPort = %d, want %d", file.TCPPort, 33680)
	}
}

func TestMarshalUnmarshalRoundTrip(t *testing.T) {
	t.Run("Discover", func(t *testing.T) {
		original := &Discover{List: []string{"127.0.0.1:1378", "192.168.1.1:1379"}}
		marshaled := original.Marshal()
		result, err := Unmarshal(marshaled)
		if err != nil {
			t.Fatalf("Unmarshal() error: %v", err)
		}

		discover, ok := result.(*Discover)
		if !ok {
			t.Fatalf("Expected *Discover, got %T", result)
		}

		if len(discover.List) != len(original.List) {
			t.Errorf("List length = %d, want %d", len(discover.List), len(original.List))
		}
	})

	t.Run("Get", func(t *testing.T) {
		original := &Get{Name: "test.pdf"}
		marshaled := original.Marshal()
		result, err := Unmarshal(marshaled)
		if err != nil {
			t.Fatalf("Unmarshal() error: %v", err)
		}

		get, ok := result.(*Get)
		if !ok {
			t.Fatalf("Expected *Get, got %T", result)
		}

		if get.Name != original.Name {
			t.Errorf("Name = %q, want %q", get.Name, original.Name)
		}
	})

	t.Run("File", func(t *testing.T) {
		original := &File{Method: 1, TCPPort: 33680}
		marshaled := original.Marshal()
		result, err := Unmarshal(marshaled)
		if err != nil {
			t.Fatalf("Unmarshal() error: %v", err)
		}

		file, ok := result.(*File)
		if !ok {
			t.Fatalf("Expected *File, got %T", result)
		}

		if file.Method != original.Method {
			t.Errorf("Method = %d, want %d", file.Method, original.Method)
		}
		if file.TCPPort != original.TCPPort {
			t.Errorf("TCPPort = %d, want %d", file.TCPPort, original.TCPPort)
		}
	})
}
