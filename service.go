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
	"296db2c2-2c98-3e47-8e2a-e85bdfc1beae": "Ferrovial, S.A. , previously Grupo Ferrovial, is a Spanish multinational company involved in the design, construction, financing, operation (DBFO) and maintenance of transport, urban and services infrastructure.",
	"013f7fa7-aa26-3e20-84f1-fb8e5f7383ff": "Barclays is a British multinational banking and financial services company headquartered in London.",
}

var resultsMap = map[string]organisation{}

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

	org, found := resultsMap[uuid]

	if !found {
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

		org = organisation{
			Title: results[0].Title,
			IndustryClassification: results[0].IndustryClassification,
			ID: results[0].ID,
		}

		if len(results[0].Stories) > 0 && results[0].Stories[0].ID != "" {
			org.Stories = ocs.enrichContentList(results[0].Stories[0:5])
		}

		subsidContent := []content{}

		subsidQuery := &neoism.CypherQuery{
			Statement: `
			MATCH (n:Organisation {uuid:{uuid}})-[:SUB_ORGANISATION_OF]-(s:Organisation)-[:MENTIONS]-(c:Content)
			WHERE c.publishedDateEpoch > {secondsSinceEpoch}
			WITH c, {Label:s.prefLabel} as Tags
			WITH c, collect(Tags) as Tags
			RETURN c.title as Title, c.uuid as ID, Tags as Tags, c.publishedDate as PublishedDate
			LIMIT(5)`,
			Parameters: neoism.Props{"uuid": uuid, "secondsSinceEpoch": secondsSinceEpoch},
			Result:     &subsidContent,
		}

		if err := ocs.conn.CypherBatch([]*neoism.CypherQuery{subsidQuery}); err != nil {
			return organisation{}, false, err
		}

		log.Printf("Subsids: %v", subsidContent)

		if len(subsidContent) > 0 {
			org.SubsidStories = ocs.enrichContentList(subsidContent)
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
				LIMIT(5)`,
				Parameters: neoism.Props{"uuid": uuid, "secondsSinceEpoch": secondsSinceEpoch},
				Result:     &indClassContent,
			}

			if err := ocs.conn.CypherBatch([]*neoism.CypherQuery{indClassQuery}); err != nil {
				return organisation{}, false, err
			}

			log.Printf("IndClass: %v", indClassContent)

			if len(indClassContent) > 0 {
				org.IndClassStories = ocs.enrichContentList(indClassContent)
			}
		}

		recReadsStories := getContentFromRecommendedReads(uuid, ocs.recReadsURL)

		if len(recReadsStories) > 0 {
			org.RecommendedReadsStories = ocs.enrichContentList(recReadsStories)
		}

		resultsMap[uuid] = org
		log.Printf("Cached org %s", uuid)

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

	reqURL := fmt.Sprintf("%s/recommended-reads-api/recommend/contextual/doc?count=5&sort=rel&explain=false", recReadsURL)
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

func (ocs simpleOrganisationContentService) enrichContent(story content, index int, ch chan<- contentResult) {
	reqURL := fmt.Sprintf("http://api.ft.com/enrichedcontent/%s", story.ID)

	enriched := ocs.getEnrichedContent(reqURL)

	log.Printf("Standfirst=%s", enriched.Standfirst)

	story.Standfirst = enriched.Standfirst

	for _, annotation := range enriched.Annotations {
		if annotation.Type == "GENRE" && annotation.PrefLabel == "Comment" {
			commentTag := tag{URL: "https://www.ft.com/opinion", Label: "Comment"}
			story.Tags = append(story.Tags, commentTag)
		}
	}

	// get the image
	if enriched.MainImage.ID != "" {
		imageSet := ocs.getEnrichedContent(enriched.MainImage.ID)

		members := imageSet.Members

		image := ocs.getEnrichedContent(members[0].ID)
		story.ImageURL = image.BinaryURL
	}

	if story.ID == "ea207b7c-7020-3255-88d3-da429b6b8013" {
		story.ImageURL = "http://www.etbtravelnews.com/wp-content/uploads/2016/11/The-ultimate-milk-run-Qantas-1024x700.jpg"
	}

	ch <- contentResult{index, story}
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

func (ocs simpleOrganisationContentService) enrichContentList(storyList []content) []content {

	ch := make(chan contentResult)

	for i, story := range storyList {
		go ocs.enrichContent(story, i, ch)
	}

	for i := 0; i < len(storyList); i++ {
		result := <-ch
		storyList[result.index] = result.story
	}

	return storyList
}

type contentResult struct {
	index int
	story content
}
