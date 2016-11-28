package main

type content struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	Standfirst    string `json:"standfirst"`
	Byline        string `json:"byline"`
	PublishedDate string `json:"publishedDate"`
	ImageUrl      string `json:"image"`
	Tags          []tag  `json:"tags"`
}
type organisation struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Stories     []content `json:"stories"`
}
type tag struct {
	URL   string `json:"url"`
	Label string `json:"label"`
}
