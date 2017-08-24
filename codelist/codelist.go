package codelist

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/data"
)

// ErrInvalidCodelistAPIResponse is returned when the codelist api does not respond
// with a valid status
type ErrInvalidCodelistAPIResponse struct {
	expectedCode int
	actualCode   int
	uri          string
}

func (e ErrInvalidCodelistAPIResponse) Error() string {
	return fmt.Sprintf("invalid response from codelist api - should be: %d, got: %d, path: %s",
		e.expectedCode,
		e.actualCode,
		e.uri,
	)
}

var _ error = ErrInvalidCodelistAPIResponse{}

// Client is a codelist api client which can be used to make requests to the server
type Client struct {
	cli *http.Client
	url string
}

// New creates a new instance of Client with a given filter api url
func New(codelistAPIURL string) *Client {
	return &Client{
		cli: &http.Client{Timeout: 5 * time.Second},
		url: codelistAPIURL,
	}
}

// Healthcheck calls the healthcheck endpoint on the api and alerts the caller of any errors
func (c *Client) Healthcheck() error {
	resp, err := c.cli.Get(c.url + "/healthcheck")
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return &ErrInvalidCodelistAPIResponse{http.StatusOK, resp.StatusCode, "/healthcheck"}
	}

	return nil
}

// Name returns the name of the api the client connects to
func (c *Client) Name() string {
	return "code-list-api"
}

// GetValues ...
func (c *Client) GetValues(id string) (vals data.DimensionValues, err error) {
	uri := fmt.Sprintf("%s/code-lists/%s/codes", c.url, id)
	resp, err := c.cli.Get(uri)
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = &ErrInvalidCodelistAPIResponse{http.StatusOK, resp.StatusCode, uri}
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	err = json.Unmarshal(b, &vals)
	return
}

func (c *Client) GetIdNameMap(id string) (map[string]string, error) {
	uri := fmt.Sprintf("%s/code-lists/%s/codes", c.url, id)
	resp, err := c.cli.Get(uri)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		err = &ErrInvalidCodelistAPIResponse{http.StatusOK, resp.StatusCode, uri}
		return nil, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var vals data.DimensionValues
	if err = json.Unmarshal(b, &vals); err != nil {
		return nil, err
	}

	idNames := make(map[string]string)
	for _, val := range vals.Items {
		idNames[val.ID] = val.Label
	}

	return idNames, nil
}
