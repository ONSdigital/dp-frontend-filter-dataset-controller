package model

import core "github.com/ONSdigital/dp-renderer/v2/model"

// Age represents an age selection page
type Age struct {
	core.Page
	Data     AgeData `json:"data"`
	FilterID string  `json:"filter_id"`
}

// Data represents the data for the age page
type AgeData struct {
	Youngest       string     `json:"youngest"`
	Oldest         string     `json:"oldest"`
	FirstSelected  string     `json:"first_selected"`
	LastSelected   string     `json:"last_selected"`
	Ages           []AgeValue `json:"ages"`
	CheckedRadio   string     `json:"checked_radio"`
	FormAction     Link       `json:"form_action"`
	HasAllAges     bool       `json:"has_all_ages"`
	AllAgesOption  string     `json:"all_ages_option"`
	FeedbackAPIURL string     `json:"feedback_api_url"`
}

// Value represents a single age value
type AgeValue struct {
	Label      string
	Option     string
	IsSelected bool
}
