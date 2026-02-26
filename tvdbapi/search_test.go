package tvdbapi

import (
	"testing"
)

func TestDamerauLevenshtein(t *testing.T) {
	tests := []struct {
		name     string
		s1       string
		s2       string
		expected int
	}{
		// Empty strings
		{"empty s1", "", "abc", 3},
		{"empty s2", "abc", "", 3},
		{"both empty", "", "", 0},

		// Identical strings
		{"identical", "hello", "hello", 0},
		{"identical unicode", "héllo", "héllo", 0},

		// Insertions
		{"single insertion", "abc", "abcd", 1},
		{"single insertion at start", "bc", "abc", 1},
		{"single insertion at end", "abc", "dabc", 1},

		// Deletions
		{"single deletion", "abcd", "abc", 1},
		{"single deletion at start", "abc", "bc", 1},
		{"single deletion at end", "abc", "abd", 1},

		// Substitutions
		{"single substitution", "abc", "axc", 1},
		{"multiple substitutions", "abc", "xyz", 3},

		// Transpositions (adjacent)
		{"single transposition", "ab", "ba", 1},
		{"transposition in middle", "abcde", "abced", 1},
		{"double transposition", "ca", "ac", 1}, // OSA: adjacent swap counts as 1

		// Combined operations
		{"insertion and substitution", "abc", "axbc", 1},
		{"deletion and substitution", "axbc", "abc", 1},

		// Unicode
		{"unicode accent", "café", "cafe", 1},
		{"unicode cyrillic", "привет", "привет", 0},

		// Real-world TV show name cases (use NormalizeShowName first for case-insensitive)
		{"tv show the prefix", "The Office", "Office", 4},
		{"tv show articles", "The Office", "The Office", 0},
		{"tv show case sensitive", "the office", "The Office", 2}, // case difference counts as substitution
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DamerauLevenshtein(tt.s1, tt.s2)
			if result != tt.expected {
				t.Errorf("DamerauLevenshtein(%q, %q) = %d, want %d", tt.s1, tt.s2, result, tt.expected)
			}
		})
	}
}

func TestDamerauLevenshteinKnownValues(t *testing.T) {
	// These are known Damerau-Levenshtein distances from literature
	knownValues := []struct {
		s1       string
		s2       string
		expected int
	}{
		{"kitten", "sitting", 3},  // k→s, e→i, +g
		{"Saturday", "Sunday", 3}, // Satuday → Sunday
		{"flaw", "lawn", 2},       // OSA: f→l, +n, w (2 operations)
		{"algorithm", "altruistic", 6},
		{"CA", "ABC", 3}, // C→A, +B, A→C
		{"ABC", "CA", 3}, // A→C, B deleted, C→A
	}

	for _, tt := range knownValues {
		t.Run(tt.s1+"_"+tt.s2, func(t *testing.T) {
			result := DamerauLevenshtein(tt.s1, tt.s2)
			if result != tt.expected {
				t.Errorf("DamerauLevenshtein(%q, %q) = %d, want %d", tt.s1, tt.s2, result, tt.expected)
			}
		})
	}
}

// Benchmarks

func BenchmarkDamerauLevenshtein_ShortStrings(b *testing.B) {
	// Benchmark with short strings (typical TV show titles)
	s1, s2 := "The Office", "Office"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DamerauLevenshtein(s1, s2)
	}
}

func BenchmarkDamerauLevenshtein_MediumStrings(b *testing.B) {
	// Benchmark with medium strings
	s1, s2 := "Marvel's Agents of S.H.I.E.L.D.", "Agents of Shield"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DamerauLevenshtein(s1, s2)
	}
}

func BenchmarkDamerauLevenshtein_LongStrings(b *testing.B) {
	// Benchmark with longer strings
	s1 := "Breaking Bad"
	s2 := "Better Call Saul"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DamerauLevenshtein(s1, s2)
	}
}

func BenchmarkDamerauLevenshtein_IdenticalLongStrings(b *testing.B) {
	// Benchmark with identical long strings
	s := "The Quick Brown Fox Jumps Over The Lazy Dog"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DamerauLevenshtein(s, s)
	}
}

func BenchmarkDamerauLevenshtein_Unicode(b *testing.B) {
	// Benchmark with Unicode strings
	s1, s2 := "Stranger Things", "Stránger Thîngs"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DamerauLevenshtein(s1, s2)
	}
}

func BenchmarkDamerauLevenshtein_EmptyAndShort(b *testing.B) {
	// Benchmark with edge case: empty vs short string
	s1, s2 := "", "abc"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DamerauLevenshtein(s1, s2)
	}
}
