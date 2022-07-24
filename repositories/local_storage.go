package repositories

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"GoConcurrency-Bootcamp-2022/models"
)

type LocalStorage struct{}

const filePath = "resources/pokemons.csv"

func (l LocalStorage) Write(pokemons []models.Pokemon) error {
	file, fErr := os.Create(filePath)
	defer file.Close()
	if fErr != nil {
		return fErr
	}

	w := csv.NewWriter(file)
	records := buildRecords(pokemons)
	if err := w.WriteAll(records); err != nil {
		return err
	}

	return nil
}

func (l LocalStorage) ReadByLine() <-chan models.Pokemon {
	pokChan := make(chan models.Pokemon, 8)
	go func() {
		defer close(pokChan)
		file, ferr := os.Open(filePath)
		if ferr != nil {
			log.Fatal(ferr)
		}
		defer file.Close()

		csvReader := csv.NewReader(file)
		line := 0
		for {
			start := time.Now()
			record, err := csvReader.Read()
			after := time.Since(start)
			fmt.Printf("csvReader.Read takes: %v", after)
			if line == 0 {
				line++
				continue
			}
			if err == io.EOF {
				return
			}
			if err != nil {
				return
			}
			pokemon, err := parseCSVRecord(record)
			if err != nil {
				return
			}
			// here we save the batch to proccess
			pokChan <- *pokemon
		}

	}()
	return pokChan
}

func (l LocalStorage) Read() ([]models.Pokemon, error) {
	file, fErr := os.Open(filePath)
	defer file.Close()
	if fErr != nil {
		return nil, fErr
	}

	r := csv.NewReader(file)
	records, rErr := r.ReadAll()
	if rErr != nil {
		return nil, rErr
	}

	pokemons, err := parseCSVData(records)
	if err != nil {
		return nil, err
	}

	return pokemons, nil
}

func buildRecords(pokemons []models.Pokemon) [][]string {
	headers := []string{"id", "name", "height", "weight", "flat_abilities"}
	records := [][]string{headers}
	for _, p := range pokemons {
		record := fmt.Sprintf("%d,%s,%d,%d,%s",
			p.ID,
			p.Name,
			p.Height,
			p.Weight,
			p.FlatAbilityURLs)
		records = append(records, strings.Split(record, ","))
	}

	return records
}

func parseCSVRecord(record []string) (*models.Pokemon, error) {
	var pokemon models.Pokemon

	id, err := strconv.Atoi(record[0])
	if err != nil {
		return nil, err
	}

	height, err := strconv.Atoi(record[2])
	if err != nil {
		return nil, err
	}

	weight, err := strconv.Atoi(record[3])
	if err != nil {
		return nil, err
	}

	pokemon = models.Pokemon{
		ID:              id,
		Name:            record[1],
		Height:          height,
		Weight:          weight,
		Abilities:       nil,
		FlatAbilityURLs: record[4],
		EffectEntries:   nil,
	}

	return &pokemon, nil
}

func parseCSVData(records [][]string) ([]models.Pokemon, error) {
	var pokemons []models.Pokemon
	for i, record := range records {
		if i == 0 {
			continue
		}

		id, err := strconv.Atoi(record[0])
		if err != nil {
			return nil, err
		}

		height, err := strconv.Atoi(record[2])
		if err != nil {
			return nil, err
		}

		weight, err := strconv.Atoi(record[3])
		if err != nil {
			return nil, err
		}

		pokemon := models.Pokemon{
			ID:              id,
			Name:            record[1],
			Height:          height,
			Weight:          weight,
			Abilities:       nil,
			FlatAbilityURLs: record[4],
			EffectEntries:   nil,
		}
		pokemons = append(pokemons, pokemon)
	}

	return pokemons, nil
}
