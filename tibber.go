package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type tibberResponse struct {
	Data tibberData `json:"data"`
}

type tibberData struct {
	Viewer tibberViewer `json:"viewer"`
}

type tibberViewer struct {
	Homes []tibberHome `json:"homes"`
}

type tibberHome struct {
	CurrentSubscription tibberSubscription `json:"currentSubscription"`
}

type tibberSubscription struct {
	PriceInformation tibberPriceInformation `json:"priceInfo"`
}

type tibberPriceInformation struct {
	Current  tibberPrice   `json:"current"`
	Today    []tibberPrice `json:"today"`
	Tomorrow []tibberPrice `json:"tomorrow"`
}

type tibberPrice struct {
	Total    float64   `json:"total"`
	StartsAt time.Time `json:"startsAt"`
}

func readPrices(token string) ([]tibberPrice, error) {
	prices, err := readCurrentConsumption(token)
	if err != nil {
		return []tibberPrice{}, err
	}

	allPrices := append(prices.Data.Viewer.Homes[0].CurrentSubscription.PriceInformation.Today)
	allPrices = append(allPrices, prices.Data.Viewer.Homes[0].CurrentSubscription.PriceInformation.Tomorrow...)

	// Tomorrow's prices are sometimes empty
	if len(prices.Data.Viewer.Homes[0].CurrentSubscription.PriceInformation.Tomorrow) == 0 {
		log.Printf("No price data for tomorrow")
	}

	return allPrices, nil
}

func readCurrentConsumption(token string) (tibberResponse, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://api.tibber.com/v1-beta/gql", nil)
	if err != nil {
		return tibberResponse{}, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Add("Content-Type", "application/json")

	q := req.URL.Query()
	q.Add("query", "{viewer{homes{currentSubscription{priceInfo{current{total}today{total startsAt}tomorrow{total startsAt}}}}}}")
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return tibberResponse{}, err
	}

	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return tibberResponse{}, err
	}

	var response tibberResponse
	err = json.Unmarshal(bodyBytes, &response)
	if err != nil {
		return tibberResponse{}, err
	}

	return response, nil
}
