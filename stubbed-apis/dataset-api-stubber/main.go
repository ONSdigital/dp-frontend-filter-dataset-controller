package datasetapistubber

import (
	"io/ioutil"
	"net/http"

	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/server"
	"github.com/gorilla/mux"
)

func Start() {
	r := mux.NewRouter()

	bindAddr := ":20012"

	s := server.New(bindAddr, r)

	r.Path("/datasets/{id}/editions/{edition}/versions/{version}").HandlerFunc(getDataset)

	log.Debug("listening...", log.Data{
		"bind_address": bindAddr,
	})

	if err := s.ListenAndServe(); err != nil {
		log.Error(err, nil)
		return
	}

}

func getDataset(w http.ResponseWriter, req *http.Request) {
	b, err := ioutil.ReadFile("stubbed-apis/dataset-api-stubber/dataset.json")
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(500)
		return
	}

	w.Write(b)
}
