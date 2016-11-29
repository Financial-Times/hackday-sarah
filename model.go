package main

type content struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	Standfirst    string `json:"standfirst"`
	Byline        string `json:"byline"`
	PublishedDate string `json:"publishedDate"`
	ImageURL      string `json:"image"`
	Tags          []tag  `json:"tags"`
}
type organisation struct {
	ID                      string    `json:"id"`
	Title                   string    `json:"title"`
	Description             string    `json:"description"`
	IndustryClassification  string    `json:"industryClassification"`
	Stories                 []content `json:"stories"`
	SubsidStories           []content `json:"subsidiaryStories"`
	IndClassStories         []content `json:"industryClassificationStories"`
	RecommendedReadsStories []content `json:"recommendedReadsStories"`
}
type tag struct {
	URL   string `json:"url"`
	Label string `json:"label"`
}

// rec recReadsURL
type recommendedReads struct {
	Articles []article `json:"articles"`
	Status   string    `json:"status"`
	Type     string    `json:"type"`
	Version  string    `json:"version"`
}

type article struct {
	ID         string  `json:"id"`
	Popularity int     `json:"popularity"`
	Published  string  `json:"published"`
	Score      float64 `json:"score"`
	Title      string  `json:"title"`
}

type enrichedContent struct {
	AlternativeTitles struct {
		PromotionalTitle string `json:"promotionalTitle"`
	} `json:"alternativeTitles"`
	Annotations []struct {
		APIURL     string   `json:"apiUrl"`
		DirectType string   `json:"directType"`
		ID         string   `json:"id"`
		LeiCode    string   `json:"leiCode"`
		Predicate  string   `json:"predicate"`
		PrefLabel  string   `json:"prefLabel"`
		Type       string   `json:"type"`
		Types      []string `json:"types"`
	} `json:"annotations"`
	APIURL          string   `json:"apiUrl"`
	BodyXML         string   `json:"bodyXML"`
	Brands          []string `json:"brands"`
	Byline          string   `json:"byline"`
	CanBeSyndicated string   `json:"canBeSyndicated"`
	Comments        struct {
		Enabled bool `json:"enabled"`
	} `json:"comments"`
	ID            string    `json:"id"`
	MainImage     mainImage `json:"mainImage"`
	Members       []member  `json:"members"`
	BinaryURL     string    `json:"binaryUrl"`
	PrefLabel     string    `json:"prefLabel"`
	PublishedDate string    `json:"publishedDate"`
	RequestURL    string    `json:"requestUrl"`
	Standfirst    string    `json:"standfirst"`
	Standout      struct {
		EditorsChoice bool `json:"editorsChoice"`
		Exclusive     bool `json:"exclusive"`
		Scoop         bool `json:"scoop"`
	} `json:"standout"`
	Title  string   `json:"title"`
	Types  []string `json:"types"`
	WebURL string   `json:"webUrl"`
}

type member struct {
	ID string `json:"id"`
}

type mainImage struct {
	ID string `json:"id"`
}
