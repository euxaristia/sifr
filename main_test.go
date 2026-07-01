package main

import (
	"os"
	"testing"
)

func TestDecryptCaesar(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		shift    int
		expected string
	}{
		{
			name:     "ROT13 basic lowercase",
			input:    "cvpbpgs",
			shift:    13,
			expected: "picoctf",
		},
		{
			name:     "ROT13 mixed case and symbols",
			input:    "cvpbPGS{guvf_vf_n_grfg}",
			shift:    13,
			expected: "picoCTF{this_is_a_test}",
		},
		{
			name:     "Shift 5 right",
			input:    "abcxyzABCxyz",
			shift:    5,
			expected: "fghcdeFGHcde",
		},
		{
			name:     "Shift negative (-5)",
			input:    "fghcdeFGHcde",
			shift:    -5,
			expected: "abcxyzABCxyz",
		},
		{
			name:     "Shift 0",
			input:    "Hello, World! 123",
			shift:    0,
			expected: "Hello, World! 123",
		},
		{
			name:     "Large shift (>26)",
			input:    "abc",
			shift:    27, // Equivalent to 1
			expected: "bcd",
		},
		{
			name:     "Large negative shift (<-26)",
			input:    "bcd",
			shift:    -27, // Equivalent to -1 (or +25)
			expected: "abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := decryptCaesar(tt.input, tt.shift)
			if actual != tt.expected {
				t.Errorf("decryptCaesar(%q, %d) = %q; want %q", tt.input, tt.shift, actual, tt.expected)
			}
		})
	}
}

func TestContainsPattern(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		pattern  string
		expected bool
	}{
		{"exact match", "picoCTF{flag}", "picoCTF", true},
		{"case insensitive match", "PICOctf{flag}", "picoctf", true},
		{"pattern not in text", "hello world", "picoCTF", false},
		{"empty pattern", "hello world", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := containsPattern(tt.text, tt.pattern)
			if actual != tt.expected {
				t.Errorf("containsPattern(%q, %q) = %v; want %v", tt.text, tt.pattern, actual, tt.expected)
			}
		})
	}
}

func TestHighlight(t *testing.T) {
	text := "picoCTF{hello_world}"
	pattern := "picoCTF"

	// Color enabled
	actualColored := highlight(text, pattern, true)
	expectedColored := "\033[1;32mpicoCTF\033[0m{hello_world}"
	if actualColored != expectedColored {
		t.Errorf("highlight(..., true) = %q; want %q", actualColored, expectedColored)
	}

	// Color disabled
	actualPlain := highlight(text, pattern, false)
	if actualPlain != text {
		t.Errorf("highlight(..., false) = %q; want %q", actualPlain, text)
	}

	// Case insensitive highlight
	textMixed := "PicoCtf{hello_world}"
	actualMixed := highlight(textMixed, pattern, true)
	expectedMixed := "\033[1;32mPicoCtf\033[0m{hello_world}"
	if actualMixed != expectedMixed {
		t.Errorf("highlight(mixed case, true) = %q; want %q", actualMixed, expectedMixed)
	}
}

func TestFileExists(t *testing.T) {
	if fileExists("non_existent_file_xyz_123.txt") {
		t.Error("expected fileExists to return false for non-existent file")
	}
	if !fileExists("main.go") {
		t.Error("expected fileExists to return true for main.go")
	}
}

func TestProcessShift(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		shift    int
		pattern  string
		expected string
	}{
		{
			name:     "picoCTF prefix already plaintext, shift inner only",
			input:    "picoCTF{bqnrrhmfsgdqtahbnmphkgrwqj}",
			shift:    1, // b->c, q->r, etc.
			pattern:  "picoCTF",
			expected: "picoCTF{crossingtherubiconqilhsxrk}",
		},
		{
			name:     "picoCTF prefix not plaintext, shift entire string",
			input:    "cvpbPGS{guvf_vf_n_grfg}",
			shift:    13,
			pattern:  "picoCTF",
			expected: "picoCTF{this_is_a_test}",
		},
		{
			name:     "custom prefix already plaintext, shift inner only",
			input:    "flag{abc}",
			shift:    1,
			pattern:  "flag",
			expected: "flag{bcd}",
		},
		{
			name:     "no brackets, shift entire string",
			input:    "cvpbPGS",
			shift:    13,
			pattern:  "picoCTF",
			expected: "picoCTF",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := processShift(tt.input, tt.shift, tt.pattern)
			if actual != tt.expected {
				t.Errorf("processShift(%q, %d, %q) = %q; want %q", tt.input, tt.shift, tt.pattern, actual, tt.expected)
			}
		})
	}
}

func TestDecryptNewCaesar(t *testing.T) {
	input := "fegdeogdgecoeocgcgchcfcffccfca"
	expected := "et_tu?_77866c61"
	actual, err := decryptNewCaesar(input, 15)
	if err != nil {
		t.Fatalf("decryptNewCaesar returned error: %v", err)
	}
	if actual != expected {
		t.Errorf("decryptNewCaesar(%q, 15) = %q; want %q", input, actual, expected)
	}

	_, err = decryptNewCaesar("xyz", 0)
	if err == nil {
		t.Error("expected error for invalid characters, but got nil")
	}

	_, err = decryptNewCaesar("abc", 0)
	if err == nil {
		t.Error("expected error for odd length, but got nil")
	}
}

func TestIsValidNewCaesar(t *testing.T) {
	err := isValidNewCaesar("picoCTF{fegdeogdgecoeocgcgchcfcffccfca}", "picoCTF")
	if err != nil {
		t.Errorf("isValidNewCaesar returned error for valid input: %v", err)
	}

	err = isValidNewCaesar("picoCTF{fegdeogdgecoeocgcgchcfcffccfcz}", "picoCTF")
	if err == nil {
		t.Error("expected error for invalid character, but got nil")
	}

	err = isValidNewCaesar("picoCTF{feg}", "picoCTF")
	if err == nil {
		t.Error("expected error for odd length, but got nil")
	}
}

func TestIsPrintable(t *testing.T) {
	if !isPrintable("hello world") {
		t.Error("expected 'hello world' to be printable")
	}
	if isPrintable("hello\x01world") {
		t.Error("expected string with control char to not be printable")
	}
}

func TestParsePowerShellEncodedCommand(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()

	// Test with matching -encodedCommand (UTF-16LE base64 for "fegdeogdgecoeocgcgchcfcffccfca")
	os.Args = []string{"sifr", "picoCTF", "-encodedCommand", "ZgBlAGcAZABlAG8AZwBkAGcAZQBjAG8AZQBvAGMAZwBjAGcAYwBoAGMAZgBjAGYAZgBjAGMAZgBjAGEA"}
	expected := "picoCTF{fegdeogdgecoeocgcgchcfcffccfca}"
	actual, ok := parsePowerShellEncodedCommand("picoCTF")
	if !ok {
		t.Error("expected parsePowerShellEncodedCommand to return true, but got false")
	}
	if actual != expected {
		t.Errorf("parsePowerShellEncodedCommand(...) = %q; want %q", actual, expected)
	}

	// Test without -encodedCommand
	os.Args = []string{"sifr", "picoCTF"}
	_, ok = parsePowerShellEncodedCommand("picoCTF")
	if ok {
		t.Error("expected parsePowerShellEncodedCommand to return false when flag is absent, but got true")
	}
}
