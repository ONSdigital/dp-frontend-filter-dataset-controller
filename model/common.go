package model

// Link represents the data required to display a link
type Link struct {
	URL   string `json:"url"`
	Label string `json:"label"`
}

// Filter represents the data for a single filter item
type Filter struct {
	Label     string `json:"label"`
	RemoveURL string `json:"remove_url"`
	ID        string `json:"id"`
}

// LatestVersion represents the data to display the latest version
type LatestVersion struct {
	DatasetLandingPageURL          string `json:"dataset_landing_page_url"`
	FilterJourneyWithLatestJourney string `json:"filter_journey_with_latest_version"`
}
