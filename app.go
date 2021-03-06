package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/Financial-Times/neo-utils-go/neoutils"
	"github.com/gorilla/mux"
)

var httpClient = http.Client{
	Transport: &http.Transport{
		MaxIdleConnsPerHost: 128,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
	},
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	log.Printf("PORT=%s", port)

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	neo4jURL := os.Getenv("NEO4J_URL")

	log.Printf("neo4jURL=%s", neo4jURL)

	if neo4jURL == "" {
		log.Fatal("$NEO4J_URL must be set")
	}

	recReadsURL := os.Getenv("REC_READS_URL")

	log.Printf("recReadsURL=%s", recReadsURL)

	if recReadsURL == "" {
		log.Fatal("$REC_READS_URL must be set")
	}

	apiKey := os.Getenv("API_KEY")

	log.Printf("apiKey=%s", apiKey)

	if apiKey == "" {
		log.Fatal("$API_KEY must be set")
	}

	conf := neoutils.ConnectionConfig{
		BatchSize:     1024,
		Transactional: false,
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost: 100,
			},
			Timeout: 1 * time.Minute,
		},
		BackgroundConnect: true,
	}
	db, err := neoutils.Connect(neo4jURL, &conf)

	if err != nil {
		log.Fatalf("Error connecting to neo4j %s", err)
	}

	och := organisationContentHandler{newOrganisationContentService(db, recReadsURL, apiKey)}

	r := mux.NewRouter()
	r.HandleFunc("/organisations/{uuid}", och.getContentRelatedToOrganisation).Methods("GET")
	r.HandleFunc("/__gtg", och.goodToGo).Methods("GET")
	http.Handle("/", r)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}

type organisationContentHandler struct {
	ocs organisationContentService
}

func (och *organisationContentHandler) getContentRelatedToOrganisation(writer http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	uuid := vars["uuid"]

	contentForRequestedOrganisation, found, err := och.ocs.getContentByOrganisationUUID(uuid)

	if err != nil {
		writer.WriteHeader(http.StatusServiceUnavailable)
	}

	if !found {
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	enc := json.NewEncoder(writer)
	//should have error handling here
	enc.Encode(contentForRequestedOrganisation)
}

func (och *organisationContentHandler) goodToGo(writer http.ResponseWriter, req *http.Request) {
	writer.WriteHeader(http.StatusOK)
}
