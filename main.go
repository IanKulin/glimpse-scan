package main

// Collects metrics from a https://github.com/IanKulin/vitals-glimpse
// end point and stores them in InfluxDB
//
// Endpoint output looks like this:
// {"title":"vitals-glimpse",
//  "version":0.2,
//  "mem_status":"mem_okay",
//  "mem_percent":46,
//  "disk_status":"disk_okay",
//  "disk_percent":79,
//  "cpu_status":"cpu_okay",
//  "cpu_percent":0
// }


import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)


// Vitals structure to hold metrics data
type Vitals struct {
	Title       string  `json:"title"`
	Version     float32 `json:"version"`
	MemPercent  int     `json:"mem_percent"`
	DiskPercent int     `json:"disk_percent"`
	CpuPercent  int     `json:"cpu_percent"`
}


type ServerConfig struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}


func main() {

    org := os.Getenv("INFLUXDB_ORG")
    bucket := os.Getenv("INFLUXDB_BUCKET")
    token := os.Getenv("INFLUXDB_ADMIN_TOKEN")
    url := os.Getenv("INFLUXDB_URL")
	pollingIntervalStr := os.Getenv("POLLING_INTERVAL_MINUTES")

    if org == "" || bucket == "" || token == "" || url == "" || pollingIntervalStr == "" {
        log.Fatal("Missing required environment variables")
    }

    fmt.Println("InfluxDB Organization:", org)
    fmt.Println("InfluxDB Bucket:", bucket)
    fmt.Println("InfluxDB URL:", url)
    fmt.Println("Polling Interval (minutes):", pollingIntervalStr)

	pollingInterval, err := strconv.Atoi(pollingIntervalStr)
	if err != nil {
		log.Fatalf("Error parsing POLLING_INTERVAL_MINUTES: %v", err)
	}

	// Load server configurations from the JSON file
	servers, err := loadServersConfig("data/servers.json")
	if err != nil {
		log.Fatalf("Error loading servers config: %v", err)
	}

	// Create InfluxDB client
	client := influxdb2.NewClient(url, token)
	defer client.Close()

	// Get non-blocking write client
	writeAPI := client.WriteAPI(org, bucket)

	// Get errors channel and log errors
	errorsCh := writeAPI.Errors()
	go func() {
		for err := range errorsCh {
			fmt.Printf("write error: %s\n", err.Error())
		}
	}()

	// Poll servers every pollingInterval minutes
	// from "POLLING_INTERVAL_MINUTES" environment variable 
	for {
		for _, server := range servers {
			go pollAndStoreMetrics(server.Name, server.URL, writeAPI)
		}
		time.Sleep(time.Duration(pollingInterval) * time.Minute)
	}
}


// Loads server configuration from a JSON file
func loadServersConfig(filepath string) ([]ServerConfig, error) {
	file, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var servers []ServerConfig
	err = json.Unmarshal(file, &servers)
	if err != nil {
		return nil, err
	}

	return servers, nil
}


// Fetch metrics from a server and store them in InfluxDB
func pollAndStoreMetrics(serverName, serverURL string, writeAPI api.WriteAPI) {
	resp, err := http.Get(serverURL)
	if err != nil {
		fmt.Printf("Failed to poll server %s (%s): %v\n", serverName, serverURL, err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read response from server %s (%s): %v\n", serverName, serverURL, err)
		return
	}

	var vitals Vitals
	if err := json.Unmarshal(body, &vitals); err != nil {
		fmt.Printf("Failed to parse JSON from server %s (%s): %v\n", serverName, serverURL, err)
		return
	}

	// we're expecting JSON version 0.2 or higher
	if vitals.Version < 0.2 {
		fmt.Printf("Server %s (%s) returned invalid JSON version %f\n", serverName, serverURL, vitals.Version)
		return
	}

	// we're expecting title to be "vitals-glimpse"
	if vitals.Title != "vitals-glimpse" {
		fmt.Printf("Server %s (%s) returned invalid title %s\n", serverName, serverURL, vitals.Title)
		return
	}
	
	// create point for InfluxDB
	point := influxdb2.NewPoint(
		"server_metrics",
		map[string]string{
			"server": serverName,
		},
		map[string]interface{}{
			"mem_percent":  vitals.MemPercent,
			"disk_percent": vitals.DiskPercent,
			"cpu_percent":  vitals.CpuPercent,
		},
		time.Now(),
	)

	// write point to InfluxDB
	writeAPI.WritePoint(point)
	fmt.Printf("Data written for server %s\n", serverName)
}
