package main

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
)

var NoMatch = errors.New("no match")

const Timeout = 2 * time.Second

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

func match(pattern, word string, w int, set LetterSet, lettermods []LetterMod, score WordScores) (LetterSet, []LetterMod, uint, error) {
	if score.Multiplier == 0 {
		score.Multiplier = 1
	}

	for len(pattern) > 0 {
		switch pattern[0] {
		case '%', '*':
			var bestset LetterSet
			var bestmods []LetterMod
			var besttotal uint
			isvalid := false

			for i := w; i <= len(word); i++ {
				localset := set
				localscore := score
				localmods := slices.Clone(lettermods)

				valid := true
				if pattern[0] == '%' {
					for j := w; j < i; j++ {
						pts, ok := localset.Consume(word[j])
						if !ok {
							valid = false
							break
						}
						localscore.Add(pts)
						localmods[j].Consumed = true
					}
				} else {
					for j := w; j < i; j++ {
						localscore.Add(LetterPoints.Get(word[j]))
					}
				}
				if !valid {
					break
				}

				if matchedset, mods, totalscore, err := match(pattern[1:], word, i, localset, localmods, localscore); err == nil {
					if !isvalid || totalscore > besttotal {
						bestset = matchedset
						bestmods = mods
						besttotal = totalscore
					}
					isvalid = true
				} else if err != NoMatch {
					return set, nil, 0, err
				}
			}

			if !isvalid {
				return set, nil, 0, NoMatch
			}
			return bestset, bestmods, besttotal, nil

		case '?':
			if w >= len(word) {
				return set, nil, 0, NoMatch
			}
			letterscore, ok := set.Consume(word[w])
			if !ok {
				return set, nil, 0, NoMatch
			}
			lettermods[w].Consumed = true
			score.Add(letterscore)
			pattern = pattern[1:]
			w++

		case '&':
			if w >= len(word) {
				return set, nil, 0, NoMatch
			}
			score.Add(LetterPoints.Get(word[w]))
			pattern = pattern[1:]
			w++

		case '1', '2', '3', '4', '5', '6', '7', '8', '9':
			if len(pattern) < 2 {
				return set, nil, 0, fmt.Errorf("expected W or L after digit, got end-of-pattern")
			}
			if w == 0 {
				return set, nil, 0, fmt.Errorf("unexpected multiplier at beginning of string")
			}
			mul := uint(pattern[0] - '0')
			switch pattern[1] {
			case 'w':
				score.Multiplier *= mul
				if lettermods[w-1].LMul > 0 {
					return set, nil, 0, fmt.Errorf("W-multiplier after L-multiplier")
				}
				lettermods[w-1].WMul = mul
			case 'l':
				score.LastLetter *= mul
				if lettermods[w-1].WMul > 0 {
					return set, nil, 0, fmt.Errorf("L-multiplier after W-multiplier")
				}
				lettermods[w-1].LMul = mul
			default:
				return set, nil, 0, fmt.Errorf("expected W or L after digit, got `%s`", pattern[:2])
			}
			pattern = pattern[2:]

		default:
			if w >= len(word) || pattern[0] != word[w] {
				return set, nil, 0, NoMatch
			}
			score.Add(LetterPoints.Get(word[w]))
			pattern = pattern[1:]
			w++
		}
	}

	if len(pattern) != 0 || w != len(word) {
		return set, nil, 0, NoMatch
	}
	return set, lettermods, score.Sum(), nil
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
		set, mods, score, err := match(pattern, w, 0, set, lettermods, WordScores{})
		if err == nil {
			result.Add(ScoredWord{set, mods, w, int(score)})
		} else if err != NoMatch {
			return nil, 0, err
		}
	}

	return result.Arr, max(0, result.N-len(result.Arr)), nil
}
