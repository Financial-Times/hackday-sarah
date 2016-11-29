package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Financial-Times/neo-utils-go/neoutils"
	"github.com/jmcvetta/neoism"
)

type organisationContentService interface {
	getContentByOrganisationUUID(uuid string) (organisation, bool, error)
}

type simpleOrganisationContentService struct {
	conn        neoutils.NeoConnection
	recReadsURL string
	apiKey      string
}

func newOrganisationContentService(conn neoutils.NeoConnection, recReadsURL string, apiKey string) simpleOrganisationContentService {
	return simpleOrganisationContentService{conn, recReadsURL, apiKey}
}

func (ocs simpleOrganisationContentService) getContentByOrganisationUUID(uuid string) (organisation, bool, error) {

	results := []organisation{}

	now := time.Now()

	threeMonthsAgo := now.AddDate(0, -3, 0)

	secondsSinceEpoch := threeMonthsAgo.Unix()

	query := &neoism.CypherQuery{
		Statement: `
		MATCH (o:Organisation {uuid:{uuid}})
		OPTIONAL MATCH (o)--(i:IndustryClassification)
    OPTIONAL MATCH (o)-[:MENTIONS]-(c:Content)
    WHERE c.publishedDateEpoch > {secondsSinceEpoch}
    WITH o, i, {Title:c.title, ID:c.uuid, PublishedDate:c.publishedDate} as stories
    WITH o, i, collect(stories) as stories
    RETURN o.prefLabel as Title, i.prefLabel as IndustryClassification, stories as Stories, o.uuid as ID`,
		Parameters: neoism.Props{"uuid": uuid, "secondsSinceEpoch": secondsSinceEpoch},
		Result:     &results,
	}

	if err := ocs.conn.CypherBatch([]*neoism.CypherQuery{query}); err != nil {
		return organisation{}, false, err
	} else if len(results) == 0 {
		errMsg := fmt.Sprintf("No organisation found for uuid:%s", uuid)
		log.Print(errMsg)
		return organisation{}, false, nil
	}

	org := organisation{
		Title: results[0].Title,
		IndustryClassification: results[0].IndustryClassification,
		ID: results[0].ID,
	}

	if len(results[0].Stories) > 0 && results[0].Stories[0].ID != "" {
		org.Stories = results[0].Stories
	}

	subsidContent := []content{}

	subsidQuery := &neoism.CypherQuery{
		Statement: `
		MATCH (n:Organisation {uuid:{uuid}})-[:SUB_ORGANISATION_OF]-(s:Organisation)-[:MENTIONS]-(c:Content)
		WHERE c.publishedDateEpoch > {secondsSinceEpoch}
		WITH c, {Label:s.prefLabel} as Tags
		WITH c, collect(Tags) as Tags
		RETURN c.title as Title, c.uuid as ID, Tags as Tags, c.publishedDate as PublishedDate`,
		Parameters: neoism.Props{"uuid": uuid, "secondsSinceEpoch": secondsSinceEpoch},
		Result:     &subsidContent,
	}

	if err := ocs.conn.CypherBatch([]*neoism.CypherQuery{subsidQuery}); err != nil {
		return organisation{}, false, err
	}

	log.Printf("Subsids: %v", subsidContent)

	if len(subsidContent) > 0 {
		org.SubsidStories = subsidContent
	}

	if org.IndustryClassification != "" {
		indClassContent := []content{}

		indClassQuery := &neoism.CypherQuery{
			Statement: `MATCH (n:Organisation {uuid:{uuid}})--(i:IndustryClassification)--(comp:Organisation)-[m:MENTIONS]-(c:Content)
			WHERE c.publishedDateEpoch > {secondsSinceEpoch}
			WITH c, {Label:comp.prefLabel} as Tags
			WITH c, collect(Tags) as Tags
			RETURN DISTINCT c.title as Title, c.uuid as ID, Tags as Tags, c.publishedDate as PublishedDate
			ORDER BY PublishedDate DESC
			LIMIT(10)`,
			Parameters: neoism.Props{"uuid": uuid, "secondsSinceEpoch": secondsSinceEpoch},
			Result:     &indClassContent,
		}

		if err := ocs.conn.CypherBatch([]*neoism.CypherQuery{indClassQuery}); err != nil {
			return organisation{}, false, err
		}

		log.Printf("IndClass: %v", indClassContent)

		if len(indClassContent) > 0 {
			org.IndClassStories = indClassContent
		}
	}

	getContentFromRecommendedReads(uuid, ocs.recReadsURL)

	return org, true, nil

}

//curl -X POST --header 'Content-Type: application/json' --header 'Accept: application/json'
//-d '{ "doc": {"title": "This is the title", "content": "Ferrovial S.A., previously Grupo Ferrovial,
//is a Spanish multinational company involved in the design, construction, financing, operation and
//maintenance of transport, urban and services infrastructure"} }'
//'http://rr-recommendation-api-p-eu.ft.com:8080/recommended-reads-api/recommend/contextual/doc?count=10&sort=rel&explain=false'

func getContentFromRecommendedReads(uuid string, recReadsURL string) []content {
	reqURL := fmt.Sprintf("%s/recommended-reads-api/recommend/contextual/doc?count=10&sort=rel&explain=false", recReadsURL)
	bodyString := fmt.Sprintf(`{ "doc": {"title": "This is the title", "content": "%s"} }`, `Ferrovial S.A., previously Grupo Ferrovial, is a Spanish multinational company involved in the design, construction, financing, operation and maintenance of transport, urban and services infrastructure`)
	request, err := http.NewRequest("POST", reqURL, strings.NewReader(bodyString))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	if err != nil {
		log.Printf("Could not create request for reqURL=%s, err=%s", reqURL, err)
	}
	resp, err := httpClient.Do(request)
	defer resp.Body.Close()
	if err != nil {
		log.Printf("Error for reqURL=%s, err=%s", reqURL, err)
	}
	log.Printf("Response=%v", resp.StatusCode)
	if http.StatusOK != resp.StatusCode && http.StatusNotFound != resp.StatusCode {
		log.Printf("Unexpected status code for reqURL=%s, code=%v", reqURL, resp.StatusCode)
	}

	if err != nil {
		log.Printf("Error for reqURL=%s, err=%s", reqURL, err)
	}

	target := recommendedReads{}

	json.NewDecoder(resp.Body).Decode(target)

	return []content{}

}
