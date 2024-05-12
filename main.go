package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"time"
)

var tibberDemoToken = "5K4MVS-OjfWhK_4yrjOlFe1F6kJXPVf7eQYggo8ebAE"
var flagTibberToken = flag.String("tibberToken", lookupEnv("TIBBER_TOKEN", tibberDemoToken), "Your Tibber developer API token")
var flagAwtrixIP = flag.String("awtrixIP", lookupEnv("AWTRIX_IP", "127.0.0.1"), "The IPv4 address of your Awtrix light device")

var customAppName = "tibberPrices"
var chartBarCount = 36 - 12

var knownPrices []tibberPrice

func main() {
	flag.Parse()
	if *flagTibberToken == tibberDemoToken {
		log.Print("Using Tibber demo token. Please provide your own developer token via --tibberToken for real data")
	}

	for {
		updatePrices()
		updateDisplay()

		nextUpdate := durationUntilNextFullHour()
		log.Printf("Sleeping for %s", nextUpdate)
		time.Sleep(nextUpdate)
	}
}

func durationUntilNextFullHour() time.Duration {
	now := time.Now()
	nextFullHour := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 1, 0, now.Location())
	nextFullHour = nextFullHour.Add(1 * time.Hour)
	return time.Until(nextFullHour)
}

func updatePrices() {
	log.Println("Fetching Tibber prices...")
	prices, err := readPrices(*flagTibberToken)
	if err != nil {
		log.Printf("Could not fetch prices: %v", err)
		return
	}

	log.Print("Detecting price changes...")
	detectPriceChanges(knownPrices, prices)

	log.Print("Updating known prices")
	knownPrices = prices
}

func updateDisplay() {
	if len(knownPrices) == 0 {
		log.Printf("No prices available, skipping display update")
		return
	}

	historicPrices, upcomingPrices := splitPrices(knownPrices)

	// Limit historic prices to the last 4
	if len(historicPrices) >= 4 {
		historicPrices = historicPrices[len(historicPrices)-4:]
	}
	relevantPrices := append(historicPrices, upcomingPrices...)
	if len(relevantPrices) > chartBarCount {
		relevantPrices = relevantPrices[:chartBarCount]
	}

	// Print prices
	log.Printf("Identified the following relevant prices")
	for _, price := range relevantPrices {
		log.Printf("Starting at %s: %f", price.StartsAt, price.Total)
	}

	commandsText := []AwtrixDrawCommand{{Command: "dt", X: 0, Y: 1, Text: fmt.Sprintf("%d¢", roundedPrice(historicPrices[len(historicPrices)-1].Total)), Color: "#FFFFFF"}}
	commandsChart := mapToDrawingCommands(relevantPrices)
	app := AwtrixApp{Draw: append(commandsText, commandsChart...)}

	log.Printf("Drawing %d prices...", len(commandsChart))
	err := postApplication(*flagAwtrixIP, customAppName, app)
	if err != nil {
		log.Printf("Could not update custom application: %v", err)
	}
}

func splitPrices(prices []tibberPrice) ([]tibberPrice, []tibberPrice) {
	var historicPrices []tibberPrice
	var upcomingPrices []tibberPrice

	for _, price := range prices {
		if price.StartsAt.Before(time.Now()) {
			historicPrices = append(historicPrices, price)
		} else if price.StartsAt.After(time.Now()) {
			upcomingPrices = append(upcomingPrices, price)
		} else {
			log.Fatalf("Can't place price %+v", price)
		}
	}

	return historicPrices, upcomingPrices
}

func mapToDrawingCommands(prices []tibberPrice) []AwtrixDrawCommand {
	var commands []AwtrixDrawCommand

	// Find min and max price
	minPrice := prices[0].Total
	maxPrice := prices[0].Total
	for _, price := range prices {
		if price.Total < minPrice {
			minPrice = price.Total
		}
		if price.Total > maxPrice {
			maxPrice = price.Total
		}

	}

	// Map price range to pixel range
	yMin := 1
	yMax := 8
	slope := 1.0 * float64(yMax-yMin) / (maxPrice - minPrice)
	xOffset := 12

	for i, price := range prices {
		scaledPrice := float64(yMin) + slope*(price.Total-minPrice)
		color := mapPriceToColor(price)
		log.Printf("Mapping price %f to %d (Min: %f, Max: %f, Color: %s)", price.Total, int(scaledPrice), minPrice, maxPrice, color)
		command := AwtrixDrawCommand{Command: "df", X: xOffset + i, Y: yMax - int(scaledPrice), Width: 1, Height: yMax, Color: color}
		commands = append(commands, command)
	}

	return commands
}

func roundedPrice(price float64) int {
	return int(math.Round(price * 100))
}

func mapPriceToColor(price tibberPrice) string {
	if price.StartsAt.Day() == time.Now().Day() && price.StartsAt.Hour() == time.Now().Hour() {
		return "#FFFFFF"
	}

	switch {
	case price.Total <= 0:
		return "#215d6e"
	case price.Total < 0.25:
		return "#5ba023"
	case price.Total < 0.28:
		return "#7b9632"
	case price.Total < 0.30:
		return "#9b8c41"
	case price.Total < 0.33:
		return "#ba8250"
	case price.Total < 0.35:
		return "#da785f"
	default:
		return "#fa6e6e"
	}
}

func detectPriceChanges(oldPrices []tibberPrice, newPrices []tibberPrice) {
	for _, newPrice := range newPrices {
		for _, oldPrice := range oldPrices {
			if newPrice.StartsAt == oldPrice.StartsAt && newPrice.Total != oldPrice.Total {
				log.Printf("Price for %s changed from %f to %f", newPrice.StartsAt, oldPrice.Total, newPrice.Total)
			}
		}
	}
}
