package use_cases

import (
	"strings"
	"sync"

	"GoConcurrency-Bootcamp-2022/models"
)

type api interface {
	FetchPokemon(id int, wg *sync.WaitGroup) (models.Pokemon, error)
}

type writer interface {
	Write(pokemons []models.Pokemon) error
}

type Fetcher struct {
	api     api
	storage writer
	wg      sync.WaitGroup
}

func NewFetcher(api api, storage writer) Fetcher {
	return Fetcher{api, storage, sync.WaitGroup{}}
}

type FetchResponse struct {
	Err     error
	Pokemon models.Pokemon
}

func (f Fetcher) Fetch(from, to int) chan FetchResponse {
	fetchResChan := make(chan FetchResponse)
	nPokemons := to - from + 1
	f.wg.Add(nPokemons)
	for id := from; id <= to; id++ {
		go func(id int) {
			pokemon, err := f.api.FetchPokemon(id, &f.wg)
			if err != nil {
				fetchResChan <- FetchResponse{Err: err}
			} else {
				var flatAbilities []string
				for _, t := range pokemon.Abilities {
					flatAbilities = append(flatAbilities, t.Ability.URL)
				}
				pokemon.FlatAbilityURLs = strings.Join(flatAbilities, "|")
				fetchResChan <- FetchResponse{Pokemon: pokemon}
			}
		}(id)
	}
	go func() {
		f.wg.Wait()
		close(fetchResChan)
	}()
	return fetchResChan
}

func (f Fetcher) WriteToCSV(res chan FetchResponse) <-chan error {
	var pokemons []models.Pokemon
	errChan := make(chan error)
	go func() {
		defer close(errChan)
		for v := range res {
			if v.Err != nil {
				errChan <- v.Err
			}
			pokemons = append(pokemons, v.Pokemon)
		}
		errChan <- f.storage.Write(pokemons)
	}()
	return errChan
}
