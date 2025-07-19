package redos

import (
	"regexp/syntax"
	"testing"
)

func TestRegexSyntax_toSyntaxFlags(t *testing.T) {
	tests := []struct {
		name     string
		syntax   RegexSyntax
		expected syntax.Flags
	}{
		{
			name:     "perl_syntax",
			syntax:   SyntaxPerl,
			expected: syntax.Perl,
		},
		{
			name:     "posix_syntax",
			syntax:   SyntaxPOSIX,
			expected: syntax.POSIX,
		},
		{
			name:     "invalid_syntax_defaults_to_perl",
			syntax:   RegexSyntax(999),
			expected: syntax.Perl,
		},
		{
			name:     "zero_value_defaults_to_perl",
			syntax:   RegexSyntax(0),
			expected: syntax.Perl,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.syntax.toSyntaxFlags()
			if result != test.expected {
				t.Errorf("toSyntaxFlags() = %v; want %v", result, test.expected)
			}
		})
	}
}

func TestNewDefaultLimit(t *testing.T) {
	limit := NewDefaultLimit()
	if limit == nil {
		t.Fatal("NewDefaultLimit() returned nil")
	}
	if limit.MaxRepeat != DefaultMaxRepeatLimit {
		t.Errorf("MaxRepeat = %v; want %v", limit.MaxRepeat, DefaultMaxRepeatLimit)
	}
	if limit.StarHeight != DefaultStarHeightLimit {
		t.Errorf("StarHeight = %v; want %v", limit.StarHeight, DefaultStarHeightLimit)
	}
}

func TestNewValidator(t *testing.T) {
	t.Run("default_validator", func(t *testing.T) {
		v := NewValidator()
		if v == nil {
			t.Fatal("NewValidator() returned nil")
		}
		if v.limit == nil {
			t.Fatal("Validator.limit is nil")
		}
		if v.limit.MaxRepeat != DefaultMaxRepeatLimit {
			t.Errorf("MaxRepeat = %v; want %v", v.limit.MaxRepeat, DefaultMaxRepeatLimit)
		}
		if v.limit.StarHeight != DefaultStarHeightLimit {
			t.Errorf("StarHeight = %v; want %v", v.limit.StarHeight, DefaultStarHeightLimit)
		}
		if v.syntax != syntax.Perl {
			t.Errorf("syntax = %v; want %v", v.syntax, syntax.Perl)
		}
	})

	t.Run("with_custom_limit", func(t *testing.T) {
		customLimit := &Complexity{
			MaxRepeat:  100,
			StarHeight: 2,
		}
		v := NewValidator(WithLimit(customLimit))
		if v.limit.MaxRepeat != 100 {
			t.Errorf("MaxRepeat = %v; want 100", v.limit.MaxRepeat)
		}
		if v.limit.StarHeight != 2 {
			t.Errorf("StarHeight = %v; want 2", v.limit.StarHeight)
		}
	})

	t.Run("with_custom_syntax", func(t *testing.T) {
		v := NewValidator(WithSyntax(SyntaxPOSIX))
		if v.syntax != syntax.POSIX {
			t.Errorf("syntax = %v; want %v", v.syntax, syntax.POSIX)
		}
	})

	t.Run("with_multiple_options", func(t *testing.T) {
		customLimit := &Complexity{
			MaxRepeat:  200,
			StarHeight: 3,
		}
		v := NewValidator(
			WithLimit(customLimit),
			WithSyntax(SyntaxPOSIX),
		)
		if v.limit.MaxRepeat != 200 {
			t.Errorf("MaxRepeat = %v; want 200", v.limit.MaxRepeat)
		}
		if v.limit.StarHeight != 3 {
			t.Errorf("StarHeight = %v; want 3", v.limit.StarHeight)
		}
		if v.syntax != syntax.POSIX {
			t.Errorf("syntax = %v; want %v", v.syntax, syntax.POSIX)
		}
	})
}

