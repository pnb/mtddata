package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

func fetchDepartures(stopID string) ([]byte, error) {
	resp, err := http.Get(os.Getenv("MTDDATA_API_URL") +
		"/json/getdeparturesbystop?key=" + os.Getenv("MTDDATA_API_KEY") +
		"&stop_id=" + stopID + "&pt=60") // pt is "preview time"/future (max 60 minutes)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code: %d",
			resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func saveToJSONL(data []byte) error {
	file, err := os.OpenFile(os.Getenv("MTDDATA_OUTPUT_PATH"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(string(data) + "\n")
	return err
}

func main() {
	// Verify settings are in env
	if os.Getenv("MTDDATA_API_URL") == "" {
		log.Fatal("MTDDATA_API_URL is not set in the environment")
	}
	if os.Getenv("MTDDATA_API_KEY") == "" {
		log.Fatal("MTDDATA_API_KEY is not set in the environment")
	}
	if os.Getenv("MTDDATA_OUTPUT_PATH") == "" {
		log.Fatal("MTDDATA_OUTPUT_PATH is not set in the environment")
	}
	if os.Getenv("MTDDATA_STOP_IDS") == "" {
		log.Fatal("MTDDATA_STOP_IDS is not set in the environment")
	}
	if os.Getenv("MTDDATA_UPDATE_INTERVAL_SECONDS") == "" {
		log.Fatal("MTDDATA_UPDATE_INTERVAL_SECONDS is not set in the environment")
	}
	// Fetch data regularly, forever
	updateInterval, err := time.ParseDuration(
		os.Getenv("MTDDATA_UPDATE_INTERVAL_SECONDS") + "s")
	if err != nil {
		log.Fatal("Failed to parse MTDDATA_UPDATE_INTERVAL_SECONDS as a duration:", err)
	}
	ticker := time.NewTicker(updateInterval)
	defer ticker.Stop()
	for range ticker.C {
		stopIDs := strings.Split(os.Getenv("MTDDATA_STOP_IDS"), ",")
		for _, stopID := range stopIDs {
			data, err := fetchDepartures(stopID)
			if err != nil {
				log.Print("Failed to fetch departures for stop", stopID, ":", err)
				continue
			}
			if err := saveToJSONL(data); err != nil {
				log.Print("Failed to save departures data for stop", stopID, ":", err)
				continue
			}
		}
		log.Println("Fetched and saved departures data for", len(stopIDs), "stops")
	}
}
