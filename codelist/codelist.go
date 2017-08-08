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
