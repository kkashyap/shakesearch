package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"index/suffixarray"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"os"
)

func main() {
	searcher := Searcher{}
	err := searcher.Load("completeworks.txt")
	if err != nil {
		log.Fatal(err)
	}

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/search", handleSearch(searcher))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	fmt.Printf("Listening on port %s...", port)
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		log.Fatal(err)
	}
}

var ListOfWorks = []string {"THE SONNETS", "ALL’S WELL THAT ENDS WELL", "THE TRAGEDY OF ANTONY AND CLEOPATRA", "AS YOU LIKE IT", "THE COMEDY OF ERRORS", "THE TRAGEDY OF CORIOLANUS", "CYMBELINE", 
								"THE TRAGEDY OF HAMLET, PRINCE OF DENMARK", "THE FIRST PART OF KING HENRY THE FOURTH", "THE SECOND PART OF KING HENRY THE FOURTH", "THE LIFE OF KING HENRY THE FIFTH", "THE FIRST PART OF HENRY THE SIXTH", "THE SECOND PART OF KING HENRY THE SIXTH", "THE THIRD PART OF KING HENRY THE SIXTH", "KING HENRY THE EIGHTH",
								"KING JOHN", "THE TRAGEDY OF JULIUS CAESAR", "THE TRAGEDY OF KING LEAR", "LOVE’S LABOUR’S LOST", "THE TRAGEDY OF MACBETH", "MEASURE FOR MEASURE", "MEASURE FOR MEASURE", "THE MERCHANT OF VENICE", "THE MERRY WIVES OF WINDSOR", "A MIDSUMMER NIGHT’S DREAM", "MUCH ADO ABOUT NOTHING", "THE TRAGEDY OF OTHELLO, MOOR OF VENICE",
								"PERICLES, PRINCE OF TYRE", "KING RICHARD THE SECOND", "KING RICHARD THE THIRD", "THE TRAGEDY OF ROMEO AND JULIET", "THE TAMING OF THE SHREW", "THE TEMPEST", "THE LIFE OF TIMON OF ATHENS", "THE TRAGEDY OF TITUS ANDRONICUS", "THE HISTORY OF TROILUS AND CRESSIDA", "TWELFTH NIGHT; OR, WHAT YOU WILL",
								"THE TWO GENTLEMEN OF VERONA", "THE TWO GENTLEMEN OF VERONA", "THE TWO NOBLE KINSMEN", "THE WINTER’S TALE", "A LOVER’S COMPLAINT", "THE PASSIONATE PILGRIM", "THE PHOENIX AND THE TURTLE", "THE RAPE OF LUCRECE", "VENUS AND ADONIS"}       

type Searcher struct {
	CompleteWorks string
	SuffixArray   *suffixarray.Index
	WorkIndicesMap map[string]*WorkIndices
}

type SearchResult struct {
	ResultString string
	WorkTitle string
}

type WorkIndices struct {
	StartIndex int
	EndIndex int
}

func handleSearch(searcher Searcher) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query, ok := r.URL.Query()["q"]
		if !ok || len(query[0]) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("missing search query in URL params"))
			return
		}
		results := searcher.Search(query[0])
		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		err := enc.Encode(results)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("encoding failure"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(buf.Bytes())
	}
}

func (s *Searcher) Load(filename string) error {         
	s.WorkIndicesMap = make(map[string]*WorkIndices)
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Load: %w", err)
	}
	s.CompleteWorks = string(dat)
	s.SuffixArray = suffixarray.New(dat)
	for index, work := range ListOfWorks {
		s.WorkIndicesMap[work] = &WorkIndices{}
		if index == 0 {
			s.WorkIndicesMap[work].StartIndex = 0
		} else {
			lookupResult := s.SuffixArray.Lookup([]byte(work), 1)
			if len(lookupResult) > 0 {
				currIndex := lookupResult[0]
				s.WorkIndicesMap[work].StartIndex = currIndex
				s.WorkIndicesMap[ListOfWorks[index - 1]].EndIndex = currIndex - 1
				if index == len(ListOfWorks) - 1 {
					s.WorkIndicesMap[work].EndIndex = len(s.CompleteWorks)
				}
			}
		}
	}
	return nil
}

func (s *Searcher) findWorkTitle(searchIndex int) string {
	for _, work := range ListOfWorks {
		if searchIndex >= s.WorkIndicesMap[work].StartIndex && searchIndex < s.WorkIndicesMap[work].EndIndex {
			return work
		}
	}
	return "Unidentified Source Material"
}

func (s *Searcher) Search(query string) []SearchResult {
	re := regexp.MustCompile("(?i)" + query)
	idxs := s.SuffixArray.FindAllIndex(re, -1)
	results := []SearchResult{}
	for _, idx := range idxs {
		result := SearchResult{
			ResultString: s.CompleteWorks[idx[0]-250:idx[0]+250],
			WorkTitle: s.findWorkTitle(idx[0]) }
		results = append(results, result)
	}
	return results
}
