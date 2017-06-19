package handlers

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/ONSdigital/dp-frontent-filter-dataset-controller/renderer"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
)

// CMD represents the handlers for CMD
type CMD struct {
	r renderer.Renderer
}

// NewCMD creates a new instance of CMD
func NewCMD(r renderer.Renderer) *CMD {
	return &CMD{r: r}
}

// Landing handles the controller functionality for the landing page
func (c *CMD) Landing(w http.ResponseWriter, req *http.Request) {
	b, err := c.r.Do("dataset/startpage", nil)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(b); err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// CreateJobID controls the creating of a job idea when a new user journey is
// requested
func (c *CMD) CreateJobID(w http.ResponseWriter, req *http.Request) {
	// TODO: This is a stubbed job id - replace with real job id from backend once
	// code has been written
	jobID := rand.Intn(100000000)
	jid := strconv.Itoa(jobID)

	log.Trace("created job id", log.Data{"job_id": jid})
	http.Redirect(w, req, "/dataset/cmd/"+jid, 301)
}

// Middle controls the rendering of a "middle" cmd page - this will be replaced
// by other handlers when further pages are defined by UX
func (c *CMD) Middle(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	page := make(map[string]interface{})
	data := make(map[string]interface{})

	data["job_id"] = vars["jobID"]
	page["data"] = data

	body, err := json.Marshal(page)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	b, err := c.r.Do("dataset/middlepage", body)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(b); err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// PreviewAndDownload will control the rendering of the preview and download page
func (c *CMD) PreviewAndDownload(w http.ResponseWriter, req *http.Request) {
	b, err := c.r.Do("dataset/finishpage", nil)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := w.Write(b); err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
