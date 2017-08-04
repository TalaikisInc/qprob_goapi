package models

import "encoding/json"

type Meta struct {
	CatTotal  int
	PostTotal int
	TagTotal  int
}

type MetaJSON struct {
	CatTotal  int `json:"cat_total"`
	PostTotal int `json:"post_total"`
	TagTotal  int `json:"tag_total"`
}

func (c *Meta) MarshalJSON() ([]byte, error) {
	return json.Marshal(MetaJSON{
		c.CatTotal,
		c.PostTotal,
		c.TagTotal,
	})
}

func (c *Meta) UnmarshalJSON(b []byte) error {
	temp := &MetaJSON{}

	if err := json.Unmarshal(b, &temp); err != nil {
		return err
	}

	c.CatTotal = temp.CatTotal
	c.PostTotal = temp.PostTotal
	c.TagTotal = temp.TagTotal

	return nil
}
