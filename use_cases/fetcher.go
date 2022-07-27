package use_cases

import (
	"strings"
	"sync"

	"GoConcurrency-Bootcamp-2022/models"
)

type api interface {
	FetchPokemon(id int) (models.Pokemon, error)
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

func (f Fetcher) FetchV2(from, to int, doneChan chan bool) <-chan models.Pokemon {
	availableWorkers := to - from + 1
	resChan := make(chan models.Pokemon, availableWorkers)

	pokChan := make(chan models.Pokemon, availableWorkers)

	runningStateWorkers := 100
	runningStateWorkersChan := make(chan int, runningStateWorkers)

	for id := from; id <= availableWorkers; id++ {
		go func(index int) {
			runningStateWorkersChan <- index
			{
				pokemon, err := f.api.FetchPokemon(index)
				if err != nil {
					doneChan <- true
					close(resChan)
					return
				}
				resChan <- pokemon

			}
			<-runningStateWorkersChan
		}(id)
	}
	go func() {
		defer close(pokChan)
		for availableWorkers > 0 {
			pokemon := <-resChan
			var flatAbilities []string
			for _, t := range pokemon.Abilities {
				flatAbilities = append(flatAbilities, t.Ability.URL)
			}
			pokemon.FlatAbilityURLs = strings.Join(flatAbilities, "|")
			pokChan <- pokemon
			availableWorkers--
		}
	}()

	return pokChan
}

func (f Fetcher) WriteToCSV(res <-chan models.Pokemon, doneChan chan bool) <-chan error {
	var pokemons []models.Pokemon
	errChan := make(chan error)
	go func() {
		defer close(errChan)
		for v := range res {
			select {
			case <-doneChan:
				return
			default:
				pokemons = append(pokemons, v)
			}
		}
		errChan <- f.storage.Write(pokemons)
	}()
	return errChan
}
