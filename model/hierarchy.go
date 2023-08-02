package model

import core "github.com/ONSdigital/dp-renderer/v2/model"

// Page ...
type Hierarchy struct {
	core.Page
	Data     HierarchyData `json:"data"`
	FilterID string        `json:"filter_id"`
}

// Hierarchy ...
type HierarchyData struct {
	Title           string   `json:"title"`
	SaveAndReturn   Link     `json:"save_and_return"`
	Cancel          Link     `json:"cancel"`
	FiltersAmount   string   `json:"filters_amount"`
	AddAllFilters   AddAll   `json:"add_all"`
	FilterList      []List   `json:"filter_list"`
	FiltersAdded    []Filter `json:"filters_added"`
	RemoveAll       Link     `json:"remove_all"`
	GoBack          Link     `json:"go_back"`
	DimensionName   string   `json:"dimension_name"`
	Parent          string   `json:"parent"`
	Type            string   `json:"type"`
	Metadata        Metadata `json:"metadata"`
	DatasetTitle    string   `json:"dataset_title"`
	SearchURL       string   `json:"search_url"`
	IsSearchResults bool     `json:"is_search_results"`
	Query           string   `json:"query"`
	IsSearchError   bool     `json:"is_search_error"`
	LandingPageURL  string   `json:"landing_page_url"`
	HasData         bool     `json:"has_data"`
}

// AddAll ...
type AddAll struct {
	Amount string `json:"amount"`
	URL    string `json:"url"`
}

// Metadata ...
type Metadata struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// List ...
type List struct {
	Label    string `json:"label"`
	Selected bool   `json:"selected"`
	SubNum   string `json:"sub_num"`
	ID       string `json:"id"`
	SubType  string `json:"sub_type"`
	SubURL   string `json:"sub_url"`
	HasData  bool   `json:"has_data"`
}
