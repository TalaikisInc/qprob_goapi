package models

import "encoding/json"

type Category struct {
	Title     string
	Slug      string
	Thumbnail string
	PostCnt   int
	TotalCats int
}

type CategoryJSON struct {
	Title     string `json:"title, omitempty"`
	Slug      string `json:"slug, omitempty"`
	Thumbnail string `json:"thumbnail"`
	PostCnt   int    `json:"post_count"`
	TotalCats int    `json:"cat_total"`
}

func (c *Category) MarshalJSON() ([]byte, error) {
	return json.Marshal(CategoryJSON{
		c.Title,
		c.Slug,
		c.Thumbnail,
		c.PostCnt,
		c.TotalCats,
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
	c.TotalCats = temp.TotalCats

	return nil
}

type CategoryTopLevel struct {
	Total    int
	Category []Category
}

type CategoryTopLevelJSON struct {
	Total    int        `json:"total"`
	Category []Category `json:"categories"`
}

func (c *CategoryTopLevel) MarshalJSON() ([]byte, error) {
	return json.Marshal(CategoryTopLevelJSON{
		c.Total,
		c.Category,
	})
}

func (c *CategoryTopLevel) UnmarshalJSON(b []byte) error {
	temp := &CategoryTopLevelJSON{}

	if err := json.Unmarshal(b, &temp); err != nil {
		return err
	}

	c.Total = temp.Total
	c.Category = temp.Category

	return nil
}
