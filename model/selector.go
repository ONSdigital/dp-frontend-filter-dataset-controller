package model

import core "github.com/ONSdigital/dp-renderer/v2/model"

// Selector represents the data for a selector page
type Selector struct {
	core.Page
	Pagination core.Pagination `json:"pagination,omitempty"`
	Data       ListSelector    `json:"data"`
	FilterID   string          `json:"job_id"`
}

// ListSelector represents the metadata for a selector page
type ListSelector struct {
	Title         string   `json:"title"`
	AddFromRange  Link     `json:"add_from_range"`
	AddAllChecked bool     `json:"add_all_checked"`
	SaveAndReturn Link     `json:"save_and_return"`
	Cancel        Link     `json:"cancel"`
	FiltersAmount int      `json:"filters_amount"`
	FiltersAdded  []Filter `json:"filters_added"`
	AddAllInRange Link     `json:"add_all"`
	RemoveAll     Link     `json:"remove_all"`
	RangeData     Range    `json:"range_values"`
	DatasetTitle  string   `json:"dataset_title"`
}

// Range represents the data to display a range
type Range struct {
	URL    string  `json:"url"`
	Values []Value `json:"values"`
}

// Value represents the data to display an individual value
type Value struct {
	Label      string `json:"label"`
	ID         string `json:"id"`
	IsSelected bool   `json:"is_selected"`
}
