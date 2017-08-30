package models

import "encoding/json"

type Tag struct {
	Title     string
	Slug      string
	PostCnt   int
	TotalTags int
}

type TagJSON struct {
	Title     string `json:"title, omitempty"`
	Slug      string `json:"slug, omitempty"`
	PostCnt   int    `json:"post_count"`
	TotalTags int    `json:"total_tags"`
}

func (c *Tag) MarshalJSON() ([]byte, error) {
	return json.Marshal(TagJSON{
		c.Title,
		c.Slug,
		c.PostCnt,
		c.TotalTags,
	})
}

func (c *Tag) UnmarshalJSON(b []byte) error {
	temp := &TagJSON{}

	if err := json.Unmarshal(b, &temp); err != nil {
		return err
	}

	c.Title = temp.Title
	c.Slug = temp.Slug
	c.PostCnt = temp.PostCnt
	c.TotalTags = temp.TotalTags

	return nil
}
