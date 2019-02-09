// Copyright 2016-2019 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

/*
Package dsquery provides mechanisms for managing templated queries to be executed against a data source.

The types in this package are agnostic to the type of data source being used (e.g. RDBMS, NoSQL, cache). Instead, these types are
concerned with reading templated queries from files and populating those templates with variables a runtime. The actual
execution of queries is the responsibility of clients that understand the type of data source in use (see the rdbms package).

Most Granitic applications requiring access to a data source will enable the QueryManager facility which provides access
to an instance of the TemplatedQueryManager type defined in this package. Instructions on configuring and using the
QueryManager facility can be found at http://granitic.io/1.0/ref/query-manager also see the package documentation for the
facility/querymanager package for some basic examples.

*/
package dsquery

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/graniticio/granitic/v2/config"
	"github.com/graniticio/granitic/v2/ioc"
	"github.com/graniticio/granitic/v2/logging"
	"github.com/graniticio/granitic/v2/types"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const requiredPrefix = "!"

// A QueryManager is a type that is able to populate a pre-defined template query given a set of named parameters
// and return a complete query ready for execution against some data source.
type QueryManager interface {
	// BuildQueryFromId finds a template with the supplied query ID and uses the supplied named parameters to populate
	// that template. Returns the populated query or an error if the template could not be found or there was a problem
	// populating the query.
	BuildQueryFromId(qid string, params map[string]interface{}) (string, error)

	// FragmentFromId is used to recover a template which does not have any parameters to populate (a fragment). This is most commonly
	// used when code needs to dynamically construct a query from several fragments and templates. Returns the fragment
	// or an error if the fragment could not be found.
	FragmentFromId(qid string) (string, error)
}

// NewTemplatedQueryManager creates a new, empty TemplatedQueryManager.
func NewTemplatedQueryManager() *TemplatedQueryManager {
	qm := new(TemplatedQueryManager)
	qm.fragments = make(map[string]string)

	return qm
}

// An implementation of QueryManager that reads files containing template queries, tokenizes them and populates them
// on demand with maps of named parameters. This is the implementation provided by the QueryManager facility. See
// http://granitic.io/1.0/ref/query-manager for details.
type TemplatedQueryManager struct {

	// The path to a folder where template files are stored.
	TemplateLocation string

	// A regular expression that allows variable names to be recognised in queries.
	VarMatchRegEx string

	// Logger used by Granitic framework components. Automatically injected.
	FrameworkLogger logging.Logger

	// Lines in a template file starting with this string are considered to indicate the start of a new query template. The remainder
	// of the line will be used as the ID of that query template.
	QueryIdPrefix string

	// Whether or not query IDs should have leading and trailing whitespace removed.
	TrimIdWhiteSpace bool

	// A component able to handle missing parameter values and the escaping of supplied parameters
	ValueProcessor ParamValueProcessor

	// Whether or not a stock ParamValueProcessor should be injected into this component (set to false if defining your own)
	CreateDefaultValueProcessor bool

	// The character sequence that indicates a new line in a template file (e.g. \n)
	NewLine            string
	tokenisedTemplates map[string]*queryTemplate
	fragments          map[string]string
	state              ioc.ComponentState
}

// See QueryManager.FragmentFromId
func (qm *TemplatedQueryManager) FragmentFromId(qid string) (string, error) {

	f := qm.fragments[qid]

	if f != "" {
		return f, nil
	}

	p := make(map[string]interface{})

	f, err := qm.BuildQueryFromId(qid, p)

	if err != nil {
		qm.fragments[qid] = f
	}

	return f, err

}

// See QueryManager.BuildQueryFromId
func (qm *TemplatedQueryManager) BuildQueryFromId(qid string, params map[string]interface{}) (string, error) {
	template := qm.tokenisedTemplates[qid]

	if template == nil {
		return "", errors.New("Unknown query " + qid)
	}

	return qm.buildQueryFromTemplate(qid, template, params)
}

