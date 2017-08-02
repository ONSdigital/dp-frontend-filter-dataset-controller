package filter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/data"
)

// ErrInvalidFilterAPIResponse is returned when the filter api does not respond
// with a valid status
type ErrInvalidFilterAPIResponse struct {
	expectedCode int
	actualCode   int
	uri          string
}

func (e ErrInvalidFilterAPIResponse) Error() string {
	return fmt.Sprintf("invalid response from filter api - should be: %d, got: %d, path: %s",
		e.expectedCode,
		e.actualCode,
		e.uri,
	)
}

var _ error = ErrInvalidFilterAPIResponse{}

// Client is a filter api client which can be used to make requests to the server
type Client struct {
	cli *http.Client
	url string
}

// New creates a new instance of Client with a given filter api url
func New(filterAPIURL string) *Client {
	return &Client{
		cli: &http.Client{Timeout: 5 * time.Second},
		url: filterAPIURL,
	}
}

// GetDimension returns information on a requested dimension name for a given filterID
func (c *Client) GetDimension(filterID, name string) (dim data.FilterDimension, err error) {
	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s", c.url, filterID, name)
	resp, err := c.cli.Get(uri)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = &ErrInvalidFilterAPIResponse{http.StatusOK, resp.StatusCode, uri}
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if err = json.Unmarshal(b, &dim); err != nil {
		return
	}

	return
}

// GetDimensions will return the dimensions associated with the provided filter id
func (c *Client) GetDimensions(filterID string) (dims []data.FilterDimension, err error) {
	uri := fmt.Sprintf("%s/filters/%s/dimensions", c.url, filterID)
	resp, err := c.cli.Get(uri)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = &ErrInvalidFilterAPIResponse{http.StatusOK, resp.StatusCode, uri}
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	dimensions := struct {
		Dims []data.FilterDimension `json:"dimensions"`
	}{}

	if err = json.Unmarshal(b, &dimensions); err != nil {
		return
	}

	dims = dimensions.Dims
	return
}

// GetDimensionOptions ...
func (c *Client) GetDimensionOptions(filterID, name string) (fdv data.DimensionValues, err error) {
	uri := fmt.Sprintf("%s/filters/%s/dimensions/%s/options", c.url, filterID, name)
	resp, err := c.cli.Get(uri)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = &ErrInvalidFilterAPIResponse{http.StatusOK, resp.StatusCode, uri}
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	err = json.Unmarshal(b, &fdv)
	return
}

// GetJobState ...
func (c *Client) GetJobState(filterID string) (f data.Filter, err error) {
	uri := fmt.Sprintf("%s/filters/%s", c.url, filterID)
	resp, err := c.cli.Get(uri)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = &ErrInvalidFilterAPIResponse{http.StatusOK, resp.StatusCode, uri}
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	err = json.Unmarshal(b, &f)
	return
}
