package utils

import (
	"regexp"
	"strings"

	"golang.org/x/text/unicode/norm"
)

// translitTable mirrors web/admin/src/lib/slug.ts so agent-generated slugs
// match what admins see in the admin panel slug input.
var translitTable = map[rune]string{
	'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d", 'е': "e", 'ё': "e",
	'ж': "zh", 'з': "z", 'и': "i", 'й': "y", 'к': "k", 'л': "l", 'м': "m",
	'н': "n", 'о': "o", 'п': "p", 'р': "r", 'с': "s", 'т': "t", 'у': "u",
	'ф': "f", 'х': "h", 'ц': "ts", 'ч': "ch", 'ш': "sh", 'щ': "shch",
	'ъ': "", 'ы': "y", 'ь': "", 'э': "e", 'ю': "yu", 'я': "ya",
	'і': "i", 'ї': "yi", 'є': "ye", 'ґ': "g",
}

var slugNonAlnum = regexp.MustCompile(`[^a-z0-9]+`)

// Slugify converts arbitrary text (titles in any language) into an URL slug:
// lowercase, Cyrillic transliterated, accents stripped, non-alphanumerics
// collapsed into dashes. Returns "" when nothing survives.
func Slugify(text string) string {
	text = strings.ToLower(text)
	var b strings.Builder
	for _, r := range text {
		if mapped, ok := translitTable[r]; ok {
			b.WriteString(mapped)
			continue
		}
		b.WriteRune(r)
	}
	// NFKD + drop combining marks (é -> e), mirroring the TS implementation.
	decomposed := norm.NFKD.String(b.String())
	var clean strings.Builder
	for _, r := range decomposed {
		if isCombiningMark(r) {
			continue
		}
		clean.WriteRune(r)
	}
	slug := slugNonAlnum.ReplaceAllString(clean.String(), "-")
	return strings.Trim(slug, "-")
}

// isCombiningMark reports whether r is a combining diacritical mark
// (U+0300–U+036F range used by NFKD decompositions).
func isCombiningMark(r rune) bool {
	return r >= 0x0300 && r <= 0x036F
}
