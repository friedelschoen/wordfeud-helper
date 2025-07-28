package main

import (
	"os"
	"strings"
	"testing"
)

var benchWordList []string

func init() {
	data, err := os.ReadFile("wordlist.txt")
	if err != nil {
		panic(err)
	}
	benchWordList = strings.Fields(string(data))
}

var benchPatterns = []struct {
	name    string
	letters string
	pattern string
}{
	{"wildcard1", "abcdeorstuvn?", "v%r?n"},
	{"wildcard2", "aeilnorstuv", "s?l?u%"},
	{"bonus2w3l", "etrinoasdl?", "2w?3l"},
	{"greedyall", "abcdefghijklmnopqrstuvwxyz??", "%"},
	{"vowels", "aaeeiioouu", "m%u?s?"},
	{"stars", "", "***"},
	{"singlestar", "", "*"},
	{"bonus-heavy", "aaaaaaabbccdddddeeeeeeeeeeeeeeeeeeffggghhiiiijjkkklllmmmnnnnnnnnnnnooooooppqrrrrrssssstttttuuuvvwwxyzz??", "%?3l%?3l%"},
}

func BenchmarkMatchFunction(b *testing.B) {
	for _, entry := range benchPatterns {
		b.Run(entry.name, func(b *testing.B) {
			pattern, err := preparePattern(entry.pattern)
			if err != nil {
				b.Fatalf("pattern error: %v", err)
			}

			b.ResetTimer()
			for b.Loop() {
				set := NewLetterSet([]byte(entry.letters))
				for _, word := range benchWordList {
					lettermods := make([]LetterMod, len(word))
					match(pattern, word, set, lettermods, WordScores{})
				}
			}
		})
	}
}
