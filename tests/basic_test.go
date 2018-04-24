package tests

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/bringhub/fabric/flatten"
)

type (
	TestCase struct {
		Actual   string
		Expected string
	}
)

var testCases []TestCase

func init() {
	var fimExpected, fimActual string
	var err error
	// load file
	if fimExpected, err = getAndProcess("./feed_item_mapping_expected.json"); err != nil {
		os.Exit(1)
	}
	fmt.Println("fimExpected: ", fimExpected)
	if fimActual, err = getAndProcess("./feed_item_mapping_actual.json"); err != nil {
		os.Exit(1)
	}
	fmt.Println("fimActual: ", fimActual)
	testCases = []TestCase{
		TestCase{Expected: fimExpected, Actual: fimActual},
	}
}

func getAndProcess(fname string) (string, error) {
	raw, err := ioutil.ReadFile(fname)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	temp := map[string]interface{}{}
	if err = json.Unmarshal(raw, &temp); err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	flat := flatten.Flatten(temp, false)
	if raw, err = json.Marshal(flat); err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	return string(raw), nil
}
func TestMap(t *testing.T) {
	for i, test := range testCases {
		if test.Actual != test.Expected {
			t.Errorf("[%d] TestMap(): got \n%v\n\t\t want \n%v", i, test.Actual, test.Expected)
		}
	}
}
