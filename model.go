package main

type content struct {
	ID                string `json:"id"`
	Title             string `json:"title"`
	Standfirst        string `json:"standfirst"`
	Byline            string `json:"byline"`
	PublishedDate     string `json:"publishedDate"`
}
type organisation struct {
	Title string `json:"title"`
	Description string `json:"description"`
	Stories [] content `json:"stories"`
}
