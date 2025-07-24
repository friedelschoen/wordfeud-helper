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
	list := make([]string, 0, strings.Count(string(content), "\n"))
	for line := range strings.SplitSeq(string(content), "\n") {
		if len(line) > 0 {
			list = append(list, line)
		}
	}
	return list
}

func main() {
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/search", handleSearch)
	// http.header.Add("Access-Control-Allow-Origin", "go.googlesource.com")
	http.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("./static"))))
	http.Handle("/static/fonts/", http.StripPrefix("/static/fonts", http.FileServer(http.Dir("./go-image/font/gofont/ttfs"))))

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
		Count    int
	}{
		Letters:  letters,
		Pattern:  pattern,
		Results:  scores,
		Overflow: overflow,
		Count:    len(scores),
	})

	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}
