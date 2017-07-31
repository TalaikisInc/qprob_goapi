package models

import "encoding/json"

type Tag struct {
	Title   string
	Slug    string
	PostCnt int
}

type TagJSON struct {
	Title   string `json:"title, omitempty"`
	Slug    string `json:"slug, omitempty"`
	PostCnt int    `json:"post_count"`
}

func (c *Tag) MarshalJSON() ([]byte, error) {
	return json.Marshal(TagJSON{
		c.Title,
		c.Slug,
		c.PostCnt,
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

	return nil
}
