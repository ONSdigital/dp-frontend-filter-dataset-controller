package filterapistubber

import (
	"io/ioutil"
	"net/http"

	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/server"
	"github.com/gorilla/mux"
)

func Start() {
	r := mux.NewRouter()

	bindAddr := ":22100"

	s := server.New(bindAddr, r)

	r.Path("/filters/{filterID}").HandlerFunc(getJob)
	r.Path("/filters/{filterID}/dimensions").HandlerFunc(getDimensions)
	r.Path("/filters/{filterID}/dimensions/time").HandlerFunc(getTime)
	r.Path("/filters/{filterID}/dimensions/time/options").HandlerFunc(getTimeOptions)
	r.Path("/filters/{filterID}/dimensions/CPI/options").HandlerFunc(getGoodsAndServices)
	r.Path("/filters/{filterID}/dimensions/goods-and-services/options").HandlerFunc(getGoodsAndServices)

	log.Debug("listening...", log.Data{
		"bind_address": bindAddr,
	})

	if err := s.ListenAndServe(); err != nil {
		log.Error(err, nil)
		return
	}

}

func getJob(w http.ResponseWriter, req *http.Request) {
	b, err := ioutil.ReadFile("stubbed-apis/filter-api-stubber/job.json")
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(500)
		return
	}

	w.Write(b)
}

func getTime(w http.ResponseWriter, req *http.Request) {
	b, err := ioutil.ReadFile("stubbed-apis/filter-api-stubber/time.json")
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(500)
		return
	}

	w.Write(b)
}

func getDimensions(w http.ResponseWriter, req *http.Request) {
	b, err := ioutil.ReadFile("stubbed-apis/filter-api-stubber/dimensions.json")
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(500)
		return
	}

	w.Write(b)
}

func getTimeOptions(w http.ResponseWriter, req *http.Request) {
	b, err := ioutil.ReadFile("stubbed-apis/filter-api-stubber/time-options.json")
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(500)
		return
	}

	w.Write(b)
}

func getGoodsAndServices(w http.ResponseWriter, req *http.Request) {
	b, err := ioutil.ReadFile("stubbed-apis/filter-api-stubber/goods-and-services-options.json")
	if err != nil {
		log.Error(err, nil)
		w.WriteHeader(500)
		return
	}

	w.Write(b)
}
