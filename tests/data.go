package tests

// es-mappings:json
type FeedItem struct {
	ID string `es-mapping:"_id,text"`

	TestThing struct {
		A string `es-mapping:"a,keyword"`
		B Test   `es-mapping:"b,object"`
	} `es-mapping:",object"`
}

type Test struct {
	C string `es-mapping:"c,keyword"`
}
