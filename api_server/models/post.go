package models

import (
	"encoding/json"

	"github.com/xenu256/qprob_goapi/api_server/strip"
)

type Post struct {
	Title      string
	Slug       string
	URL        string
	Summary    string
	Content    string
	Date       string
	Sentiment  float32
	Image      string
	Wordcloud  string
	CategoryID Category
	Hits       int
	TotalPosts int
	Status     int
	Tags       []Tag
}

type PostJSON struct {
	Title      string   `json:"title, omitempty"`
	Slug       string   `json:"slug, omitempty"`
	URL        string   `json:"url, omitempty"`
	Summary    string   `json:"summary"`
	Content    string   `json:"content"`
	Date       string   `json:"date, omitempty"`
	Sentiment  float32  `json:"sentiment"`
	Image      string   `json:"image"`
	Wordcloud  string   `json:"wordcloud"`
	CategoryID Category `json:"category_id, omitempty"`
	Hits       int      `json:"hits"`
	TotalPosts int      `json:"total_posts"`
	Status     int      `json:"status"`
	Tags       []Tag    `json:"tags"`
}

func (p *Post) MarshalJSON() ([]byte, error) {
	return json.Marshal(PostJSON{
		p.Title,
		p.Slug,
		p.URL,
		strip.StripTags(p.Summary),
		p.Content,
		p.Date,
		p.Sentiment,
		p.Image,
		p.Wordcloud,
		p.CategoryID,
		p.Hits,
		p.TotalPosts,
		p.Status,
		p.Tags,
	})
}

func (p *Post) UnmarshalJSON(b []byte) error {
	temp := &PostJSON{}

	if err := json.Unmarshal(b, &temp); err != nil {
		return err
	}

	p.Title = temp.Title
	p.Slug = temp.Slug
	p.URL = temp.URL
	p.Summary = temp.Summary
	p.Content = temp.Content
	p.Date = temp.Date
	p.Sentiment = temp.Sentiment
	p.Image = temp.Image
	p.Wordcloud = temp.Wordcloud
	p.CategoryID = temp.CategoryID
	p.Hits = temp.Hits
	p.TotalPosts = temp.TotalPosts
	p.Status = temp.Status
	p.Tags = temp.Tags

	return nil
}
