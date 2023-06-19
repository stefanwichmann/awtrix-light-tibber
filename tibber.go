package main

import (
	"io"
	"log"
	"net/http"
)

func readCurrentConsumption(token string) string {
	client := &http.Client{}

	req, err := http.NewRequest("GET", "https://api.tibber.com/v1-beta/gql", nil)
	if err != nil {
		log.Fatal("request")
	}

	req.Header.Add("Authorization", "Bearer Gd06aSgUdatxeRTZGXzuOAd8Ypl1bbBnNETQnQUC2FE")
	req.Header.Add("Content-Type", "application/json")

	q := req.URL.Query()
	q.Add("query", "{viewer{homes{currentSubscription{priceInfo{current{total}today{total startsAt}tomorrow{total startsAt}}}}}}")
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Request")
	}

	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)
	return bodyString
}
