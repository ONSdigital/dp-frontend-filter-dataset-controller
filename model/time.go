package model

import core "github.com/ONSdigital/dp-renderer/v2/model"

// Time represents a time selection page
type Time struct {
	core.Page
	Data     TimeData `json:"data"`
	FilterID string   `json:"filter_id"`
}

// Data represents the metadata for the time page
type TimeData struct {
	LatestTime         TimeValue        `json:"latest_value"`
	FirstTime          TimeValue        `json:"fist_time"`
	Values             []TimeValue      `json:"values"`
	Months             []string         `json:"months"`
	Years              []string         `json:"years"`
	CheckedRadio       string           `json:"checked_radio"`
	FormAction         Link             `json:"form_action"`
	SelectedStartMonth string           `json:"selected_start_month"`
	SelectedStartYear  string           `json:"selected_start_year"`
	SelectedEndMonth   string           `json:"selected_end_month"`
	SelectedEndYear    string           `json:"selected_end_year"`
	Type               string           `json:"type"`
	DatasetTitle       string           `json:"dataset_title"`
	GroupedSelection   GroupedSelection `json:"grouped_selection"`
}

// TimeValue represents the data to display a single time value
type TimeValue struct {
	Month      string `json:"month,omitempty"`
	Year       string `json:"year,omitempty"`
	Option     string `json:"option"`
	IsSelected bool   `json:"is_selected"`
}

// GroupedSelection represents the data required to a group a selection
type GroupedSelection struct {
	Months    []Month `json:"months"`
	YearStart string  `json:"year_start"`
	YearEnd   string  `json:"year_end"`
}

// Month represents the data required to display a month
type Month struct {
	Name       string `json:"name"`
	IsSelected bool   `json:"is_selected"`
}
