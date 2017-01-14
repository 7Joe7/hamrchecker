package db

import (
	"github.com/7joe7/hamrchecker/resources"
)

func GetSearches() []*resources.Search {
	return getSearches()
}

func Save(searches []*resources.Search) error {
	return save(searches)
}

func Load() []*resources.Search {
	return load()
}

func AddSearch(search *resources.Search) (int, error) {
	return addSearch(search)
}

func RemoveSearch(id int) error {
	return removeSearch(id)
}