func TestValidator_Complexity(t *testing.T) {
	v := NewValidator()

	t.Run("empty_string", func(t *testing.T) {
		c, err := v.Complexity("")
		if err != nil {
			t.Fatalf("Complexity(\"\") error = %v; want nil", err)
		}
		if c.MaxRepeat != 0 {
			t.Errorf("MaxRepeat = %v; want 0", c.MaxRepeat)
		}
		if c.StarHeight != 0 {
			t.Errorf("StarHeight = %v; want 0", c.StarHeight)
		}
	})

	tests := []struct {
		name               string
		pattern            string
		expectedMaxRepeat  int
		expectedStarHeight int
		expectedMaxSize    int
	}{
		{
			name:               "simple_literal",
			pattern:            "hello",
			expectedMaxRepeat:  0,
			expectedStarHeight: 0,
			expectedMaxSize:    20, // "hello" = 5 runes * 4 bytes = 20
		},
		{
			name:               "simple_star",
			pattern:            "a*",
			expectedMaxRepeat:  0,
			expectedStarHeight: 1,
			expectedMaxSize:    4, // "a" = 1 rune * 4 bytes = 4
		},
		{
			name:               "simple_repeat",
			pattern:            "a{1,10}",
			expectedMaxRepeat:  10,
			expectedStarHeight: 0,
			expectedMaxSize:    4, // "a" = 1 rune * 4 bytes = 4
		},
		{
			name:               "email_pattern",
			pattern:            `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
			expectedMaxRepeat:  0,
			expectedStarHeight: 1,
			expectedMaxSize:    112, // character classes and literals
		},
		{
			name:               "complex_pattern",
			pattern:            `^[a-zA-Z0-9_]{3,20}$`,
			expectedMaxRepeat:  20,
			expectedStarHeight: 0,
			expectedMaxSize:    32, // character class with many runes
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c, err := v.Complexity(test.pattern)
			if err != nil {
				t.Fatalf("Complexity(%q) error = %v; want nil", test.pattern, err)
			}
			if c.MaxRepeat != test.expectedMaxRepeat {
				t.Errorf("MaxRepeat = %v; want %v", c.MaxRepeat, test.expectedMaxRepeat)
			}
			if c.StarHeight != test.expectedStarHeight {
				t.Errorf("StarHeight = %v; want %v", c.StarHeight, test.expectedStarHeight)
			}
		})
	}

	t.Run("invalid_regex", func(t *testing.T) {
		_, err := v.Complexity("(invalid")
		if err == nil {
			t.Error("Complexity(\"(invalid\") error = nil; want error")
		}
	})
}

func TestValidator_Validate(t *testing.T) {
	t.Run("valid_patterns", func(t *testing.T) {
		v := NewValidator()

		validPatterns := []string{
			"hello",
			"a*",
			"a{1,10}",
			`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
			`^[a-zA-Z0-9_]{3,20}$`,
		}

		for _, pattern := range validPatterns {
			t.Run(pattern, func(t *testing.T) {
				err := v.Validate(pattern)
				if err != nil {
					t.Errorf("Validate(%q) error = %v; want nil", pattern, err)
				}
			})
		}
	})

	t.Run("invalid_syntax", func(t *testing.T) {
		v := NewValidator()
		err := v.Validate("(invalid")
		if err == nil {
			t.Error("Validate(\"(invalid\") error = nil; want error")
		}
	})

	t.Run("exceeds_max_repeat", func(t *testing.T) {
		v := NewValidator(WithLimit(&Complexity{
			MaxRepeat:  10,
			StarHeight: 1,
		}))
		err := v.Validate("a{1,20}")
		if err == nil {
			t.Error("Validate(\"a{1,20}\") error = nil; want error")
		}
	})

	t.Run("exceeds_star_height", func(t *testing.T) {
		v := NewValidator(WithLimit(&Complexity{
			MaxRepeat:  50,
			StarHeight: 1,
		}))
		err := v.Validate("(a+)+")
		if err == nil {
			t.Error("Validate(\"(a+)+\") error = nil; want error")
		}
	})

	t.Run("exceeds_both_limits", func(t *testing.T) {
		v := NewValidator(WithLimit(&Complexity{
			MaxRepeat:  10,
			StarHeight: 1,
		}))
		err := v.Validate("(a{1,20})+")
		if err == nil {
			t.Error("Validate(\"(a{1,20})+\") error = nil; want error")
		}
	})
}

