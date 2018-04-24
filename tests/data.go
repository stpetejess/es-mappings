package tests

import "time"

// es-mappings:json
type FeedItem struct {
	ID string `es-mappings:"_id,text"`

	TestThing struct {
		A string `es-mappings:"a,keyword"`
		B Test   `es-mappings:"b,object"`
	} `es-mappings:",object"`
}

type Test struct {
	C string    `es-mappings:"c,keyword"`
	D time.Time `es-mappings:"d,date"`
}
