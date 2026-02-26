package tvdbapi

import (
	"regexp"
	"strings"
)

// stopWords are common words to remove from TV show titles
var stopWords = map[string]bool{
	"the": true, "a": true, "an": true, "and": true, "or": true,
	"but": true, "of": true, "in": true, "on": true, "at": true,
	"to": true, "for": true, "with": true, "by": true, "from": true,
	"as": true, "is": true, "was": true, "are": true, "were": true,
	"be": true, "been": true, "being": true, "has": true, "have": true,
	"had": true, "do": true, "does": true, "did": true, "will": true,
	"would": true, "could": true, "should": true, "may": true, "might": true,
	"must": true, "shall": true, "can": true, "need": true, "dare": true,
	"ought": true, "used": true,
}

// accentMap maps accented characters to their ASCII equivalents
var accentMap = map[rune]rune{
	'à': 'a', 'á': 'a', 'â': 'a', 'ã': 'a', 'ä': 'a', 'å': 'a', 'ā': 'a', 'ą': 'a', 'ă': 'a',
	'è': 'e', 'é': 'e', 'ê': 'e', 'ë': 'e', 'ē': 'e', 'ę': 'e', 'ě': 'e', 'ė': 'e',
	'ì': 'i', 'í': 'i', 'î': 'i', 'ï': 'i', 'ī': 'i', 'į': 'i',
	'ò': 'o', 'ó': 'o', 'ô': 'o', 'õ': 'o', 'ö': 'o', 'ø': 'o', 'ō': 'o', 'ő': 'o',
	'ù': 'u', 'ú': 'u', 'û': 'u', 'ü': 'u', 'ū': 'u', 'ű': 'u', 'ů': 'u',
	'ñ': 'n', 'ń': 'n',
	'ç': 'c', 'ć': 'c', 'č': 'c',
	'ş': 's', 'ś': 's', 'š': 's',
	'ž': 'z', 'ź': 'z', 'ż': 'z',
	'ý': 'y', 'ÿ': 'y',
	'ď': 'd', 'đ': 'd',
	'ł': 'l',
	'ß': 's',
}

func removeSubtitle(name string) string {
	idx := strings.Index(name, ":")
	if idx != -1 {
		name = name[:idx]
	}

	return name
}

// NormalizeShowName normalizes a TV show name for comparison/searching
func NormalizeShowName(name string) string {
	name = strings.ToLower(name)
	name = convertAccents(name)

	regex := regexp.MustCompile(`[^a-z0-9\s]`)
	name = regex.ReplaceAllString(name, "")

	name = strings.Join(strings.Fields(name), " ")

	name = removeStopWords(name)

	return name
}

// convertAccents converts accented characters to their ASCII equivalents
func convertAccents(s string) string {
	return strings.Map(func(r rune) rune {
		replacement, found := accentMap[r]
		if found {
			return replacement
		}
		return r
	}, s)
}

// removeStopWords removes common stop words from the title
func removeStopWords(s string) string {
	words := strings.Fields(s)
	filtered := make([]string, 0, len(words))

	for _, word := range words {
		// Skip stop words but keep words shorter than 3 chars if they're not stop words
		// This preserves Roman numerals like "II" and "III"
		if !stopWords[word] {
			filtered = append(filtered, word)
		}
	}

	return strings.Join(filtered, " ")
}

// DamerauLevenshtein calculates the Damerau-Levenshtein distance between two strings.
// Uses an optimized rolling array approach with two 1D slices and tracks the previous
// diagonal value for correct transposition handling, reducing memory from O(m*n) to O(n).
func DamerauLevenshtein(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	r1 := []rune(s1)
	r2 := []rune(s2)

	// Ensure s1 is the shorter string for memory optimization
	if len(r1) > len(r2) {
		r1, r2 = r2, r1
	}

	rows := len(r1) + 1
	cols := len(r2) + 1

	// Use two 1D slices instead of 2D matrix (rolling array)
	prevRow := make([]int, cols)
	currRow := make([]int, cols)

	// Initialize first row
	for j := 0; j < cols; j++ {
		prevRow[j] = j
	}

	// Track the value from two rows back for transposition
	// prevPrevRow[j] stores the value from row i-2 at column j
	prevPrevRow := make([]int, cols)
	// Initialize prevPrevRow for row 0 (used when i=1 for transpositions)
	// prevPrevRow is all zeros since there is no row -1

	// Fill the matrix row by row
	for i := 1; i < rows; i++ {
		currRow[0] = i // First column

		for j := 1; j < cols; j++ {
			cost := 0
			if r1[i-1] != r2[j-1] {
				cost = 1
			}

			delCost := prevRow[j] + 1      // deletion
			insCost := currRow[j-1] + 1    // insertion
			subCost := prevRow[j-1] + cost // substitution

			// Minimum of deletion, insertion, substitution
			minVal := delCost
			if insCost < minVal {
				minVal = insCost
			}
			if subCost < minVal {
				minVal = subCost
			}
			currRow[j] = minVal

			// Transposition: if i>1 && j>1 && s1[i-1]==s2[j-2] && s1[i-2]==s2[j-1]
			if i > 1 && j > 1 && r1[i-1] == r2[j-2] && r1[i-2] == r2[j-1] {
				// Need prevPrevRow[j-2] which is from row i-2, column j-2
				transCost := prevPrevRow[j-2] + cost
				if transCost < currRow[j] {
					currRow[j] = transCost
				}
			}
		}

		copy(prevPrevRow, prevRow)
		prevRow, currRow = currRow, prevRow
	}

	return prevRow[cols-1]
}
