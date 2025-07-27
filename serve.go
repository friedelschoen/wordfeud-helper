package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"
	"strings"
	"unicode"
)

var wordlist []string
var prologue, epilogue, indexpage []byte
var addr = "0.0.0.0:8080"

func loggingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s %s\n", r.RemoteAddr, r.Method, r.URL)
		next.ServeHTTP(w, r)
	})
}

func serve() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/search", handleSearch)
	mux.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir(*workdirFlag+"/static"))))
	mux.Handle("/static/fonts/", http.StripPrefix("/static/fonts", http.FileServer(http.Dir(*workdirFlag+"/go-image/font/gofont/ttfs"))))

	log.Printf("Listening on http://%s\n", addr)
	log.Fatal(http.ListenAndServe(addr, loggingHandler(mux)))
}

func cleanString(input string) string {
	return strings.Map(func(r rune) rune {
		r = unicode.ToLower(r)
		if unicode.IsLetter(r) || unicode.IsDigit(r) || strings.ContainsRune("%?&*", r) {
			return r
		}
		return -1
	}, input)
}

func writeEpilog(w io.Writer, letters, pattern string, checkPoints bool) {
	epi := string(epilogue)
	epi = strings.Replace(epi, "%letters%", letters, 1)
	epi = strings.Replace(epi, "%pattern%", pattern, 1)
	if checkPoints {
		epi = strings.Replace(epi, "%checked-letters%", "", 1)
		epi = strings.Replace(epi, "%checked-points%", "checked", 1)
	} else {
		epi = strings.Replace(epi, "%checked-letters%", "checked", 1)
		epi = strings.Replace(epi, "%checked-points%", "", 1)
	}
	w.Write([]byte(epi))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Write(prologue)
	w.Write(indexpage)
	writeEpilog(w, "", "", false)
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	letters := r.URL.Query().Get("letters")
	pattern := r.URL.Query().Get("pattern")
	sortPoints := r.URL.Query().Get("sort") == "points"
	letters = cleanString(letters)
	pattern = cleanString(pattern)

	if pattern == "" {
		pattern = "%"
	}
	scores, err := FindWords(wordlist, letters, pattern)
	slices.SortFunc(scores, func(a, b ScoredWord) int {
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
	})

	w.Write(prologue)

	fmt.Fprintf(w, "<h1>Wordfeud Helper</h1>\n")
	switch len(scores) {
	case 0:
		fmt.Fprintf(w, "<h2>Geen woorden gevonden</h2>\n")
	case 1:
		fmt.Fprintf(w, "<h2>&Eacute;&eacute;n woord gevonden</h2>\n")
	default:
		fmt.Fprintf(w, "<h2>%d woorden gevonden</h2>\n", len(scores))
	}
	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "<p><strong>Letters:</strong> <code>%s</code> - <strong>Patroon:</strong> <code>%s</code></p>\n", letters, pattern)
	if err != nil {
		fmt.Fprintf(w, "<h3 style='color: red;'>%v</h3>\n", err)
	}

	if len(scores) > 0 {
		overflow := 0
		if len(scores) > 100 {
			overflow = len(scores) - 100
			scores = scores[:100]
		}

		current := 0
		for _, word := range scores {
			if sortPoints {
				if word.Score != current {
					current = word.Score
					if current == 1 {
						fmt.Fprintf(w, "</span>\n<h3>1 Punt</h3>\n")
					} else {
						fmt.Fprintf(w, "</span>\n<h3>%d Punten</h3>\n", current)
					}
				} else {
					fmt.Fprintf(w, ", ")
				}
			} else {
				if len(word.Word) != current {
					current = len(word.Word)
					if current == 1 {
						fmt.Fprintf(w, "\n<h3>1 Letter</h3>\n")
					} else {
						fmt.Fprintf(w, "\n<h3>%d Letters</h3>\n", current)
					}
				} else {
					fmt.Fprintf(w, ", ")
				}
			}
			fmt.Fprintf(w, "<span class='mono'>\n")
			wasconsumed := false
			for i, chr := range word.Word {
				if word.Mods[i].Consumed != wasconsumed {
					wasconsumed = word.Mods[i].Consumed
					if wasconsumed {
						fmt.Fprintf(w, "<span style='color: red;'>")
					} else {
						fmt.Fprintf(w, "</span>")
					}
				}

				if word.Mods[i].WMul > 0 {
					fmt.Fprintf(w, "<span class='mul w'><span class='main'>%c</span><span class='sub'>%d</span></span>", chr, word.Mods[i].WMul)
				} else if word.Mods[i].LMul > 0 {
					fmt.Fprintf(w, "<span class='mul l'><span class='main'>%c</span><span class='sub'>%d</span></span>", chr, word.Mods[i].LMul)
				} else {
					fmt.Fprintf(w, "%c", chr)
				}
			}
			if wasconsumed {
				fmt.Fprintf(w, "</span>")
			}

			fmt.Fprintf(w, "</span>\n")
			fmt.Fprintf(w, "<sub>%d</sub>", word.Score)
		}
		fmt.Fprintln(w)
		if overflow > 0 {
			fmt.Fprintf(w, "<p><strong>en %d meer...</strong></p>\n", overflow)
		}
	} else {
		fmt.Fprintf(w, "<p><em>Geen woorden gevonden</em></p>\n")
	}
	fmt.Fprintf(w, "<p><a href='.'>&ldsh; Ga terug</a></p>\n")
	fmt.Fprintf(w, "<h2>Zoek opnieuw</h2>\n")
	writeEpilog(w, letters, pattern, sortPoints)
}
