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
