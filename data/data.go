package data

// Dimension represents a dimension to be filtered upon
type Dimension struct {
	Name      string    `json:"name"`
	Values    []string  `json:"values"`
	Hierarchy Hierarchy `json:"hierarchy,omitempty"`
}

// FilterDimension represents a dimension response from the filter api
type FilterDimension struct {
	Name string `json:"name"`
	URI  string `json:"dimension_url"`
}

// DimensionValues ...
type DimensionValues struct {
	Items           []DimensionValueItem `json:"items"`
	NumberOfResults int                  `json:"number_of_results"`
}

// DimensionValueItem ...
type DimensionValueItem struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Name  string `json:"name"`
}

// Filter represents a response model from the filter api
type Filter struct {
	FilterID   string              `json:"filter_id"`
	Dataset    string              `json:"dataset"`
	Edition    string              `json:"edition"`
	Version    string              `json:"version"`
	State      string              `json:"state"`
	Dimensions []Dimension         `json:"dimensions"`
	Downloads  map[string]Download `json:"downloads"`
	Events     map[string][]Event  `json:"events"`
}

// Download represents a download within a filter from api response
type Download struct {
	URL  string `json:"url"`
	Size string `json:"size"`
}

// Event represents an event from a filter api response
type Event struct {
	Time    string `json:"time"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

// Dataset represents a response model from the dataset api
type Dataset struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	URL         string  `json:"url"`
	ReleaseDate string  `json:"release_date"`
	NextRelease string  `json:"next_release"`
	Edition     string  `json:"edition"`
	Version     string  `json:"version"`
	Contact     Contact `json:"contact"`
}

// Contact represents a response model within a dataset
type Contact struct {
	Name      string `json:"name"`
	Telephone string `json:"telephone"`
	Email     string `json:"email"`
}

type Hierarchy struct {
	ID       string   `json:"id"`
	Label    string   `json:"label"`
	Children []Child  `json:"children"`
	Parents  []Parent `json:"parents"`
}

type Metadata struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Child struct {
	ID               string `json:"id"`
	Label            string `json:"label"`
	URI              string `json:"uri"`
	NumberofChildren int    `json:"noOfChildren"`
	HasDataPoint     bool   `json:"hasDataPoint"`
}

type Parent struct {
	URI   string `json:"uri"`
	Label string `json:"label"`
}
