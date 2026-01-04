package auth

import (
	"testing"
)

func TestGenerateBackupCodes(t *testing.T) {
	tests := []struct {
		name     string
		count    int
		expected int
	}{
		{"default count", 0, 10},
		{"negative count", -5, 10},
		{"custom count 5", 5, 5},
		{"custom count 8", 8, 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			codes, err := GenerateBackupCodes(tt.count)
			if err != nil {
				t.Fatalf("GenerateBackupCodes(%d) returned error: %v", tt.count, err)
			}

			if len(codes) != tt.expected {
				t.Errorf("expected %d codes, got %d", tt.expected, len(codes))
			}

			// Check format of each code (XXXX-XXXX)
			for i, code := range codes {
				if len(code) != 9 {
					t.Errorf("code %d has wrong length: %s (expected 9 chars)", i, code)
				}
				if code[4] != '-' {
					t.Errorf("code %d missing hyphen at position 4: %s", i, code)
				}
			}

			// Check uniqueness
			seen := make(map[string]bool)
			for _, code := range codes {
				if seen[code] {
					t.Errorf("duplicate code found: %s", code)
				}
				seen[code] = true
			}
		})
	}
}

func TestNormalizeBackupCode(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"ABCD-EFGH", "ABCDEFGH"},
		{"abcd-efgh", "ABCDEFGH"},
		{"AbCd-EfGh", "ABCDEFGH"},
		{"ABCDEFGH", "ABCDEFGH"},
		{"abcdefgh", "ABCDEFGH"},
		{"AB-CD-EF-GH", "ABCDEFGH"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NormalizeBackupCode(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeBackupCode(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidateBackupCode(t *testing.T) {
	storedCodes := []string{"ABCD-1234", "EFGH-5678", "IJKL-9012"}

	tests := []struct {
		name           string
		input          string
		expectedIdx    int
		expectedRemain int
	}{
		{"exact match first", "ABCD-1234", 0, 2},
		{"exact match middle", "EFGH-5678", 1, 2},
		{"exact match last", "IJKL-9012", 2, 2},
		{"lowercase match", "abcd-1234", 0, 2},
		{"no hyphen match", "ABCD1234", 0, 2},
		{"lowercase no hyphen", "efgh5678", 1, 2},
		{"not found", "XXXX-YYYY", -1, 3},
		{"empty input", "", -1, 3},
		{"partial match", "ABCD", -1, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy to avoid mutation across tests
			codesCopy := make([]string, len(storedCodes))
			copy(codesCopy, storedCodes)

			idx, remaining := ValidateBackupCode(tt.input, codesCopy)
			if idx != tt.expectedIdx {
				t.Errorf("ValidateBackupCode(%q) returned index %d, expected %d", tt.input, idx, tt.expectedIdx)
			}
			if len(remaining) != tt.expectedRemain {
				t.Errorf("ValidateBackupCode(%q) returned %d remaining codes, expected %d", tt.input, len(remaining), tt.expectedRemain)
			}

			// If matched, verify the code was removed
			if idx >= 0 {
				normalizedInput := NormalizeBackupCode(tt.input)
				for _, code := range remaining {
					if NormalizeBackupCode(code) == normalizedInput {
						t.Errorf("matched code %q still present in remaining codes", tt.input)
					}
				}
			}
		})
	}
}

func TestValidateBackupCodeEmptyList(t *testing.T) {
	idx, remaining := ValidateBackupCode("ABCD-1234", []string{})
	if idx != -1 {
		t.Errorf("expected -1 for empty list, got %d", idx)
	}
	if len(remaining) != 0 {
		t.Errorf("expected 0 remaining codes, got %d", len(remaining))
	}
}

func TestBackupCodesAreRandom(t *testing.T) {
	// Generate two batches and ensure they're different
	codes1, err := GenerateBackupCodes(10)
	if err != nil {
		t.Fatalf("GenerateBackupCodes returned error: %v", err)
	}

	codes2, err := GenerateBackupCodes(10)
	if err != nil {
		t.Fatalf("GenerateBackupCodes returned error: %v", err)
	}

	// Check that at least some codes are different
	sameCount := 0
	for i := range codes1 {
		if codes1[i] == codes2[i] {
			sameCount++
		}
	}

	// It's astronomically unlikely that all 10 codes would be the same
	if sameCount == 10 {
		t.Error("two batches of backup codes are identical - randomness issue")
	}
}
