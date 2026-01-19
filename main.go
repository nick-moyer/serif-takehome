package main

import (
	"compress/gzip"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	IndexFileURL   = "https://antm-pt-prod-dataz-nogbd-nophi-us-east1.s3.amazonaws.com/anthem/2026-01-01_anthem_index.json.gz"
	OutputFileName = "output.txt"
)

func main() {
	// Start the Timer
	startTime := time.Now()

	// Schedule the "Stop Timer" function to run when main() exits
	defer func() {
		elapsed := time.Since(startTime)
		log.Printf("Total execution time: %s", elapsed)
	}()

	// Create connection
	log.Println("Connecting to stream:...")
	resp, err := http.Get(IndexFileURL)
	if err != nil {
		log.Fatalf("Failed to fetch URL: %v", err)
	}
	defer resp.Body.Close() // Cleanup

	if resp.StatusCode != 200 {
		log.Fatalf("Server returned error: %d", resp.StatusCode)
	}

	// Create output file
	log.Printf("Creating output file: %s\n", OutputFileName)
	output, err := os.Create(OutputFileName)
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
		return
	}
	defer output.Close() // Cleanup

	// Create the counter for bytes read
	counter := &ByteCounter{Reader: resp.Body}

	// Create reader
	gz, err := gzip.NewReader(counter)
	if err != nil {
		log.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer gz.Close() // Cleanup

	// Create decoder for JSON data in the gzip stream
	decoder := json.NewDecoder(gz)

	// Send to the pipeline
	RunPipeline(decoder, counter, resp.ContentLength, output, startTime)
}
