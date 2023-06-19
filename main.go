package main

import "log"

func main() {
	log.Println("Hello World")
	result := readCurrentConsumption("")
	log.Printf(result)
}
