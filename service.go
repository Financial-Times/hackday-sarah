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

	oneMonthAgo := now.AddDate(0, -1, 0)

	secondsSinceEpoch := oneMonthAgo.Unix()

	query := &neoism.CypherQuery{
		Statement: `
		MATCH (o:Organisation {uuid:{uuid}})
    OPTIONAL MATCH (o)-[:MENTIONS]-(c:Content)
    WHERE c.publishedDateEpoch > {secondsSinceEpoch}
    WITH o, {Title:c.title, ID:c.uuid} as stories
    WITH o, collect(stories) as stories
    RETURN o.prefLabel as Title, stories as Stories, o.uuid as ID`,
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

	return results[0], true, nil

}
