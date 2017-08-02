package v2handlers

import (
	"log"
	"strconv"
)

func StrToInt(s string) (string, int) {
	d, err := strconv.Atoi(s)
	if err != nil {
		log.Fatal(err)
	}
	return s, d
}
