package renderer

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ONSdigital/dp-frontend-filter-dataset-controller/config"
	"github.com/ONSdigital/go-ns/healthcheck"
)

// Renderer provides an interface for a service template renderer
type Renderer interface {
	healthcheck.Client
	Do(path string, b []byte) ([]byte, error)
}

// ErrInvalidRendererResponse is returned when the renderer service does not respons
// with a status 200
type ErrInvalidRendererResponse struct {
	responseCode int
}

func (e ErrInvalidRendererResponse) Error() string {
	return fmt.Sprintf("invalid response from renderer service - status %d", e.responseCode)
}

// Renderer represents a template renderer for dp-frontend-filter-dataset-controller
type renderer struct {
	client *http.Client
	url    string
}

// New creates an instance of renderer with a default client
func New() Renderer {
	cfg := config.Get()

	return &renderer{
		client: &http.Client{},
		url:    cfg.RendererURL,
	}
}

// Healthcheck calls the healthcheck endpoint on the renderer and returns any errors
func (r *renderer) Healthcheck() error {
	resp, err := r.client.Get(r.url + "/healthcheck")
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return ErrInvalidRendererResponse{resp.StatusCode}
	}

	return nil
}

// Name returns the name of service the renderer is associated with
func (r *renderer) Name() string {
	return "renderer"
}

// Do sends a request to the renderer service to render a given template
func (r *renderer) Do(path string, b []byte) ([]byte, error) {
	// Renderer required JSON to be sent so if byte array is empty, set it to be
	// empty json
	if b == nil {
		b = []byte(`{}`)
	}

	req, err := http.NewRequest("POST", r.url+"/"+path, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrInvalidRendererResponse{resp.StatusCode}
	}

	return ioutil.ReadAll(resp.Body)
}