func (qm *TemplatedQueryManager) buildQueryFromTemplate(qid string, template *queryTemplate, params map[string]interface{}) (string, error) {

	var b bytes.Buffer

	vp := qm.ValueProcessor
	log := qm.FrameworkLogger
	trace := log.IsLevelEnabled(logging.Trace)

	for _, token := range template.Tokens {

		if token.Type == fragmentToken {
			b.WriteString(token.Content)
		} else {

			key := token.Content

			if trace {
				log.LogTracef("Processing parameter %s", key)
			}

			required := strings.HasPrefix(key, requiredPrefix)

			if required {
				key = strings.Replace(key, requiredPrefix, "", 1)
			}

			paramValue := params[key]

			vc := ParamValueContext{
				Value:   paramValue,
				Key:     key,
				QueryId: qid,
			}

			if paramValue == nil {

				if trace {
					log.LogTracef("Parameter %s is unset", key)
				}

				if required {
					return "", errors.New(fmt.Sprintf("Parameter %s is required for query %s but has not been set", key, qid))
				}

				if err := vp.SubstituteUnset(&vc); err != nil {

					//ValueProcessor does not allow this parameter to be unset
					return "", err
				}

			}

			//Perform any required escaping on the parameter value
			vp.EscapeParamValue(&vc)

			switch t := vc.Value.(type) {
			default:
				return "", errors.New(fmt.Sprintf("TemplatedQueryManager: Value for parameter %s is not a supported type. (type is %T)", key, t))
			case string:
				b.WriteString(t)
			case *types.NilableString:
				b.WriteString(t.String())
			case types.NilableString:
				b.WriteString(t.String())
			case int:
				b.WriteString(strconv.Itoa(t))
			case int64:
				b.WriteString(strconv.FormatInt(t, 10))
			case *types.NilableInt64:
				b.WriteString(strconv.FormatInt(t.Int64(), 10))
			case types.NilableInt64:
				b.WriteString(strconv.FormatInt(t.Int64(), 10))
			}

		}

	}

	q := b.String()

	if qm.FrameworkLogger.IsLevelEnabled(logging.Debug) {
		qm.FrameworkLogger.LogDebugf("\n" + q)
	}

	return q, nil

}

// StartComponent is called by the IoC container. Loads, parses and tokenizes query templates. Returns an error
// if there was a problem loading, parsing or tokenizing.
func (qm *TemplatedQueryManager) StartComponent() error {

	if qm.state != ioc.StoppedState {
		return nil
	}

	if qm.ValueProcessor == nil {
		m := fmt.Sprintf("No ValueProcessor available for QueryManager. If you have set QueryManager.CreateDefaultValueProcessor to false you must define a component that implements ParamValueProcessor")
		return errors.New(m)
	}

	qm.state = ioc.StartingState

	fl := qm.FrameworkLogger
	fl.LogDebugf("Starting QueryManager")
	fl.LogDebugf(qm.TemplateLocation)

	queryFiles, err := config.FileListFromPath(qm.TemplateLocation)

	if err == nil {

		qm.tokenisedTemplates = qm.parseQueryFiles(queryFiles)
		fl.LogDebugf("Started QueryManager with %d queries", len(qm.tokenisedTemplates))

		qm.state = ioc.RunningState

		return nil

	}

	return fmt.Errorf("Unable to start QueryManager due to problem loading query files: %s", err.Error())

}

func (qm *TemplatedQueryManager) parseQueryFiles(files []string) map[string]*queryTemplate {
	fl := qm.FrameworkLogger
	tokenisedTemplates := map[string]*queryTemplate{}
	re := regexp.MustCompile(qm.VarMatchRegEx)

	for _, filePath := range files {

		fl.LogDebugf("Parsing query file %s", filePath)

		file, err := os.Open(filePath)

		if err != nil {
			fl.LogErrorf("Unable to open %s for parsing: %s", filePath, err.Error())
			continue
		}

		defer file.Close()

		scanner := bufio.NewScanner(file)
		qm.scanAndParse(scanner, tokenisedTemplates, re)
	}

	return tokenisedTemplates
}

