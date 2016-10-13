package error

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/types"
	"github.com/graniticio/granitic/ws"
	"strings"
)

type FrameworkServiceErrorFinder interface {
	UnmarshallError() *ws.CategorisedError
}

type ErrorCodeSource interface {
	ErrorCodesInUse() (codes types.StringSet, sourceName string)
	ValidateMissing() bool
}
type ServiceErrorManager struct {
	errors           map[string]*ws.CategorisedError
	FrameworkLogger  logging.Logger
	PanicOnMissing   bool
	errorCodeSources []ErrorCodeSource
	componentName    string
}

func (sem *ServiceErrorManager) ComponentName() string {
	return sem.componentName
}

func (sem *ServiceErrorManager) SetComponentName(name string) {
	sem.componentName = name
}

func (sem *ServiceErrorManager) Find(code string) *ws.CategorisedError {
	e := sem.errors[code]

	if e == nil {
		message := fmt.Sprintf("%s could not find error with code %s", sem.componentName, code)

		if sem.PanicOnMissing {
			panic(message)

		} else {
			sem.FrameworkLogger.LogWarnf(message)

		}

	}

	return e

}

func (sem *ServiceErrorManager) LoadErrors(definitions []interface{}) {

	l := sem.FrameworkLogger
	sem.errors = make(map[string]*ws.CategorisedError)

	for i, d := range definitions {

		e := d.([]interface{})

		category, err := ws.CodeToCategory(e[0].(string))

		if err != nil {
			l.LogWarnf("Error index %d: %s", i, err.Error())
			continue
		}

		code := e[1].(string)

		if len(strings.TrimSpace(code)) == 0 {
			l.LogWarnf("Error index %d: No code supplied", i)
			continue

		} else if sem.errors[code] != nil {
			l.LogWarnf("Error index %d: Duplicate code", i)
			continue
		}

		message := e[2].(string)

		if len(strings.TrimSpace(message)) == 0 {
			l.LogWarnf("Error index %d: No message supplied", i)
			continue
		}

		ce := ws.NewCategorisedError(category, code, message)

		sem.errors[code] = ce

	}
}

func (sem *ServiceErrorManager) RegisterCodeSource(ecs ErrorCodeSource) {
	if sem.errorCodeSources == nil {
		sem.errorCodeSources = make([]ErrorCodeSource, 0)
	}

	sem.errorCodeSources = append(sem.errorCodeSources, ecs)
}

func (sem *ServiceErrorManager) AllowAccess() error {

	failed := make(map[string][]string)

	for _, es := range sem.errorCodeSources {

		c, n := es.ErrorCodesInUse()

		for _, ec := range c.Contents() {

			if sem.errors[ec] == nil {
				addMissingCode(n, ec, failed)
			}

		}

	}

	if len(failed) > 0 {

		var m bytes.Buffer

		m.WriteString(fmt.Sprintf("Some components are using error codes that do not have a corresponding error message: \n"))

		for k, v := range failed {

			m.WriteString(fmt.Sprintf("%s: %q\n", k, v))
		}

		return errors.New(m.String())

	}

	return nil
}

func addMissingCode(source, code string, failed map[string][]string) {

	fs := failed[source]

	if fs == nil {
		fs = make([]string, 0)
	}

	fs = append(fs, code)

	failed[source] = fs

}
