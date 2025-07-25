package main

type LetterSet struct {
	Available    [26]uint8
	Jokers       int
	JokerLetters [26]uint8
	Consumed     [26]uint8
}

func NewLetterSet(letters []byte) LetterSet {
	var s LetterSet
	for _, letter := range letters {
		if letter == '?' {
			s.Jokers++
			continue
		}
		l := letter - 'a'
		if l < 0 || l >= 26 {
			continue
		}
		s.Available[l]++
	}
	return s
}

func (s *LetterSet) Consume(letter byte) bool {
	l := letter - 'a'
	if l < 0 || l >= 26 {
		return false
	}
	if s.Available[l] > 0 {
		s.Available[l]--
		s.Consumed[l]++
		return true
	}

	if s.Jokers > 0 {
		s.Jokers--
		s.JokerLetters[l]++
		s.Consumed[l]++
		return true
	}
	return false
}
