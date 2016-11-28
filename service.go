package main

type organisationContentService interface {
	getContentByOrganisationUUID(uuid string) ([]content, bool)
}

type simpleOrganisationContentService struct {
	articles map[string][]content
}

func newOrganisationContentService() simpleOrganisationContentService {
	articlesForOrg1 := []content{}
	articlesForOrg1 = append(articlesForOrg1, content{Title: "Test title"})

	m := make(map[string][]content)

	m["123"] = articlesForOrg1

	return simpleOrganisationContentService{articles: m}
}

func (ocs simpleOrganisationContentService) getContentByOrganisationUUID(uuid string) ([]content, bool) {

	contentForRequestedOrganisation, found := ocs.articles[uuid]
	return contentForRequestedOrganisation, found
}
