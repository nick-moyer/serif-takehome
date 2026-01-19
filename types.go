package main

// Maps to the JSON object in the stream
type ReportingStructure struct {
	ReportingPlans []struct {
		PlanName string `json:"plan_name"`
	} `json:"reporting_plans"`

	InNetworkFiles []struct {
		Description string `json:"description"`
		Location    string `json:"location"`
	} `json:"in_network_files"`
}
