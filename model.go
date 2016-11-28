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
	ID                     string    `json:"id"`
	Title                  string    `json:"title"`
	Description            string    `json:"description"`
	IndustryClassification string    `json:"industryClassification"`
	Stories                []content `json:"stories"`
	SubsidStories          []content `json:"subsidiaryStories"`
	IndClassStories        []content `json:"industryClassificationStories"`
}
type tag struct {
	URL   string `json:"url"`
	Label string `json:"label"`
}
