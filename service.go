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

var descMap = map[string]string{
	"a14bcf4b-556d-31a6-8bbc-3d53d0366999": "Agropur cooperative processes and distributes dairy products. Its products include industrial cheese, yogurt, and products associated with fluid milk.",
}

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

	for i, story := range org.Stories {
		org.Stories[i] = ocs.enrichContent(story)
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

	recReadsStories := getContentFromRecommendedReads(uuid, ocs.recReadsURL)

	if len(recReadsStories) > 0 {
		org.RecommendedReadsStories = recReadsStories
	}

	return org, true, nil

}

func getContentFromRecommendedReads(uuid string, recReadsURL string) []content {
	desc, found := descMap[uuid]

	if !found {
		log.Printf("No description found for uuid=%s", uuid)
		return []content{}
	}

	log.Printf("Description=%s", desc)

	reqURL := fmt.Sprintf("%s/recommended-reads-api/recommend/contextual/doc?count=10&sort=rel&explain=false", recReadsURL)
	bodyString := fmt.Sprintf(`{ "doc": {"title": "This is the title", "content": "%s"} }`, desc)
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

	json.NewDecoder(resp.Body).Decode(&target)

	contList := []content{}

	for _, art := range target.Articles {
		cont := content{
			Title: art.Title,
			ID:    art.ID,
		}
		contList = append(contList, cont)
	}

	log.Printf("RecommendedReads stories=%s", contList)

	return contList

}

func (ocs simpleOrganisationContentService) enrichContent(story content) content {
	reqURL := fmt.Sprintf("http://api.ft.com/enrichedcontent/%s", story.ID)

	enriched := ocs.getEnrichedContent(reqURL)

	log.Printf("Standfirst=%s", enriched.Standfirst)

	story.Standfirst = enriched.Standfirst

	// get the image
	if enriched.MainImage.ID != "" {
		log.Printf("ImageSet ID=%s", enriched.MainImage.ID)
		imageSet := ocs.getEnrichedContent(enriched.MainImage.ID)

		members := imageSet.Members

		log.Printf("ImageSet ID=%s", members[0].ID)
		image := ocs.getEnrichedContent(members[0].ID)
		log.Printf("BinaryURL=%s", image.BinaryURL)
		story.ImageURL = image.BinaryURL
	}

	return story
}

func (ocs simpleOrganisationContentService) getEnrichedContent(reqURL string) enrichedContent {
	request, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		log.Printf("Could not create request for reqURL=%s, err=%s", reqURL, err)
	}
	request.Header.Set("X-Api-Key", ocs.apiKey)
	resp, err := httpClient.Do(request)
	defer resp.Body.Close()
	if err != nil {
		log.Printf("Error for reqURL=%s, err=%s", reqURL, err)
	}

	enriched := enrichedContent{}

	json.NewDecoder(resp.Body).Decode(&enriched)

	return enriched
}
