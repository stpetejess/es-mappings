package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/bringhub/fabric/flatten"
)

var FeedItemMapping string

func init() {
	// load file
	raw, err := ioutil.ReadFile("./feed_item_mapping_expected.json")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	temp := map[string]interface{}{}
	if err = json.Unmarshal(raw, &temp); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	flat := flatten.Flatten(temp, false)
	if raw, err = json.Marshal(flat); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	FeedItemMapping = string(raw)
}

// es-mapper:json
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
