package config

import (
	"fmt"
	"github.com/graniticio/granitic/test"
	"testing"
)

func TestExpandToFilesAndURLs(t *testing.T) {

	p := test.TestFilePath("folders")
	u := "http://www.example.com/json"

	r, err := ExpandToFilesAndURLs([]string{u, p})

	test.ExpectNil(t, err)

	fmt.Printf("%v\n", r)

	test.ExpectInt(t, len(r), 6)

}
