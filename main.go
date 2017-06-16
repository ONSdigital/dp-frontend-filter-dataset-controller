package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/ONSdigital/dp-frontend-dataset-controller/config"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/server"
	"github.com/gorilla/mux"
)

func main() {
	log.Namespace = "dp-frontend-dataset-controller"
	cfg := config.Get()

	r := mux.NewRouter()

	r.Path("/dataset/cmd").Methods("GET").HandlerFunc(loadingPage)
	r.Path("/dataset/cmd/middle").Methods("POST").HandlerFunc(middlePage)
	r.Path("/dataset/cmd/{jobId}").Methods("GET").HandlerFunc(middlePageGet)
	r.Path("/dataset/cmd/{jobId}/finish").Methods("GET").HandlerFunc(finishPage)

	s := server.New(cfg.BindAddr, r)

	if err := s.ListenAndServe(); err != nil {
		log.Error(err, nil)
		return
	}

}

func loadingPage(w http.ResponseWriter, r *http.Request) {
	req, err := http.NewRequest("POST", "http://localhost:20010/dataset/startpage", bytes.NewBufferString(`{}`))
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(b)
}

func middlePage(w http.ResponseWriter, r *http.Request) {
	jobId := rand.Intn(100)
	jid := strconv.Itoa(jobId)
	log.Debug("jobId", log.Data{"jobId": "/dataset/cmd/" + jid})
	http.Redirect(w, r, "/dataset/cmd/"+jid, 301)
}

func middlePageGet(w http.ResponseWriter, r *http.Request) {
	page := make(map[string]interface{})
	data := make(map[string]interface{})
	vars := mux.Vars(r)
	data["job_id"] = vars["jobId"]
	page["data"] = data

	jsonb, err := json.Marshal(page)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	req, err := http.NewRequest("POST", "http://localhost:20010/dataset/middlepage", bytes.NewBuffer(jsonb))
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(b)
}

func finishPage(w http.ResponseWriter, r *http.Request) {
	req, err := http.NewRequest("POST", "http://localhost:20010/dataset/finishpage", bytes.NewBufferString(`{}`))
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(b)
}
