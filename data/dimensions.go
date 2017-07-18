package data

// Dimension represents a dimension to be filtered upon
type Dimension struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

type Filter struct {
	FilterID   string              `json:"filter_id"`
	Dataset    string              `json:"dataset"`
	Edition    string              `json:"edition"`
	Version    string              `json:"state"`
	Dimensions []Dimension         `json:"dimensions"`
	Downloads  map[string]Download `json:"downloads"`
	Events     map[string][]Event  `json:"events"`
}

type Download struct {
	URL  string `json:"url"`
	Size string `json:"size"`
}

type Event struct {
	Time    string `json:"time"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

type Dataset struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	URL         string  `json:"url"`
	ReleaseDate string  `json:"string"`
	Edition     string  `json:"edition"`
	Version     string  `json:"version"`
	Contact     Contact `json:"contact"`
}

type Contact struct {
	Name      string `json:"name"`
	Telephone string `json:"telephone"`
	Email     string `json:"email"`
}
