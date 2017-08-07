package models

import "encoding/json"

type Sentiment struct {
	Date      string
	Sentiment float32
}

type SentimentJSON struct {
	Date      string  `json:"date, omitempty"`
	Sentiment float32 `json:"sentiment"`
}

func (c *Sentiment) MarshalJSON() ([]byte, error) {
	return json.Marshal(SentimentJSON{
		c.Date,
		c.Sentiment,
	})
}

func (c *Sentiment) UnmarshalJSON(b []byte) error {
	temp := &SentimentJSON{}

	if err := json.Unmarshal(b, &temp); err != nil {
		return err
	}

	c.Date = temp.Date
	c.Sentiment = temp.Sentiment

	return nil
}
