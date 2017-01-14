package db

import (
	"log"
	"io/ioutil"
	"os"
	"fmt"
	"encoding/json"
	"sync"
	"github.com/7joe7/hamrchecker/resources"
)

var (
	db sync.Mutex
	searches []*resources.Search
)

func getSearches() []*resources.Search {
	if searches == nil {
		searches = load()
	}
	return searches
}

func save(searches []*resources.Search) error {
	searchesB, err := json.Marshal(&searches)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(resources.SEARCHES_STORE, searchesB, 777)
}

func load() []*resources.Search {
	searches := []*resources.Search{}
	searchesB, err := ioutil.ReadFile(resources.SEARCHES_STORE)
	if err != nil {
		log.Printf("Error reading searches store. %v", err)
		return searches
	}
	if err := json.Unmarshal(searchesB, &searches); err != nil {
		log.Printf("Searches store is corrupted. Backing it up to %s.old.\n", resources.SEARCHES_STORE)
		backupFileName := fmt.Sprintf("%s.old", resources.SEARCHES_STORE)
		os.Remove(backupFileName)
		if err := ioutil.WriteFile(backupFileName, searchesB, 777); err != nil {
			log.Printf("Unable to backup corrupted searches store.\n")
		}
	}
	return searches
}

func loadEmailConf() (*resources.EmailConf, error) {
	emailConf := &resources.EmailConf{}
	emailConfContent, err := ioutil.ReadFile(resources.EMAIL_CONF_STORE)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(emailConfContent, emailConf)
	if err != nil {
		return nil, err
	}
	return emailConf, nil
}

func addSearch(search *resources.Search) (int, error) {
	db.Lock()
	defer db.Unlock()
	searches = append(searches, search)
	if search.Id == 0 {
		search.Id = getNewId(searches)
	}
	if err := save(searches); err != nil {
		return search.Id, err
	}
	return search.Id, nil
}

func removeSearch(id int) error {
	db.Lock()
	defer db.Unlock()
	for i := 0; i < len(searches); i++ {
		if searches[i].Id == id {
			searches = append(searches[:i], searches[i+1:]...)
			break
		}
	}
	if err := save(searches); err != nil {
		return err
	}
	return nil
}

func getNewId(searches []*resources.Search) int {
	var h int
	for i := 0; i < len(searches); i++ {
		if h < searches[i].Id {
			h = searches[i].Id
		}
	}
	return h+1
}
