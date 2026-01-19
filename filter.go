package main

import "strings"

// Known Network IDs for New York / PPO products from EIN lookup
var TargetCodes = map[string]bool{
	"72A0": true,
	"71A0": true,
	"39B0": true,
	"42B0": true,
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
	upper := strings.ToUpper(s)

	return strings.Contains(upper, "NY") || strings.Contains(upper, "NEW YORK")
}
