package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

type ReportingStructure struct {
	ReportingPlans []struct {
		PlanName string `json:"plan_name"`
	} `json:"reporting_plans"`

	InNetworkFiles []struct {
		Description string `json:"description"`
		Location    string `json:"location"`
	} `json:"in_network_files"`
}

// Known Network IDs for New York / PPO products from EIN lookup
var TargetCodes = map[string]bool{
	"72A0": true,
	"71A0": true,
	"39B0": true,
	"42B0": true,
}

const (
	IndexFileURL   = "https://antm-pt-prod-dataz-nogbd-nophi-us-east1.s3.amazonaws.com/anthem/2026-01-01_anthem_index.json.gz"
	OutputFileName = "output.txt"
)

func main() {
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

	// Create reader
	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		log.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer gz.Close() // Cleanup

	// Create decoder for JSON data in the gzip stream
	decoder := json.NewDecoder(gz)

	// Find the "reporting_structure" array in the JSON data
	foundArray := false
	for {
		t, err := decoder.Token()
		if err != nil {
			log.Fatalf("Failed to read JSON token: %v", err)
		}
		if s, ok := t.(string); ok && s == "reporting_structure" {
			if _, err := decoder.Token(); err != nil {
				log.Fatalf("Failed to decode 'reporting_structure' array: %v", err)
			}
			foundArray = true
			break
		}
	}

	if !foundArray {
		log.Fatalf("Failed to find 'reporting_structure' array")
	}

	uniqueURLs := make(map[string]bool)
	count := 0

	// Stream and filter
	for decoder.More() {
		var r ReportingStructure

		// Decode one item at a time
		if err := decoder.Decode(&r); err != nil {
			continue
		}

		// Check for plan first
		if isTargetPlan(&r) {
			for _, f := range r.InNetworkFiles {
				// Check for target location within that plan
				if isTargetLocation(f.Location, f.Description) {
					// Handle duplicates
					if !uniqueURLs[f.Location] {
						uniqueURLs[f.Location] = true
						count++

						if (count%100 == 0) {
							log.Printf("Found %d unique urls\n", count)
						}

						// Write to output file immediately
						fmt.Fprintln(output, f.Location)
					}
				}
			}
		}
	}

	log.Printf("\nSuccess! Found %d unique URLs. Saved to %s.\n", count, OutputFileName)
}

func isTargetPlan(r *ReportingStructure) bool {
    for _, p := range r.ReportingPlans {
		name := strings.ToUpper(p.PlanName)

		// Ignore non-PPO plans
		if !strings.Contains(name, "PPO") {
			continue
		}

		// Check for Anthem
		if strings.Contains(name, "ANTHEM") {
			return true
		}
	}
	return false
}

func isTargetLocation(loc, desc string) bool {
	// Check for target codes first
	for code := range TargetCodes {
		if strings.Contains(loc, "_"+code+"_") {
			return true
		}
	}

	// Text fallback (Discovery)
	return isNY(loc) || isNY(desc)
}

func isNY(s string) bool {
	return strings.Contains(s, "NY") || strings.Contains(s, "New York")
}