package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	och := organisationContentHandler{newOrganisationContentService()}

	r := mux.NewRouter()
	r.HandleFunc("/organisations/{uuid}", och.getContentRelatedToOrganisation).Methods("GET")
	http.Handle("/", r)

	log.Fatal(http.ListenAndServe("localhost:8000", nil))
}

type organisationContentHandler struct {
	ocs organisationContentService
}

func (och *organisationContentHandler) getContentRelatedToOrganisation(writer http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	uuid := vars["uuid"]

	contentForRequestedOrganisation, found := och.ocs.getContentByOrganisationUUID(uuid)
	if !found {
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	enc := json.NewEncoder(writer)
	//should have error handling here
	enc.Encode(contentForRequestedOrganisation)
}