func (qm *TemplatedQueryManager) scanAndParse(scanner *bufio.Scanner, tokenisedTemplates map[string]*queryTemplate, re *regexp.Regexp) {

	var currentTemplate *queryTemplate = nil
	var fragmentBuffer bytes.Buffer

	for scanner.Scan() {
		line := scanner.Text()

		idLine, id := qm.isIdLine(line)

		if idLine {

			if currentTemplate != nil {
				currentTemplate.Finalise()
			}

			currentTemplate = newQueryTemplate(id, &fragmentBuffer)
			tokenisedTemplates[id] = currentTemplate
			continue
		}

		if qm.isBlankLine(line) {
			continue
		}

		varTokens := re.FindAllStringSubmatch(line, -1)

		if varTokens == nil {
			currentTemplate.AddFragmentContent(line)
		} else {

			fragments := re.Split(line, -1)

			firstMatch := re.FindStringIndex(line)

			startsWithVar := (firstMatch[0] == 0)
			varCount := len(varTokens)
			fragmentCount := len(fragments)

			maxCount := intMax(varCount, fragmentCount)

			for i := 0; i < maxCount; i++ {

				varAvailable := i < varCount
				fragAvailable := i < fragmentCount

				if varAvailable && fragAvailable {

					varToken := varTokens[i][1]
					fragment := fragments[i]

					if startsWithVar {
						qm.addVar(varToken, currentTemplate)
						currentTemplate.AddFragmentContent(fragment)
					} else {
						currentTemplate.AddFragmentContent(fragment)
						qm.addVar(varToken, currentTemplate)

					}

				} else if varAvailable {
					qm.addVar(varTokens[i][1], currentTemplate)

				} else if fragAvailable {
					currentTemplate.AddFragmentContent(fragments[i])
				}

			}
		}

		currentTemplate.EndLine()

	}

	if currentTemplate != nil {
		currentTemplate.Finalise()
	}

}

func intMax(x, y int) int {
	if x > y {
		return x
	}

	return y
}

func (qm *TemplatedQueryManager) addVar(token string, currentTemplate *queryTemplate) {

	index, err := strconv.Atoi(token)

	if err == nil {
		currentTemplate.AddIndexedVar(index)
	} else {
		currentTemplate.AddLabelledVar(token)
	}
}

func (qm *TemplatedQueryManager) isIdLine(line string) (bool, string) {
	idPrefix := qm.QueryIdPrefix

	if strings.HasPrefix(line, idPrefix) {
		newId := strings.TrimPrefix(line, idPrefix)

		if qm.TrimIdWhiteSpace {
			newId = strings.TrimSpace(newId)
		}

		return true, newId

	}

	return false, ""

}

func (qm *TemplatedQueryManager) isBlankLine(line string) bool {
	return len(strings.TrimSpace(line)) == 0
}

type queryTokenType int

const (
	fragmentToken = iota
	varNameToken
	varIndexToken
)

type queryTemplate struct {
	Tokens         []*queryTemplateToken
	Id             string
	currentToken   *queryTemplateToken
	fragmentBuffer *bytes.Buffer
}

func (qt *queryTemplate) Finalise() {
	qt.closeFragmentToken()
	qt.fragmentBuffer = nil
}

func (qt *queryTemplate) AddFragmentContent(fragment string) {

	t := qt.currentToken

	if t == nil || t.Type != fragmentToken {
		t = NewQueryTemplateToken(fragmentToken)
		qt.Tokens = append(qt.Tokens, t)
		qt.currentToken = t
	}

	qt.fragmentBuffer.WriteString(fragment)
}

func (qt *queryTemplate) closeFragmentToken() {

	t := qt.currentToken
	if t != nil && t.Type == fragmentToken {
		t.Content = qt.fragmentBuffer.String()
		qt.fragmentBuffer.Reset()
	}

}

func (qt *queryTemplate) AddIndexedVar(index int) {

	qt.closeFragmentToken()
	t := qt.currentToken

	t = NewQueryTemplateToken(varIndexToken)
	t.Index = index

	qt.Tokens = append(qt.Tokens, t)
	qt.currentToken = t
}

func (qt *queryTemplate) AddLabelledVar(label string) {

	qt.closeFragmentToken()
	t := qt.currentToken

	t = NewQueryTemplateToken(varNameToken)
	t.Content = label

	qt.Tokens = append(qt.Tokens, t)
	qt.currentToken = t
}

func (qt *queryTemplate) EndLine() {
	qt.AddFragmentContent("\n")
}

func newQueryTemplate(id string, buffer *bytes.Buffer) *queryTemplate {
	t := new(queryTemplate)
	t.Id = id
	t.currentToken = nil
	t.fragmentBuffer = buffer

	return t
}

type queryTemplateToken struct {
	Type    queryTokenType
	Content string
	Index   int
}

func NewQueryTemplateToken(tokenType queryTokenType) *queryTemplateToken {
	token := new(queryTemplateToken)
	token.Type = tokenType

	return token
}

func (qtt *queryTemplateToken) GetContent() string {
	return qtt.Content
}

func (qtt *queryTemplateToken) String() string {

	switch qtt.Type {

	case fragmentToken:
		return qtt.Content
	case varNameToken:
		return fmt.Sprintf("VN:%s", qtt.Content)
	case varIndexToken:
		return fmt.Sprintf("VI:%d", qtt.Index)
	default:
		return ""

	}
}
