package main

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

var NoMatch = errors.New("no match")

const Timeout = 5 * time.Second

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

type WordScores struct {
	Multiplier uint /* applied by 2W or 3W */
	Total      uint /* score of previous letters, excluding the last */
	LastLetter uint /* score of last letter, applied with 2L or 3L */
}

func (ws *WordScores) Sum() uint {
	return (ws.Total + ws.LastLetter) * ws.Multiplier
}

func (ws *WordScores) Add(score uint) {
	ws.Total += ws.LastLetter
	ws.LastLetter = score
}

type LetterMod struct {
	LMul     uint
	WMul     uint
	Consumed bool
}

func preparePattern(pattern string) (string, error) {
	if pattern == "" {
		pattern = "%"
	}

	res := make([]rune, 0, len(pattern))
	prev := utf8.RuneError
	for _, cur := range pattern {
		if !unicode.IsLetter(cur) && !strings.ContainsRune("%?&*123456789", cur) {
			return "", fmt.Errorf("Illegal character `%c`", cur)
		}
		if unicode.IsDigit(cur) && (prev == '%' || prev == '*') {
			return "", fmt.Errorf("Multiplier after nullable character")
		}
		if cur == '%' && (prev == '%' || prev == '*') {
			continue
		}
		if cur == '*' && (prev == '%' || prev == '*') {
			continue
		}
		prev = cur
		res = append(res, unicode.ToLower(cur))
	}
	return string(res), nil
}

