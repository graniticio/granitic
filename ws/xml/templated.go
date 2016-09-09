package xml

import (
	"errors"
	"fmt"
	"github.com/graniticio/granitic/httpendpoint"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/ws"
	"golang.org/x/net/context"
	"io/ioutil"
	"os"
	"strconv"
	"text/template"
)

type TemplatedXMLResponseWriter struct {
	FrameworkLogger  logging.Logger
	StatusDeterminer ws.HttpStatusCodeDeterminer
	FrameworkErrors  *ws.FrameworkErrorGenerator
	DefaultHeaders   map[string]string
	TemplateDir      string
	StatusTemplates  map[string]string
	templates        *template.Template
	HeaderBuilder    ws.WsCommonResponseHeaderBuilder
	AbnormalTemplate string
	ErrorTemplate    string
}

func (rw *TemplatedXMLResponseWriter) Write(ctx context.Context, state *ws.WsProcessState, outcome ws.WsOutcome) error {
	var ch map[string]string

	if rw.HeaderBuilder != nil {
		ch = rw.HeaderBuilder.BuildHeaders(ctx, state)
	}

	switch outcome {
	case ws.Normal:
		return rw.writeNormal(ctx, state.WsResponse, state.HTTPResponseWriter, ch)
	case ws.Error:
		return rw.writeErrors(ctx, state.WsResponse, state.ServiceErrors, state.HTTPResponseWriter, ch)
	case ws.Abnormal:
		return rw.writeAbnormalStatus(ctx, state.Status, state.HTTPResponseWriter, ch)
	}

	return errors.New("Unsuported ws.WsOutcome value")
}

func (rw *TemplatedXMLResponseWriter) writeNormal(ctx context.Context, res *ws.WsResponse, w *httpendpoint.HTTPResponseWriter, ch map[string]string) error {

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

func (rw *TemplatedXMLResponseWriter) write(ctx context.Context, res *ws.WsResponse, w *httpendpoint.HTTPResponseWriter, ch map[string]string, t *template.Template) error {

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

func (rw *TemplatedXMLResponseWriter) writeErrors(ctx context.Context, res *ws.WsResponse, se *ws.ServiceErrors, w *httpendpoint.HTTPResponseWriter, ch map[string]string) error {

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

func (rw *TemplatedXMLResponseWriter) WriteAbnormalStatus(ctx context.Context, state *ws.WsProcessState) error {

	return rw.Write(ctx, state, ws.Abnormal)
}

func (rw *TemplatedXMLResponseWriter) writeAbnormalStatus(ctx context.Context, status int, w *httpendpoint.HTTPResponseWriter, ch map[string]string) error {

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

func (rw *TemplatedXMLResponseWriter) StartComponent() error {

	if rw.AbnormalTemplate == "" {
		return errors.New("You must specify a template for abnormal HTTP statuses via the AbnormalTemplate field.")
	}

	if rw.TemplateDir == "" {
		return errors.New("You must specify a directory containing XML templates via the TemplateDir field.")
	}

	if rw.StatusTemplates == nil {
		rw.StatusTemplates = make(map[string]string)
	}

	rw.preLoadTemplates(rw.TemplateDir)

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

		n := baseDir + "/" + f.Name()

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
