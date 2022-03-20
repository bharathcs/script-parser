package scriptparser

import (
	"regexp"
	"strings"
)

func StringSimplifierSkipIfRune(skipIfTrue func(r rune) bool) StringSimplifier {
	return func(s string) string {
		runes := []rune(s)
		var result []rune
		for _, r := range runes {
			if !skipIfTrue(r) {
				result = append(result, r)
			}
		}
		return string(result)
	}
}

// StringComparatorExactSearch
// `script.CreateSearchFunction(StringComparatorExactSearch)` gives you a SearchFunction that will look for a Line with
// an exact match with the target. Functions of the type StringSimplifier can be used to simplify strings before
// comparison.
var StringComparatorExactSearch = StringComparator(func(s1 string, s2 string) bool { return s1 == s2 })

// StringComparatorSubsetSearch
// `script.CreateSearchFunction(StringComparatorSubsetSearch)` gives you a SearchFunction that will look for a Line with
// a subset that matches the target. Functions of the type StringSimplifier can be used to simplify strings before
// comparison.
var StringComparatorSubsetSearch = StringComparator(strings.Contains)

// StringSimplifierAlphabetOnly is a StringSimplifier that extracts only alphabets from the string
var StringSimplifierAlphabetOnly = StringSimplifierSkipIfRune(func(r rune) bool {
	return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z'
})

// StringSimplifierIgnoreCase is a StringSimplifier that will have the comparator ignoring case
var StringSimplifierIgnoreCase = StringSimplifier(strings.ToUpper)

// CreateWildcardComparator creates a wildcard aware StringComparator. It takes in a wildcard regex pattern and treats
// any matches in the target string as wildcards. Note that it only looks to match the resulting target as a subset of
// dialogues, not an exact match.
func CreateWildcardComparator(wildcard regexp.Regexp) StringComparator {
	return func(possibleLine, target string) bool {
		positions := wildcard.FindAllStringIndex(target, -1)
		if len(positions) == 0 {
			return StringComparatorSubsetSearch(possibleLine, target)
		}

		var phrases []string
		for _, loc := range positions {
			if loc[1]-loc[0] > 1 {
				phrases = append(phrases, target[loc[0]:loc[1]])
			}
		}

		if len(phrases) == 0 {
			// As the target consists only of the wildcard, everything will return true.
			return true
		}

		for i, phrase := range phrases {
			if indexOfPhrase := strings.Index(possibleLine, phrase); indexOfPhrase < 0 {
				// No match
				return false
			} else if i == len(phrases)-1 {
				// Matched on the last phrase
				return true
			} else if len(possibleLine) <= len(phrase)+indexOfPhrase {
				// No more space left on the possibleLine AND more phrases left
				return false
			} else {
				// else trim until end of matched and continue
				possibleLine = possibleLine[indexOfPhrase+len(phrase)+1:]
			}
		}

		panic("This should not be reachable. Please raise a bug report! ref:'WildcardComparator'")
	}
}
