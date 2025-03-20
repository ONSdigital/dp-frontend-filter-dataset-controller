package model

import core "github.com/ONSdigital/dp-renderer/v2/model"

// Overview represents the data for a overview page
type Overview struct {
	core.Page
	Data     FilterOverview `json:"data"`
	FilterID string         `json:"filter_id"`
}

// FilterOverview represents the metadata for a overview page
type FilterOverview struct {
	Dimensions         []Dimension   `json:"dimensions"`
	UnsetDimensions    []string      `json:"unset_dimensions"`
	ClearAll           Link          `json:"clear_all"`
	Cancel             Link          `json:"cancel"`
	IsLatestVersion    bool          `json:"is_latest_version"`
	LatestVersion      LatestVersion `json:"latest_version"`
	DatasetTitle       string        `json:"dataset_title"`
	HasUnsetDimensions bool          `json:"has_unset_dimensions"`
	FeedbackAPIURL     string        `json:"feedback_api_url"`
}

// Dimension represents the data for a single dimension
type Dimension struct {
	Filter          string   `json:"filter"`
	AddedCategories []string `json:"added_categories"`
	Link            Link     `json:"link"`
	HasNoCategory   bool     `json:"has_no_category"`
}
