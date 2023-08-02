package model

// Link ...
type Link struct {
	URL   string `json:"url"`
	Label string `json:"label"`
}

// Filter ...
type Filter struct {
	Label     string `json:"label"`
	RemoveURL string `json:"remove_url"`
	ID        string `json:"id"`
}

// LatestVersion ...
type LatestVersion struct {
	DatasetLandingPageURL          string `json:"dataset_landing_page_url"`
	FilterJourneyWithLatestJourney string `json:"filter_journey_with_latest_version"`
}
