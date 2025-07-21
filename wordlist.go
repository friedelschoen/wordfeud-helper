package main

import (
	"slices"
)

var LetterPoints = [...]int{
	1,  // A
	4,  // B
	5,  // C
	2,  // D
	1,  // E
	4,  // F
	3,  // G
	4,  // H
	2,  // I
	4,  // J
	3,  // K
	3,  // L
	3,  // M
	1,  // N
	1,  // O
	4,  // P
	10, // Q
	2,  // R
	2,  // S
	2,  // T
	2,  // U
	4,  // V
	5,  // W
	8,  // X
	8,  // Y
	5,  // Z
}

func match(pattern, word string, set LetterSet) (bool, [26]uint8) {
	p, w := 0, 0

	for p < len(pattern) {
		switch pattern[p] {
		case '%', '*':
			// Probeer alle mogelijke splits: 0 tot len(word)-w
			for i := w; i <= len(word); i++ {
				local := set // kopie van LetterSet
				if pattern[p] == '%' {
					valid := true
					for j := w; j < i; j++ {
						if !local.Consume(word[j]) {
							valid = false
							break
						}
					}
					if !valid {
						return false, set.JokerLetters
					}
				}
				if ok, jokers := match(pattern[p+1:], word[i:], local); ok {
					return true, jokers
				}
			}
			return false, set.JokerLetters

		case '?', '&':
			if w >= len(word) {
				return false, set.JokerLetters
			}
			if pattern[p] == '?' && !set.Consume(word[w]) {
				return false, set.JokerLetters
			}
			p++
			w++

		default:
			if w >= len(word) || pattern[p] != word[w] {
				return false, set.JokerLetters
			}
			p++
			w++
		}
	}

	return p == len(pattern) && w == len(word), set.JokerLetters
}

func WordScore(word string, jokers [26]uint8) int {
	var sum int
	for _, letter := range word {
		l := letter - 'a'
		if l < 0 || l >= 26 {
			continue
		}
		if jokers[l] > 0 {
			jokers[l]--
			continue
		}
		sum += LetterPoints[l]
	}
	return sum
}

type ScoredWord struct {
	Word  string
	Score int
}

func FindWords(wordlist []string, letters, pattern string) []ScoredWord {
	set := NewLetterSet([]byte(letters))
	var result []ScoredWord
	for _, w := range wordlist {
		if ok, jokers := match(pattern, w, set); ok {
			result = append(result, ScoredWord{w, WordScore(w, jokers)})
		}
	}
	slices.SortFunc(result, func(a, b ScoredWord) int {
		if b.Score != a.Score {
			return b.Score - a.Score
		}
		return len(a.Word) - len(b.Word)
	})

	return result
}
