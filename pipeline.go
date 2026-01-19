package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

func RunPipeline(d *json.Decoder, counter *ByteCounter, totalSize int64, out *os.File, start time.Time) {

	// Find the "reporting_structure" array in the JSON data
	if !locateReportingStructure(d) {
		log.Fatalf("Failed to find 'reporting_structure' array in JSON")
	}

	// Initialize counters
	uniqueURLs := make(map[string]bool)
	foundCount := 0
	totalRecords := 0

	log.Println("Pipeline started. Scanning records...")

	// Stream and filter
	for d.More() {
		var r ReportingStructure

		// Decode one item at a time
		if err := d.Decode(&r); err != nil {
			log.Printf("Skipping malformed record: %v", err)
			continue
		}
		totalRecords++

		if totalRecords%1000 == 0 {
			logStatus(totalRecords, foundCount, counter.Count, totalSize, start)
		}

		// Check for plan first
		if isTargetPlan(&r) {
			for _, f := range r.InNetworkFiles {
				// Check for target location within that plan
				if isTargetLocation(f.Location, f.Description) {
					// Handle duplicates
					if !uniqueURLs[f.Location] {
						uniqueURLs[f.Location] = true
						foundCount++

						// Write to output file immediately
						fmt.Fprintln(out, f.Location)
					}
				}
			}
		}
	}

	log.Printf("\nPipeline Complete! Processed %d records in %s. Found %d URLs.", totalRecords, time.Since(start), foundCount)
}

func locateReportingStructure(d *json.Decoder) bool {
	for {
		t, err := d.Token()
		if err != nil {
			return false
		}
		if s, ok := t.(string); ok && s == "reporting_structure" {
			if _, err := d.Token(); err == nil {
				return true
			}
		}
	}
}

func logStatus(records, found int, bytesRead, totalBytes int64, start time.Time) {
	elapsed := time.Since(start).Round(time.Second)
	mb := float64(bytesRead) / 1024 / 1024

	percentStr := "Unknown%"
	if totalBytes > 0 {
		pct := (float64(bytesRead) / float64(totalBytes)) * 100
		percentStr = fmt.Sprintf("%.1f%%", pct)
	}

	log.Printf("Scan: %d | Found: %d | Progress: %.0fMB (%s) | Time: %s", records, found, mb, percentStr, elapsed)
}
