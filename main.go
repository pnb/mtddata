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

func fetchWeather() ([]byte, error) {
	resp, err := http.Get("https://api.open-meteo.com/v1/forecast?" +
		"latitude=40.1166557&longitude=88.2297261" +
		"&current=temperature_2m,relative_humidity_2m,apparent_temperature," +
		"wind_speed_10m,wind_gusts_10m,precipitation,rain,cloud_cover," +
		"surface_pressure,snowfall,showers,weather_code,pressure_msl")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Weather API request failed with status code: %d",
			resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func saveToJSONL(data []byte, filepath string) error {
	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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
	weatherPath := os.Getenv("MTDDATA_WEATHER_OUTPUT_PATH")
	if weatherPath == "" {
		log.Println("MTDDATA_WEATHER_OUTPUT_PATH not set, so weather will not be saved")
	}

	// Fetch data regularly, forever
	updateInterval, err := time.ParseDuration(
		os.Getenv("MTDDATA_UPDATE_INTERVAL_SECONDS") + "s")
	if err != nil {
		log.Fatal("Failed to parse MTDDATA_UPDATE_INTERVAL_SECONDS as a duration:", err)
	}
	log.Println("Starting; data will not be fetched until first update interval")
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
			if err := saveToJSONL(data, os.Getenv("MTDDATA_OUTPUT_PATH")); err != nil {
				log.Print("Failed to save departures data for stop", stopID, ":", err)
				continue
			}
		}
		if weatherPath != "" {
			data, err := fetchWeather()
			if err != nil {
				log.Print("Failed to fetch weather:", err)
			} else if err := saveToJSONL(data, weatherPath); err != nil {
				log.Print("Failed to save weather data:", err)
			}
		}
		log.Println("Fetched and saved departures data for", len(stopIDs), "stops")
	}
}
