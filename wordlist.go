package main

var LetterPoints = LetterCount{
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

func match(pattern, word string, set LetterSet) (bool, LetterSet) {
	p, w := 0, 0

	for p < len(pattern) {
		switch pattern[p] {
		case '%', '*':
			// Probeer alle mogelijke splits: 0 tot len(word)-w
			for i := w; i <= len(word); i++ {
				local := set
				if pattern[p] == '%' {
					valid := true
					for j := w; j < i; j++ {
						if !local.Consume(word[j]) {
							valid = false
							break
						}
					}
					if !valid {
						return false, set
					}
				}
				if ok, set := match(pattern[p+1:], word[i:], local); ok {
					return true, set
				}
			}
			return false, set

		case '?', '&':
			if w >= len(word) {
				return false, set
			}
			if pattern[p] == '?' && !set.Consume(word[w]) {
				return false, set
			}
			p++
			w++

		default:
			if w >= len(word) || pattern[p] != word[w] {
				return false, set
			}
			p++
			w++
		}
	}

	return p == len(pattern) && w == len(word), set
}

func WordScore(word string, jokers LetterCount) int {
	var sum uint
	for _, letter := range word {
		if !jokers.Decrement(byte(letter)) {
			sum += LetterPoints.Get(byte(letter))
		}
	}
	return int(sum)
}

type ScoredWord struct {
	LetterSet
	Word  string
	Score int
}

func FindWords(wordlist []string, letters, pattern string) []ScoredWord {
	set := NewLetterSet([]byte(letters))
	var result []ScoredWord
	for _, w := range wordlist {
		if ok, set := match(pattern, w, set); ok {
			result = append(result, ScoredWord{set, w, WordScore(w, set.JokerLetters)})
		}
	}

	return result
}