// TestReDoSPatterns tests typical ReDoS (Regular Expression Denial of Service) patterns
func TestReDoSPatterns(t *testing.T) {
	v := NewValidator()

	t.Run("nested_quantifiers", func(t *testing.T) {
		redosPatterns := []struct {
			name               string
			pattern            string
			expectedStarHeight int
			shouldFail         bool
		}{
			{
				name:               "classic_redos",
				pattern:            "(a+)+",
				expectedStarHeight: 2,
				shouldFail:         true,
			},
			{
				name:               "double_star",
				pattern:            "(a*)*",
				expectedStarHeight: 2,
				shouldFail:         true,
			},
			{
				name:               "plus_star_combination",
				pattern:            "(a+)*",
				expectedStarHeight: 2,
				shouldFail:         true,
			},
			{
				name:               "redos_with_suffix",
				pattern:            "(a+)+b",
				expectedStarHeight: 2,
				shouldFail:         true,
			},
			{
				name:               "triple_nesting",
				pattern:            "((a+)*)*",
				expectedStarHeight: 3,
				shouldFail:         true,
			},
		}

		for _, test := range redosPatterns {
			t.Run(test.name, func(t *testing.T) {
				c, err := v.Complexity(test.pattern)
				if err != nil {
					t.Fatalf("Complexity(%q) error = %v; want nil", test.pattern, err)
				}
				if c.StarHeight != test.expectedStarHeight {
					t.Errorf("StarHeight = %v; want %v", c.StarHeight, test.expectedStarHeight)
				}

				err = v.Validate(test.pattern)
				if test.shouldFail && err == nil {
					t.Errorf("Validate(%q) error = nil; want error", test.pattern)
				}
				if !test.shouldFail && err != nil {
					t.Errorf("Validate(%q) error = %v; want nil", test.pattern, err)
				}
			})
		}
	})

	t.Run("alternation_with_quantifiers", func(t *testing.T) {
		redosPatterns := []struct {
			name               string
			pattern            string
			expectedStarHeight int
			shouldFail         bool
		}{
			{
				name:               "alternation_with_quantifiers",
				pattern:            "(a+|b+)*",
				expectedStarHeight: 2,
				shouldFail:         true,
			},
			{
				name:               "duplicate_alternation",
				pattern:            "(a|a)*",
				expectedStarHeight: 1,
				shouldFail:         false,
			},
			{
				name:               "mixed_quantifiers",
				pattern:            "(a+|b*)*",
				expectedStarHeight: 2,
				shouldFail:         true,
			},
		}

		for _, test := range redosPatterns {
			t.Run(test.name, func(t *testing.T) {
				c, err := v.Complexity(test.pattern)
				if err != nil {
					t.Fatalf("Complexity(%q) error = %v; want nil", test.pattern, err)
				}
				if c.StarHeight != test.expectedStarHeight {
					t.Errorf("StarHeight = %v; want %v", c.StarHeight, test.expectedStarHeight)
				}

				err = v.Validate(test.pattern)
				if test.shouldFail && err == nil {
					t.Errorf("Validate(%q) error = nil; want error", test.pattern)
				}
				if !test.shouldFail && err != nil {
					t.Errorf("Validate(%q) error = %v; want nil", test.pattern, err)
				}
			})
		}
	})

	t.Run("real_world_dangerous_patterns", func(t *testing.T) {
		dangerousPatterns := []struct {
			name               string
			pattern            string
			expectedStarHeight int
			shouldFail         bool
		}{
			{
				name:               "whitespace_processing",
				pattern:            `\s*(\S+\s*)*`,
				expectedStarHeight: 2,
				shouldFail:         true,
			},
			{
				name:               "email_validation_dangerous",
				pattern:            `([a-zA-Z0-9._%+-]+)*@.*`,
				expectedStarHeight: 2,
				shouldFail:         true,
			},
			{
				name:               "safe_email_validation",
				pattern:            `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
				expectedStarHeight: 1,
				shouldFail:         false,
			},
		}

		for _, test := range dangerousPatterns {
			t.Run(test.name, func(t *testing.T) {
				c, err := v.Complexity(test.pattern)
				if err != nil {
					t.Fatalf("Complexity(%q) error = %v; want nil", test.pattern, err)
				}
				if c.StarHeight != test.expectedStarHeight {
					t.Errorf("StarHeight = %v; want %v", c.StarHeight, test.expectedStarHeight)
				}

				err = v.Validate(test.pattern)
				if test.shouldFail && err == nil {
					t.Errorf("Validate(%q) error = nil; want error", test.pattern)
				}
				if !test.shouldFail && err != nil {
					t.Errorf("Validate(%q) error = %v; want nil", test.pattern, err)
				}
			})
		}
	})
}

func TestBoundaryConditions(t *testing.T) {
	t.Run("star_height_boundary", func(t *testing.T) {
		v := NewValidator(WithLimit(&Complexity{
			MaxRepeat:  100,
			StarHeight: 1,
		}))

		// Should pass: StarHeight = 1
		err := v.Validate("a*")
		if err != nil {
			t.Errorf("Validate(\"a*\") error = %v; want nil", err)
		}

		// Should fail: StarHeight = 2
		err = v.Validate("(a*)*")
		if err == nil {
			t.Error("Validate(\"(a*)*\") error = nil; want error")
		}
	})

	t.Run("max_repeat_boundary", func(t *testing.T) {
		v := NewValidator(WithLimit(&Complexity{
			MaxRepeat:  50,
			StarHeight: 2,
		}))

		// Should pass: MaxRepeat = 50
		err := v.Validate("a{1,50}")
		if err != nil {
			t.Errorf("Validate(\"a{1,50}\") error = %v; want nil", err)
		}

		// Should fail: MaxRepeat = 51
		err = v.Validate("a{1,51}")
		if err == nil {
			t.Error("Validate(\"a{1,51}\") error = nil; want error")
		}
	})
}

func TestSyntaxDifferences(t *testing.T) {
	t.Run("perl_vs_posix", func(t *testing.T) {
		// Test a pattern that behaves differently in Perl vs POSIX
		perlValidator := NewValidator(WithSyntax(SyntaxPerl))
		posixValidator := NewValidator(WithSyntax(SyntaxPOSIX))

		// Both should handle basic patterns the same way
		pattern := "a+"
		perlC, perlErr := perlValidator.Complexity(pattern)
		posixC, posixErr := posixValidator.Complexity(pattern)

		if perlErr != nil || posixErr != nil {
			t.Fatalf("Complexity(%q) failed: perl=%v, posix=%v", pattern, perlErr, posixErr)
		}

		if perlC.StarHeight != posixC.StarHeight {
			t.Errorf("StarHeight differs: perl=%v, posix=%v", perlC.StarHeight, posixC.StarHeight)
		}
	})
}

func TestNegativeValueSkipping(t *testing.T) {
	t.Run("skip_max_repeat_only", func(t *testing.T) {
		v := NewValidator(WithLimit(&Complexity{
			MaxRepeat:  -1,
			StarHeight: 1,
		}))

		// Test that MaxRepeat calculation is skipped
		c, err := v.Complexity("a{1,100}")
		if err != nil {
			t.Fatalf("Complexity(\"a{1,100}\") error = %v; want nil", err)
		}
		if c.MaxRepeat != -1 {
			t.Errorf("MaxRepeat = %v; want -1", c.MaxRepeat)
		}
		if c.StarHeight != 0 {
			t.Errorf("StarHeight = %v; want 0", c.StarHeight)
		}

		// Test that validation skips MaxRepeat check
		err = v.Validate("a{1,999}")
		if err != nil {
			t.Errorf("Validate(\"a{1,999}\") error = %v; want nil (MaxRepeat check should be skipped)", err)
		}

		// Test that StarHeight check still works
		err = v.Validate("(a+)+")
		if err == nil {
			t.Error("Validate(\"(a+)+\") error = nil; want error (StarHeight check should work)")
		}
	})

	t.Run("skip_star_height_only", func(t *testing.T) {
		v := NewValidator(WithLimit(&Complexity{
			MaxRepeat:  10,
			StarHeight: -1,
		}))

		// Test that StarHeight calculation is skipped
		c, err := v.Complexity("(a+)+")
		if err != nil {
			t.Fatalf("Complexity(\"(a+)+\") error = %v; want nil", err)
		}
		if c.MaxRepeat != 0 {
			t.Errorf("MaxRepeat = %v; want 0", c.MaxRepeat)
		}
		if c.StarHeight != -1 {
			t.Errorf("StarHeight = %v; want -1", c.StarHeight)
		}

		// Test that validation skips StarHeight check
		err = v.Validate("((a+)*)*")
		if err != nil {
			t.Errorf("Validate(\"((a+)*)*\") error = %v; want nil (StarHeight check should be skipped)", err)
		}

		// Test that MaxRepeat check still works
		err = v.Validate("a{1,20}")
		if err == nil {
			t.Error("Validate(\"a{1,20}\") error = nil; want error (MaxRepeat check should work)")
		}
	})

	t.Run("skip_both", func(t *testing.T) {
		v := NewValidator(WithLimit(&Complexity{
			MaxRepeat:  -1,
			StarHeight: -1,
		}))

		// Test that both calculations are skipped
		c, err := v.Complexity("(a{1,100})+")
		if err != nil {
			t.Fatalf("Complexity(\"(a{1,100})+\") error = %v; want nil", err)
		}
		if c.MaxRepeat != -1 {
			t.Errorf("MaxRepeat = %v; want -1", c.MaxRepeat)
		}
		if c.StarHeight != -1 {
			t.Errorf("StarHeight = %v; want -1", c.StarHeight)
		}

		// Test that all validation is skipped
		err = v.Validate("((a{1,999})*)*")
		if err != nil {
			t.Errorf("Validate(\"((a{1,999})*)*\") error = %v; want nil (all checks should be skipped)", err)
		}
	})

	t.Run("zero_values_still_validate", func(t *testing.T) {
		v := NewValidator(WithLimit(&Complexity{
			MaxRepeat:  0,
			StarHeight: 0,
		}))

		// Zero should still validate (not be treated as negative)
		err := v.Validate("a+")
		if err == nil {
			t.Error("Validate(\"a+\") error = nil; want error (StarHeight = 1 > 0)")
		}

		err = v.Validate("a{1,1}")
		if err == nil {
			t.Error("Validate(\"a{1,1}\") error = nil; want error (MaxRepeat = 1 > 0)")
		}
	})
}
