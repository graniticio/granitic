// Copyright 2016-2022 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package dsquery

import (
	"bytes"
	"github.com/graniticio/granitic/v3/logging"
	"github.com/graniticio/granitic/v3/test"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

func TestArrayParamVars(t *testing.T) {

	qm := buildQueryManager()

	qm.TemplateLocation = test.FilePath("arrayparams")

	if err := qm.StartComponent(); err != nil {
		t.Fatalf("Unexpected %s", err.Error())
	}

	p := make(map[string]interface{})

	p["Table"] = "my_table"
	p["IDList"] = []int64{1, 2, 3}

	q, err := qm.BuildQueryFromID("SINGLE_QUERY_IN", p)

	if err != nil {
		t.Fatalf("Unexpected %s", err.Error())
	}

	if !strings.Contains(q, "(1, 2, 3)") {
		t.Fatalf("Unexpected resulting query \n%s", q)
	}

	p["IDList"] = []string{"1", "2", "3"}

	q, err = qm.BuildQueryFromID("SINGLE_QUERY_IN", p)

	if err != nil {
		t.Fatalf("Unexpected %s", err.Error())
	}

	if !strings.Contains(q, "('1', '2', '3')") {
		t.Fatalf("Unexpected resulting query \n%s", q)
	}

	p["IDList"] = []bool{true, false, true}

	q, err = qm.BuildQueryFromID("SINGLE_QUERY_IN", p)

	if err != nil {
		t.Fatalf("Unexpected %s", err.Error())
	}

	if !strings.Contains(q, "(true, false, true)") {
		t.Fatalf("Unexpected resulting query \n%s", q)
	}

}

func TestSingleSingleQueryNoVars(t *testing.T) {

	f := filepath.Join("querymanager", "single-query-no-vars")
	queryFiles := []string{test.FilePath(f)}
	qm := buildQueryManager()

	tt := qm.parseQueryFiles(queryFiles)

	members := len(tt)

	if members != 1 {
		t.Errorf("Expected one entry in tokens map, found %d", members)
	}

}

func TestSingleQueryIndexVars(t *testing.T) {

	f := filepath.Join("querymanager", "single-query-index-vars")
	queryFiles := []string{test.FilePath(f)}
	qm := buildQueryManager()

	tt := qm.parseQueryFiles(queryFiles)

	members := len(tt)

	if members != 1 {
		t.Errorf("Expected one entry in tokens map, found %d", members)
		t.FailNow()
	}

	tokenisedQuery := tt["SINGLE_QUERY_INDEX_VARS"]

	if tokenisedQuery == nil {
		t.Errorf("Expected a query named SINGLE_QUERY_INDEX_VARS")
	}

	stringQuery := ToString(tokenisedQuery.Tokens)

	refQuery := LoadRefFile("querymanager/single-query-index-vars-ref")

	if stringQuery != refQuery {
		t.Errorf("Generated query and reference query do not match. \nGEN:%s\nREF:%s\n", VisibleWhitespace(stringQuery), VisibleWhitespace(refQuery))
	}

}

func TestMultiQueryNameVars(t *testing.T) {

	f := filepath.Join("querymanager", "multi-query-name-vars")
	queryFiles := []string{test.FilePath(f)}
	qm := buildQueryManager()

	tt := qm.parseQueryFiles(queryFiles)

	members := len(tt)

	if members != 2 {
		t.Errorf("Expected two entries in tokens map, found %d", members)
		t.FailNow()
	}

	tokenisedQueryTwo := tt["MULTI_QUERY_INDEX_VARS_TWO"]

	if tokenisedQueryTwo == nil {
		t.Errorf("Expected a query named MULTI_QUERY_INDEX_VARS_TWO")
	}

	varsInQuery := NamedVarCount(tokenisedQueryTwo.Tokens)

	if varsInQuery != 3 {
		t.Errorf("Expected three named variables to be found in MULTI_QUERY_INDEX_VARS_TWO, found %d", varsInQuery)
	}

	stringQuery := ToString(tokenisedQueryTwo.Tokens)

	refQuery := LoadRefFile("querymanager/multi-query-name-vars-ref-2")

	if stringQuery != refQuery {
		t.Errorf("Generated query and reference query do not match. \nGEN:%s\nREF:%s\n", VisibleWhitespace(stringQuery), VisibleWhitespace(refQuery))
	}
}

func NamedVarCount(tokens []*queryTemplateToken) int {

	count := 0

	for _, token := range tokens {

		if token.Type == varNameToken {

			count++
		}

	}

	return count
}

func VisibleWhitespace(query string) string {

	nonewline := strings.Replace(query, "\n", "\\n", -1)

	return strings.Replace(nonewline, "\t", "\\t", -1)

}

func LoadRefFile(path string) string {

	f := test.FilePath(path)
	bytes, _ := ioutil.ReadFile(f)

	return string(bytes)
}

func ToString(tokens []*queryTemplateToken) string {

	var buffer bytes.Buffer

	for _, token := range tokens {
		buffer.WriteString(token.String())
	}

	return buffer.String()
}

func buildQueryManager() *TemplatedQueryManager {

	qm := new(TemplatedQueryManager)
	qm.QueryIDPrefix = "ID:"

	sp := new(SQLProcessor)

	qm.ValueProcessor = sp
	qm.TrimIDWhiteSpace = true
	qm.VarMatchRegEx = "\\$\\{([^\\}]*)\\}"
	qm.NewLine = "\n"
	qm.FrameworkLogger = new(logging.ConsoleErrorLogger)
	qm.ElementSeparator = ", "

	sp.BoolFalse = "false"
	sp.BoolTrue = "true"

	return qm

}
