package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
		fmt.Errorf("⚠️ Error: Could not load .env file: %v\n", err)
		os.Exit(1)
	}

	apiToken := os.Getenv("TIBBER_API_TOKEN")
	if apiToken == "" {
		fmt.Errorf("TIBBER_API_TOKEN environment variable is required")
		os.Exit(1)
	}

	title := os.Getenv("TITLE")
	if title == "" {
		title = "Default Title"
	}

	apiEndPoint := os.Getenv("TIBBER_API_ENDPOINT")

 	portFlag := flag.Int("port", 0, "HTTP server port")
	flag.Parse()

	var port int

	// Eerst proberen we de command-line flag
	if *portFlag > 0 {
		port = *portFlag
	} else {
		// Dan kijken we naar de omgevingsvariabele
		if portEnv := os.Getenv("PORT"); portEnv != "" {
			if p, err := strconv.Atoi(portEnv); err == nil && p > 0 {
				port = p
			}
		}
	}

	// Maak een nieuwe web dashboard
	webDashboard, err := NewWebDashboard(title, port, apiEndPoint, apiToken)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		os.Exit(1)
	} 

	// Start de web server
	if err := webDashboard.Start(); err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		os.Exit(1)
	}
}
