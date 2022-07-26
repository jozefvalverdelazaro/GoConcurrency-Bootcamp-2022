package controllers

import (
	"context"
	"fmt"
	"net/http"

	"GoConcurrency-Bootcamp-2022/models"

	"github.com/gin-gonic/gin"
)

type API struct {
	fetcher
	refresher
	getter
}

func NewAPI(fetcher fetcher, refresher refresher, getter getter) API {
	return API{fetcher, refresher, getter}
}

type fetcher interface {
	FetchV2(from, to int, doneChan chan bool) <-chan models.Pokemon
	WriteToCSV(res <-chan models.Pokemon, doneChan chan bool) <-chan error
}

type refresher interface {
	Refresh(context.Context) error
	RefreshV2(context.Context) <-chan error
}

type getter interface {
	GetPokemons(context.Context) ([]models.Pokemon, error)
}

//FillCSV fill the local CSV with data from PokeAPI. By default will fetch from id 1 to 10 unless there are other information on the body
func (api API) FillCSV(c *gin.Context) {

	requestBody := struct {
		From int `json:"from"`
		To   int `json:"to"`
	}{1, 10}

	if err := c.Bind(&requestBody); err != nil {
		c.Status(http.StatusBadRequest)
		fmt.Println(err)
		return
	}
	doneChan := make(chan bool)
	fetchResChan := api.FetchV2(requestBody.From, requestBody.To, doneChan)
	errChan := api.WriteToCSV(fetchResChan, doneChan)
	for err := range errChan {
		if err != nil {
			c.Status(http.StatusInternalServerError)
			fmt.Println("err: ", err)
			return
		}
	}
	c.Status(http.StatusOK)
}

//RefreshCache feeds the csv data and save in redis
func (api API) RefreshCache(c *gin.Context) {
	errChan := api.RefreshV2(c)
	for err := range errChan {
		if err != nil {
			c.Status(http.StatusInternalServerError)
			fmt.Println("err: ", err)
			return
		}
	}

	c.Status(http.StatusOK)
}

//GetPokemons return all pokemons in cache
func (api API) GetPokemons(c *gin.Context) {
	pokemons, err := api.getter.GetPokemons(c)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, pokemons)
}
