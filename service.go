package main

type organisationContentService interface {
	getContentByOrganisationUUID(uuid string) (organisation, bool)
}

type simpleOrganisationContentService struct {
	organisations map[string]organisation
}

func newOrganisationContentService() simpleOrganisationContentService {
	articlesForOrg1 := []content{}
	articlesForOrg1 = append(articlesForOrg1, content{Title: "Test title"})

	Org1 := organisation{
		Stories: articlesForOrg1,
	}

	m := make(map[string]organisation)

	m["123"] = Org1

	return simpleOrganisationContentService{organisations: m}
}

func (ocs simpleOrganisationContentService) getContentByOrganisationUUID(uuid string) (organisation, bool) {

	contentForRequestedOrganisation, found := ocs.organisations[uuid]
	return contentForRequestedOrganisation, found
}
