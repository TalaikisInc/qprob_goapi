package models

import "encoding/json"

type Category struct {
	Title string
	Slug  string
}

type CategoryJSON struct {
	Title string `json:"title, omitempty"`
	Slug  string `json:"slug, omitempty"`
}

func (c *Category) MarshalJSON() ([]byte, error) {
	return json.Marshal(CategoryJSON{
		c.Title,
		c.Slug,
	})
}

func (c *Category) UnmarshalJSON(b []byte) error {
	temp := &CategoryJSON{}

	if err := json.Unmarshal(b, &temp); err != nil {
		return err
	}

	c.Title = temp.Title
	c.Slug = temp.Slug

	return nil
}
