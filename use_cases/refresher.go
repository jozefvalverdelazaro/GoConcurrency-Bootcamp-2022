package use_cases

import (
	"context"
	"strings"

	"GoConcurrency-Bootcamp-2022/models"
)

type reader interface {
	Read() ([]models.Pokemon, error)
	ReadByLine() <-chan models.Pokemon
}

type saver interface {
	Save(context.Context, []models.Pokemon) error
}

type fetcher interface {
	FetchAbility(string) (models.Ability, error)
}

type Refresher struct {
	reader
	saver
	fetcher
}

func NewRefresher(reader reader, saver saver, fetcher fetcher) Refresher {
	return Refresher{reader, saver, fetcher}
}

func (r Refresher) RefreshV2(ctx context.Context) <-chan error {
	errChan := make(chan error)
	var pokemons []models.Pokemon
	go func() {
		defer close(errChan)
		pokeChan := r.ReadByLine()
		for p := range pokeChan {
			urls := strings.Split(p.FlatAbilityURLs, "|")
			p.EffectEntries = r.FanOutFetchAbilities(urls)
			pokemons = append(pokemons, p)

		}
		if err := r.Save(ctx, pokemons); err != nil {
			errChan <- err
			return
		}

	}()

	return errChan
}

func (r Refresher) FanOutFetchAbilities(urls []string) []string {
	availableWorkers := len(urls)
	resChan := make(chan models.Ability, availableWorkers)

	var abilities []string

	runningStateWorkers := 3
	runningStateWorkersChan := make(chan int, runningStateWorkers)

	for i := 0; i < availableWorkers; i++ {
		go func(index int) {
			runningStateWorkersChan <- index
			{
				ability, _ := r.FetchAbility(urls[index])
				resChan <- ability
			}
			<-runningStateWorkersChan
		}(i)
	}

	for availableWorkers > 0 {
		ability := <-resChan
		for _, ee := range ability.EffectEntries {
			abilities = append(abilities, ee.Effect)
		}
		availableWorkers--
	}
	return abilities
}

func (r Refresher) Refresh(ctx context.Context) error {
	pokemons, err := r.Read()
	if err != nil {
		return err
	}

	for i, p := range pokemons {
		urls := strings.Split(p.FlatAbilityURLs, "|")
		var abilities []string
		for _, url := range urls {
			ability, err := r.FetchAbility(url)
			if err != nil {
				return err
			}

			for _, ee := range ability.EffectEntries {
				abilities = append(abilities, ee.Effect)
			}
		}

		pokemons[i].EffectEntries = abilities
	}

	if err := r.Save(ctx, pokemons); err != nil {
		return err
	}

	return nil
}
