package main

import (
	"fmt"
	"log"
	"time"

	"github.com/Financial-Times/neo-utils-go/neoutils"
	"github.com/jmcvetta/neoism"
)

type organisationContentService interface {
	getContentByOrganisationUUID(uuid string) (organisation, bool, error)
}

type simpleOrganisationContentService struct {
	conn neoutils.NeoConnection
}

func newOrganisationContentService(conn neoutils.NeoConnection) simpleOrganisationContentService {
	return simpleOrganisationContentService{conn}
}

func (ocs simpleOrganisationContentService) getContentByOrganisationUUID(uuid string) (organisation, bool, error) {

	results := []organisation{}

	now := time.Now()

	threeMonthsAgo := now.AddDate(0, -3, 0)

	secondsSinceEpoch := threeMonthsAgo.Unix()

	query := &neoism.CypherQuery{
		Statement: `
		MATCH (o:Organisation {uuid:{uuid}})--(i:IndustryClassification)
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

	return org, true, nil

}
