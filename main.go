package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	//	LOC_BRANIK = 171
	//	SPORT_BADMINTON = 140
	TIME_FORMAT    = "2006-01-02"
	SEARCHES_STORE = "searches.json"
)

var (
	requestParams                        []string
	searchesMutex                        sync.Mutex
	searches                             []*search
)

func saveSearchesToFile() error {
	searchesB, err := json.Marshal(&searches)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(SEARCHES_STORE, searchesB, 777)
}

func loadSearches() {
	searchesB, err := ioutil.ReadFile(SEARCHES_STORE)
	if err != nil {
		// consider searches file non-existing
		searches = []*search{}
	}
	if err := json.Unmarshal(searchesB, &searches); err != nil {
		// searches file is corrupted
		log.Printf("Searches store is corrupted. Backing it up to %s.old.\n", SEARCHES_STORE)
		backupFileName := fmt.Sprintf("%s.old", SEARCHES_STORE)
		os.Remove(backupFileName)
		if err := ioutil.WriteFile(backupFileName, searchesB, 777); err != nil {
			log.Printf("Unable to backup corrupted searches store.\n")
		}
		searches = []*search{}
	}
}

func addSearch(s *search) {
	searchesMutex.Lock()
	defer searchesMutex.Unlock()
	now := time.Now()
	s.Start = &now
	searches = append(searches, s)
	if err := saveSearchesToFile(); err != nil {
		log.Printf("Unable to save searches to store. %v", err)
	}
}

func removeSearch(s *search) {
	searchesMutex.Lock()
	defer searchesMutex.Unlock()
	for i := 0; i < len(searches); i++ {
		if searches[i] == s {
			searches = append(searches[:i], searches[i+1:]...)
			break
		}
	}
	if err := saveSearchesToFile(); err != nil {
		log.Printf("Unable to save searches to store. %v", err)
	}
}

func removeSearchByIndex(i int) {
	searchesMutex.Lock()
	defer searchesMutex.Unlock()
	searches = append(searches[:i], searches[i+1:]...)
	if err := saveSearchesToFile(); err != nil {
		log.Printf("Unable to save searches to store. %v", err)
	}
}

func main() {
	http.HandleFunc("/", index)
	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("resources"))))
	server := &http.Server{Addr: "0.0.0.0:8080"}
	loadSearches()
	for i := 0; i < len(searches); i++ {
		if time.Now().Before(*searches[i].Till) {
			removeSearchByIndex(i)
		} else {
			go runSearch(searches[i])
		}
	}
	server.ListenAndServe()
}
