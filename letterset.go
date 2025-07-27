package main

import "iter"

type LetterCount [26]uint

func (cnt *LetterCount) Sum() (total uint) {
	for _, n := range cnt {
		total += n
	}
	return
}

func (cnt *LetterCount) Seq() iter.Seq2[byte, uint] {
	return func(yield func(byte, uint) bool) {
		for chr, n := range cnt {
			if !yield(byte(chr+'a'), n) {
				return
			}
		}
	}
}

func (cnt *LetterCount) Get(l byte) uint {
	if l < 'a' || l > 'z' {
		return 0
	}
	l -= 'a'

	return cnt[l]
}

func (cnt *LetterCount) Decrement(l byte) bool {
	if l < 'a' || l > 'z' {
		return false
	}
	l -= 'a'

	if cnt[l] > 0 {
		cnt[l]--
		return true
	}
	return false
}

func (cnt *LetterCount) Increment(l byte) {
	if l < 'a' || l > 'z' {
		return
	}
	l -= 'a'

	cnt[l]++
}

type LetterSet struct {
	Available LetterCount
	Jokers    int

	JokerLetters LetterCount
	Consumed     LetterCount
}

func NewLetterSet(letters []byte) LetterSet {
	var s LetterSet
	for _, letter := range letters {
		if letter == '?' {
			s.Jokers++
			continue
		}
		s.Available.Increment(letter)
	}
	return s
}

func (s *LetterSet) Consume(letter byte) (uint, bool) {
	if s.Available.Decrement(letter) {
		s.Consumed.Increment(letter)
		return LetterPoints.Get(letter), true
	}

	if s.Jokers > 0 {
		s.Jokers--
		s.JokerLetters.Increment(letter)
		s.Consumed.Increment(letter)
		return 0, true
	}
	return 0, false
}