func match(pattern, word string, set LetterSet, lettermods []LetterMod, score WordScores) (LetterSet, uint, error) {
	if score.Multiplier == 0 {
		score.Multiplier = 1
	}
	var prevmod *LetterMod
	for len(pattern) > 0 && len(word) > 0 {
		if strings.ContainsRune("%*?&123456789", rune(pattern[0])) {
			break
		}
		if pattern[0] != word[0] {
			return set, 0, NoMatch
		}
		pattern = pattern[1:]
		word = word[1:]
		prevmod = &lettermods[0]
		lettermods = lettermods[1:]
	}

	for len(pattern) > 0 {
		switch pattern[0] {
		case '%':
			if len(pattern) == 1 {
				for i, w := range word {
					sc, ok := set.Consume(byte(w))
					if !ok {
						return set, 0, NoMatch
					}
					lettermods[i].Consumed = true
					score.Add(sc)
				}
				return set, score.Sum(), nil
			}

			var bestset LetterSet
			var bestmods []LetterMod
			var besttotal uint
			isvalid := false

			for i := 0; i <= len(word); i++ {
				localset := set
				localscore := score
				localmods := lettermods
				if i != 0 {
					localmods = slices.Clone(lettermods)
				}

				valid := true
				for j := range i {
					pts, ok := localset.Consume(word[j])
					if !ok {
						valid = false
						break
					}
					localscore.Add(pts)
					localmods[j].Consumed = true
				}
				if !valid {
					break
				}

				if matchedset, totalscore, err := match(pattern[1:], word[i:], localset, localmods[i:], localscore); err == nil {
					if !isvalid || totalscore > besttotal {
						bestset = matchedset
						bestmods = localmods
						besttotal = totalscore
					}
					isvalid = true
				} else if err != NoMatch {
					return set, 0, err
				}
			}

			if !isvalid {
				return set, 0, NoMatch
			}
			copy(lettermods, bestmods)
			return bestset, besttotal, nil

		case '*':
			if len(pattern) == 1 {
				for _, w := range word {
					score.Add(LetterPoints.Get(byte(w)))
				}
				return set, score.Sum(), nil
			}

			var bestset LetterSet
			var bestmods []LetterMod
			var besttotal uint
			isvalid := false

			for i := 0; i <= len(word); i++ {
				localset := set
				localscore := score
				localmods := lettermods
				if i != 0 {
					localmods = slices.Clone(lettermods)
				}

				for j := range i {
					localscore.Add(LetterPoints.Get(word[j]))
				}

				if matchedset, totalscore, err := match(pattern[1:], word[i:], localset, localmods[i:], localscore); err == nil {
					if !isvalid || totalscore > besttotal {
						bestset = matchedset
						bestmods = localmods
						besttotal = totalscore
					}
					isvalid = true
				} else if err != NoMatch {
					return set, 0, err
				}
			}

			if !isvalid {
				return set, 0, NoMatch
			}
			copy(lettermods, bestmods)
			return bestset, besttotal, nil

		case '?':
			if len(word) == 0 {
				return set, 0, NoMatch
			}
			letterscore, ok := set.Consume(word[0])
			if !ok {
				return set, 0, NoMatch
			}
			lettermods[0].Consumed = true
			score.Add(letterscore)
			pattern = pattern[1:]
			word = word[1:]
			prevmod = &lettermods[0]
			lettermods = lettermods[1:]

		case '&':
			if len(word) == 0 {
				return set, 0, NoMatch
			}
			score.Add(LetterPoints.Get(word[0]))
			pattern = pattern[1:]
			word = word[1:]
			prevmod = &lettermods[0]
			lettermods = lettermods[1:]

		case '1', '2', '3', '4', '5', '6', '7', '8', '9':
			if len(pattern) < 2 {
				return set, 0, fmt.Errorf("expected W or L after digit, got end-of-pattern")
			}
			if prevmod == nil {
				return set, 0, fmt.Errorf("unexpected multiplier at beginning of string")
			}
			mul := uint(pattern[0] - '0')
			switch pattern[1] {
			case 'w':
				score.Multiplier *= mul
				if prevmod.LMul > 0 {
					return set, 0, fmt.Errorf("W-multiplier after L-multiplier")
				}
				prevmod.WMul = mul
			case 'l':
				score.LastLetter *= mul
				if prevmod.WMul > 0 {
					return set, 0, fmt.Errorf("L-multiplier after W-multiplier")
				}
				prevmod.LMul = mul
			default:
				return set, 0, fmt.Errorf("expected W or L after digit, got `%s`", pattern[:2])
			}
			pattern = pattern[2:]

		default:
			if len(word) == 0 || pattern[0] != word[0] {
				return set, 0, NoMatch
			}
			score.Add(LetterPoints.Get(word[0]))
			pattern = pattern[1:]
			word = word[1:]
			prevmod = &lettermods[0]
			lettermods = lettermods[1:]
		}
	}

	if len(pattern) != 0 || len(word) != 0 {
		return set, 0, NoMatch
	}
	return set, score.Sum(), nil
}

type ScoredWord struct {
	LetterSet
	Mods  []LetterMod
	Word  string
	Score int
}

func FindWords(wordlist []string, letters, pattern string, sortPoints bool) ([]ScoredWord, int, error) {
	set := NewLetterSet([]byte(letters))
	var result LimitedSortedSlice[ScoredWord]
	result.Limit = 100
	result.Compare = func(a, b ScoredWord) int {
		if sortPoints {
			if b.Score != a.Score {
				return b.Score - a.Score
			}
			if len(b.Word) != len(a.Word) {
				return len(b.Word) - len(a.Word)
			}
		} else {
			if len(b.Word) != len(a.Word) {
				return len(b.Word) - len(a.Word)
			}
			if b.Score != a.Score {
				return b.Score - a.Score
			}
		}
		return strings.Compare(a.Word, b.Word)
	}

	timer := time.NewTimer(Timeout)
	for _, w := range wordlist {
		select {
		case <-timer.C:
			return nil, 0, fmt.Errorf("timeout")
		default:
		}
		lettermods := make([]LetterMod, len(w))
		set, score, err := match(pattern, w, set, lettermods, WordScores{})
		if err == nil {
			result.Add(ScoredWord{set, lettermods, w, int(score)})
		} else if err != NoMatch {
			return nil, 0, err
		}
	}

	return result.Arr, max(0, result.N-len(result.Arr)), nil
}
