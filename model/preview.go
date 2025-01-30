package model

import core "github.com/ONSdigital/dp-renderer/v2/model"

// Preview represents the data for a preview page
type Preview struct {
	core.Page
	Data                 PreviewPage `json:"data"`
	EnableDatasetPreview bool        `json:"enable_dataset_preview"`
	IsPreviewLoaded      bool        `json:"is_preview_loaded"`
	IsDownloadLoaded     bool        `json:"is_download_loaded"`
	NoDimensionData      bool        `json:"no_dimension_data"`
}

// PreviewPage represents the metadata for a preview page
type PreviewPage struct {
	FilterID              string             `json:"filter_id"`
	Downloads             []Download         `json:"downloads"`
	Dimensions            []PreviewDimension `json:"dimensions"`
	IsLatestVersion       bool               `json:"is_latest_version"`
	LatestVersion         LatestVersion      `json:"latest_version"`
	CurrentVersionURL     string             `json:"current_version_url"`
	DatasetTitle          string             `json:"dataset_title"`
	DatasetID             string             `json:"dataset_id"`
	Edition               string             `json:"edition"`
	ReleaseDate           string             `json:"release_date"`
	UnitOfMeasurement     string             `json:"unit_of_measurement"`
	SingleValueDimensions []PreviewDimension `json:"single_value_dimensions"`
	FilterOutputID        string             `json:"filter_output_id"`
	EnableFeedbackAPI     bool               `json:"enable_feedback_api"`
	FeedbackAPIURL        string             `json:"feedback_api_url"`
}

// Download has the details for an individual downloadable file
type Download struct {
	Extension string `json:"extension"`
	Size      string `json:"size"`
	URI       string `json:"uri"`
	Skipped   bool   `json:"skipped"`
}

// PreviewDimension represents a single dimension for the preview page
type PreviewDimension struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}
