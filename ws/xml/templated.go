// Copyright 2016 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package xml

import (
	"context"
	"errors"
	"fmt"
	"github.com/graniticio/granitic/httpendpoint"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/ws"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"text/template"
)

// Serialises the body of a ws.WsResponse to XML using Go templates. See https://golang.org/pkg/text/template/
type TemplatedXMLResponseWriter struct {
	// Injected by the framework to allow this component to write log messages
	FrameworkLogger logging.Logger

	// A component able to calculate the correct HTTP status code to set for a response.
	StatusDeterminer ws.HttpStatusCodeDeterminer

	// Component able to generate errors if a problem is encountered during marshalling.
	FrameworkErrors *ws.FrameworkErrorGenerator

	// The common and static set of headers that should be written to all responses.
	DefaultHeaders map[string]string

	// The path (absolute or relative to application working directory) where unpopulated template files can be found.
	TemplateDir string

	// A map from an HTTP status code (e.g. '404') to the name of the template to be used to render that type of response.
	StatusTemplates map[string]string
	templates       *template.Template

	// Component able to dynamically generate additional headers to be written to the response.
	HeaderBuilder ws.WsCommonResponseHeaderBuilder

	//The name of a template to be used if the response has an abnormal (5xx) outcome.
	AbnormalTemplate string

	//The name of the default template to use if the response an error (400 or 409) outcome.
	ErrorTemplate string
	state         ioc.ComponentState
}

// See WsResponseWriter.Write
func (rw *TemplatedXMLResponseWriter) Write(ctx context.Context, state *ws.WsProcessState, outcome ws.WsOutcome) error {
	var ch map[string]string

	if rw.HeaderBuilder != nil {
		ch = rw.HeaderBuilder.BuildHeaders(ctx, state)
	}

	switch outcome {
	case ws.Normal:
		return rw.writeNormal(ctx, state.WsResponse, state.HttpResponseWriter, ch)
	case ws.Error:
		return rw.writeErrors(ctx, state.WsResponse, state.ServiceErrors, state.HttpResponseWriter, ch)
	case ws.Abnormal:
		return rw.writeAbnormalStatus(ctx, state.Status, state.HttpResponseWriter, ch)
	}

	return errors.New("Unsuported ws.WsOutcome value")
}

func (rw *TemplatedXMLResponseWriter) writeNormal(ctx context.Context, res *ws.WsResponse, w *httpendpoint.HttpResponseWriter, ch map[string]string) error {

	var t *template.Template
	var tn string

	if tn = res.Template; tn == "" {
		return errors.New("No template name set on response. Does your logic component implement ws.Templated?")
	}

	if t = rw.templates.Lookup(tn); t == nil {
		return errors.New("No such template " + tn)
	}

	return rw.write(ctx, res, w, ch, t)
}

func (rw *TemplatedXMLResponseWriter) write(ctx context.Context, res *ws.WsResponse, w *httpendpoint.HttpResponseWriter, ch map[string]string, t *template.Template) error {

	if w.DataSent {
		//This HTTP response has already been written to by another component - not safe to continue
		if rw.FrameworkLogger.IsLevelEnabled(logging.Debug) {
			rw.FrameworkLogger.LogDebugfCtx(ctx, "Response already written to.")
		}

		return nil
	}

	headers := ws.MergeHeaders(res, ch, rw.DefaultHeaders)
	ws.WriteHeaders(w, headers)

	s := rw.StatusDeterminer.DetermineCode(res)
	w.WriteHeader(s)

	e := res.Errors

	if res.Body == nil && !e.HasErrors() {
		return nil
	}

	return t.Execute(w, res)

}

func (rw *TemplatedXMLResponseWriter) writeErrors(ctx context.Context, res *ws.WsResponse, se *ws.ServiceErrors, w *httpendpoint.HttpResponseWriter, ch map[string]string) error {

	var t *template.Template
	var tn string

	res.Errors = se

	if res.Template != "" {
		tn = res.Template
	} else if rw.ErrorTemplate != "" {
		tn = rw.ErrorTemplate
	} else {
		tn = rw.AbnormalTemplate
	}

	if t = rw.templates.Lookup(tn); t == nil {
		return errors.New("No such template " + tn)
	}

	return rw.write(ctx, res, w, ch, t)
}

// See AbnormalStatusWriter.WriteAbnormalStatus
func (rw *TemplatedXMLResponseWriter) WriteAbnormalStatus(ctx context.Context, state *ws.WsProcessState) error {

	return rw.Write(ctx, state, ws.Abnormal)
}

func (rw *TemplatedXMLResponseWriter) writeAbnormalStatus(ctx context.Context, status int, w *httpendpoint.HttpResponseWriter, ch map[string]string) error {

	var t *template.Template
	var tn string

	if tn = rw.StatusTemplates[strconv.Itoa(status)]; tn == "" {
		tn = rw.AbnormalTemplate
	}

	if t = rw.templates.Lookup(tn); t == nil {
		return errors.New("No such template " + tn)
	}

	res := new(ws.WsResponse)
	res.HttpStatus = status
	var errors ws.ServiceErrors

	e := rw.FrameworkErrors.HttpError(status)
	errors.AddError(e)

	res.Errors = &errors

	return rw.write(ctx, res, w, ch, t)

}

// Called by the IoC container. Verifies that at minimum the AbnormalTemplate and TemplateDir fields are set.
// Parses all templates found in the TemplateDir
func (rw *TemplatedXMLResponseWriter) StartComponent() error {

	if rw.state != ioc.StoppedState {
		return nil
	}

	rw.state = ioc.StartingState

	if rw.AbnormalTemplate == "" {
		return errors.New("You must specify a template for abnormal HTTP statuses via the AbnormalTemplate field.")
	}

	if rw.TemplateDir == "" {
		return errors.New("You must specify a directory containing XML templates via the TemplateDir field.")
	}

	if rw.StatusTemplates == nil {
		rw.StatusTemplates = make(map[string]string)
	}

	if err := rw.preLoadTemplates(rw.TemplateDir); err != nil {
		return err
	}

	rw.state = ioc.RunningState

	return nil
}

func (rw *TemplatedXMLResponseWriter) preLoadTemplates(baseDir string) error {
	if tp, err := rw.templatePaths(rw.TemplateDir); err != nil {
		m := fmt.Sprintf("Problem converting template directory into a list of file paths %s: %s", baseDir, err)
		return errors.New(m)
	} else {

		if rw.templates, err = template.ParseFiles(tp...); err != nil {
			m := fmt.Sprintf("Problem parsing template files: %s", err)
			return errors.New(m)
		}

	}

	return nil
}

func (rw *TemplatedXMLResponseWriter) templatePaths(baseDir string) ([]string, error) {
	var di []os.FileInfo
	var err error

	tp := make([]string, 0)

	if di, err = ioutil.ReadDir(baseDir); err != nil {
		m := fmt.Sprintf("Problem opening template directory or sub-directory %s: %s", baseDir, err.Error())
		return nil, errors.New(m)
	}

	for _, f := range di {

		n := filepath.Join(baseDir, f.Name())

		if f.IsDir() {
			if a, err := rw.templatePaths(n); err == nil {
				tp = append(tp, a...)
			} else {
				return nil, err
			}
		} else {
			tp = append(tp, n)

		}

	}

	return tp, nil
}
