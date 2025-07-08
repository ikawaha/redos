package parser

import (
	"testing"
)

func TestStarHeight(t *testing.T) {
	tests := []struct {
		name     string
		regex    string
		expected int
	}{
		{"empty", "", 0},
		{"literal", "a", 0},
		{"star", "a*", 1},
		{"plus", "a+", 1},
		{"repeat", "a{1,3}", 1},
		{"nested", "(a|b)*c*", 1},
		{"complex", "(a|b)*c+d{2,4}", 1},
		{"deeply nested", "(((a|b)*c)*)*", 3},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r, err := Parse(test.regex)
			if err != nil {
				t.Fatalf("Failed to parse regex %q: %v", test.regex, err)
			}
			height := StarHeight(r)
			if height != test.expected {
				t.Errorf("StarHeight(%q) = %d; want %d", test.regex, height, test.expected)
			}
		})
	}
}

func TestMaxRepeat(t *testing.T) {
	tests := []struct {
		name     string
		regex    string
		expected int
	}{
		{"empty", "", 0},
		{"literal", "a", 0},
		{"no repeats", "a|b|c", 0},
		{"star", "a*", 0},
		{"plus", "a+", 0},
		{"repeat", "a{1,3}", 3},
		{"nested", "(a|b)*c*", 0},
		{"complex", "(a|b)*c+d{2,4}", 4},
		{"deeply nested", "(((a|b)*c)*)*", 0},
		{"multiple repeats", "a{2,5}b{3,6}c{9}", 9},
		{"unbounded repeat", "a{1,10}b{2,}", 10},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r, err := Parse(test.regex)
			if err != nil {
				t.Fatalf("Failed to parse regex %q: %v", test.regex, err)
			}
			max := MaxRepeat(r)
			if max != test.expected {
				t.Errorf("MaxRepeat(%q) = %d; want %d", test.regex, max, test.expected)
			}
		})
	}
}
