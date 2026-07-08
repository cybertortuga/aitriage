package engine

import (
	"math"
	"strings"
	"testing"
)

func TestCalculateShannonEntropy(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{
			name:     "empty string",
			input:    "",
			expected: 0.0,
		},
		{
			name:     "single character",
			input:    "a",
			expected: 0.0,
		},
		{
			name:     "all identical characters",
			input:    "aaaaa",
			expected: 0.0,
		},
		{
			name:     "two unique characters, equal frequency",
			input:    "ab",
			expected: 1.0,
		},
		{
			name:     "two unique characters, equal frequency multiple times",
			input:    "aabb",
			expected: 1.0,
		},
		{
			name:     "four unique characters",
			input:    "abcd",
			expected: 2.0,
		},
		{
			name:  "hello world",
			input: "hello world",
			// h:1, e:1, l:3, o:2, ' ':1, w:1, r:1, d:1
			// Total = 11
			// expected = 2.845350936622437
			expected: 2.845350936622437,
		},
		{
			name:     "base64 string",
			input:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			expected: 4.363826390090814,
		},
		{
			name:     "hex string",
			input:    "a1b2c3d4e5f6g7h8i9j0",
			expected: 4.321928094887363,
		},
		{
			name:     "long string of repeated characters",
			input:    strings.Repeat("z", 1000),
			expected: 0.0,
		},
		{
			name:     "all printable ascii characters",
			input:    " !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~",
			expected: 6.569855608330948,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := CalculateShannonEntropy(tt.input)
			if math.Abs(actual-tt.expected) > 1e-6 {
				t.Errorf("CalculateShannonEntropy() = %v, want %v", actual, tt.expected)
			}
		})
	}
}

func TestIsHighEntropy(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Short string (< 16 chars)",
			input:    "short",
			expected: false,
		},
		{
			name:     "Exactly 15 chars",
			input:    "123456789012345",
			expected: false,
		},
		{
			name:     "Exactly 16 chars (low entropy)",
			input:    "AAAAAAAAAAAAAAAA",
			expected: false,
		},
		{
			name:     "Long string (> 128 chars)",
			input:    strings.Repeat("A", 129),
			expected: false,
		},
		{
			name:     "High-entropy valid string (base64 token)",
			input:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9", // 36 chars, entropy ~4.36
			expected: true,
		},
		{
			name:     "High-entropy hex string",
			input:    "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", // 64 chars, entropy ~3.67
			expected: false,
		},
		{
			name:     "High-entropy random base62 string",
			input:    "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", // 62 chars, all unique
			expected: true,
		},
		{
			name:     "Low-entropy string of proper length",
			input:    "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
			expected: false,
		},
		{
			name:     "String with non-printable characters",
			input:    "invalid\x00character\x01test\x02here",
			expected: false,
		},
		{
			name:     "String with >30% non-alphanumeric characters",
			input:    "high!@#$%^&*()_+entropy~`{}|[]\\:;\"'<>,.?/",
			expected: false,
		},
		{
			name:     "Normal english sentence",
			input:    "this is a simple english sentence of length > 16",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsHighEntropy(tt.input)
			if result != tt.expected {
				t.Errorf("IsHighEntropy(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
