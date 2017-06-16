package handlers

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/ONSdigital/dp-frontend-dataset-controller/renderer"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
)

// Landing handles the controller functionality for the landing page
func Landing(w http.ResponseWriter, req *http.Request) {
	r := renderer.New()

	b, err := r.Do("dataset/startpage", nil)
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
func CreateJobID(w http.ResponseWriter, req *http.Request) {
	// TODO: This is a stubbed job id - replace with real job id from backend once
	// code has been written
	jobID := rand.Intn(10000000)
	jid := strconv.Itoa(jobID)

	log.Trace("created job id", log.Data{"job_id": jid})
	http.Redirect(w, req, "/dataset/cmd/"+jid, 301)
}

// Middle controls the rendering of a "middle" cmd page - this will be replaced
// by other handlers when further pages are defined by UX
func Middle(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	r := renderer.New()

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

	b, err := r.Do("dataset/middlepage", body)
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
func PreviewAndDownload(w http.ResponseWriter, req *http.Request) {
	r := renderer.New()

	b, err := r.Do("dataset/finishpage", nil)
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
