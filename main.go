package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
)

const WordAlignment = 64

var wordlist = getWordList()
var templates = template.Must(template.ParseGlob("templates/*.html"))

func getWordList() []string {
	content, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	return strings.Split(string(content), "\n")
}

func main() {
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/search", handleSearch)

	log.Println("Server luistert op http://localhost:8080")
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	letters := r.URL.Query().Get("letters")
	pattern := r.URL.Query().Get("pattern")
	letters = strings.ToLower(letters)
	pattern = strings.ToLower(pattern)

	if pattern == "" {
		pattern = "%"
	}

	scores := FindWords(wordlist, letters, pattern)
	overflow := 0
	if len(scores) > 100 {
		overflow = len(scores) - 100
		scores = scores[:100]
	}

	err := templates.ExecuteTemplate(w, "search.html", struct {
		Letters  string
		Pattern  string
		Results  []ScoredWord
		Overflow int
	}{
		Letters:  letters,
		Pattern:  pattern,
		Results:  scores,
		Overflow: overflow,
	})

	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}
