/*
Package querymanager provides and supports the QueryManager facility. The QueryManager provides a mechanism for
loading query templates from plain text files and allowing code to combine those templates with parameters to create a
query ready for execution.

The QueryManager is generic and is suitable for managing query templates for any data source that is interacted with via
text queries.
*/
package querymanager

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type QueryManager struct {
	TemplateLocation   string
	VarMatchRegEx      string
	FrameworkLogger    logging.Logger
	QueryIdPrefix      string
	TrimIdWhiteSpace   bool
	WrapStrings        bool
	StringWrapWith     string
	NewLine            string
	tokenisedTemplates map[string]*QueryTemplate
	state              ioc.ComponentState
}

func (qm *QueryManager) SubstituteMap(queryId string, params map[string]interface{}) (string, error) {

	template := qm.tokenisedTemplates[queryId]

	if template == nil {
		return "", errors.New("Unknown query " + queryId)
	}

	var b bytes.Buffer

	for _, token := range template.Tokens {

		if token.Type == Fragment {
			b.WriteString(token.Content)
		} else {

			paramValue := params[token.Content]

			if paramValue == nil {
				return "", errors.New(fmt.Sprintf("Query %s requires a parameter named %s but none supplied.", queryId, token.Content))
			}

			switch t := paramValue.(type) {
			default:
				return "", errors.New(fmt.Sprintf("Value for paramter %s is not a supported type.", token.Content))
			case string:
				b.WriteString(t)
			case int:
				b.WriteString(strconv.Itoa(t))
			}

		}

	}

	return b.String(), nil

}

func (qm *QueryManager) StartComponent() error {

	if qm.state != ioc.StoppedState {
		return nil
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

	} else {
		message := fmt.Sprintf("Unable to start QueryManager due to problem loading query files: %s", err.Error())
		return errors.New(message)
	}

}

func (qm *QueryManager) parseQueryFiles(files []string) map[string]*QueryTemplate {
	fl := qm.FrameworkLogger
	tokenisedTemplates := map[string]*QueryTemplate{}
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

func (qm *QueryManager) scanAndParse(scanner *bufio.Scanner, tokenisedTemplates map[string]*QueryTemplate, re *regexp.Regexp) {

	var currentTemplate *QueryTemplate = nil
	var fragmentBuffer bytes.Buffer

	for scanner.Scan() {
		line := scanner.Text()

		idLine, id := qm.isIdLine(line)

		if idLine {

			if currentTemplate != nil {
				currentTemplate.Finalise()
			}

			currentTemplate = NewQueryTemplate(id, &fragmentBuffer)
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
	} else {
		return y
	}
}

func (qm *QueryManager) addVar(token string, currentTemplate *QueryTemplate) {

	index, err := strconv.Atoi(token)

	if err == nil {
		currentTemplate.AddIndexedVar(index)
	} else {
		currentTemplate.AddLabelledVar(token)
	}
}

func (qm *QueryManager) isIdLine(line string) (bool, string) {
	idPrefix := qm.QueryIdPrefix

	if strings.HasPrefix(line, idPrefix) {
		newId := strings.TrimPrefix(line, idPrefix)

		if qm.TrimIdWhiteSpace {
			newId = strings.TrimSpace(newId)
		}

		return true, newId

	} else {
		return false, ""
	}
}

func (qm *QueryManager) isBlankLine(line string) bool {
	return len(strings.TrimSpace(line)) == 0
}

type QueryTokenType int

const (
	Fragment = iota
	VarName
	VarIndex
)

type QueryTemplate struct {
	Tokens         []*QueryTemplateToken
	Id             string
	currentToken   *QueryTemplateToken
	fragmentBuffer *bytes.Buffer
}

func (qt *QueryTemplate) Finalise() {
	qt.closeFragmentToken()
	qt.fragmentBuffer = nil
}

func (qt *QueryTemplate) AddFragmentContent(fragment string) {

	t := qt.currentToken

	if t == nil || t.Type != Fragment {
		t = NewQueryTemplateToken(Fragment)
		qt.Tokens = append(qt.Tokens, t)
		qt.currentToken = t
	}

	qt.fragmentBuffer.WriteString(fragment)
}

func (qt *QueryTemplate) closeFragmentToken() {

	t := qt.currentToken
	if t.Type == Fragment {
		t.Content = qt.fragmentBuffer.String()
		qt.fragmentBuffer.Reset()
	}

}

func (qt *QueryTemplate) AddIndexedVar(index int) {

	qt.closeFragmentToken()
	t := qt.currentToken

	t = NewQueryTemplateToken(VarIndex)
	t.Index = index

	qt.Tokens = append(qt.Tokens, t)
	qt.currentToken = t
}

func (qt *QueryTemplate) AddLabelledVar(label string) {

	qt.closeFragmentToken()
	t := qt.currentToken

	t = NewQueryTemplateToken(VarName)
	t.Content = label

	qt.Tokens = append(qt.Tokens, t)
	qt.currentToken = t
}

func (qt *QueryTemplate) EndLine() {
	qt.AddFragmentContent("\n")
}

func NewQueryTemplate(id string, buffer *bytes.Buffer) *QueryTemplate {
	t := new(QueryTemplate)
	t.Id = id
	t.currentToken = nil
	t.fragmentBuffer = buffer

	return t
}

type QueryTemplateToken struct {
	Type    QueryTokenType
	Content string
	Index   int
}

func NewQueryTemplateToken(tokenType QueryTokenType) *QueryTemplateToken {
	token := new(QueryTemplateToken)
	token.Type = tokenType

	return token
}

func (qtt *QueryTemplateToken) GetContent() string {
	return qtt.Content
}

func (qtt *QueryTemplateToken) String() string {

	switch qtt.Type {

	case Fragment:
		return qtt.Content
	case VarName:
		return fmt.Sprintf("VN:%s", qtt.Content)
	case VarIndex:
		return fmt.Sprintf("VI:%d", qtt.Index)
	default:
		return ""

	}
}
