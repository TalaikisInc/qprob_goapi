package models

import "encoding/json"

type Category struct {
	Title     string
	Slug      string
	Thumbnail string
	PostCnt   int
}

type CategoryJSON struct {
	Title     string `json:"title, omitempty"`
	Slug      string `json:"slug, omitempty"`
	Thumbnail string `json:"thumbnail"`
	PostCnt   int    `json:"post_count"`
}

func (c *Category) MarshalJSON() ([]byte, error) {
	return json.Marshal(CategoryJSON{
		c.Title,
		c.Slug,
		c.Thumbnail,
		c.PostCnt,
	})
}

func (c *Category) UnmarshalJSON(b []byte) error {
	temp := &CategoryJSON{}

	if err := json.Unmarshal(b, &temp); err != nil {
		return err
	}

	c.Title = temp.Title
	c.Slug = temp.Slug
	c.Thumbnail = temp.Thumbnail
	c.PostCnt = temp.PostCnt

	return nil
}
