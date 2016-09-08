package xml

import (
	"errors"
	"fmt"
	"github.com/graniticio/granitic/httpendpoint"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/ws"
	"golang.org/x/net/context"
	"net/http"
	"strconv"
	"text/template"
)

type StandardXMLResponseWriter struct {
	FrameworkLogger  logging.Logger
	StatusDeterminer ws.HttpStatusCodeDeterminer
	FrameworkErrors  *ws.FrameworkErrorGenerator
	DefaultHeaders   map[string]string
	TemplateDir      string
	StatusTemplates  map[string]string
	cachedTemplates  map[string]*template.Template
	HeaderBuilder    ws.WsCommonResponseHeaderBuilder
	CacheTemplates   bool
	AbnormalTemplate string
}

func (rw *StandardXMLResponseWriter) Write(ctx context.Context, state *ws.WsProcessState, outcome ws.WsOutcome) error {
	var ch map[string]string

	if rw.HeaderBuilder != nil {
		ch = rw.HeaderBuilder.BuildHeaders(ctx, state)
	}

	switch outcome {
	case ws.Normal:
		return rw.writeNormal(ctx, state.WsResponse, state.HTTPResponseWriter, ch)
	/*case ws.Error:
	return rw.writeErrors(ctx, state.ServiceErrors, state.HTTPResponseWriter, ch)*/
	case ws.Abnormal:
		return rw.writeAbnormalStatus(ctx, state.Status, state.HTTPResponseWriter, ch)
	}

	return errors.New("Unsuported ws.WsOutcome value")
}

func (rw *StandardXMLResponseWriter) writeNormal(ctx context.Context, res *ws.WsResponse, w *httpendpoint.HTTPResponseWriter, ch map[string]string) error {

	var t *template.Template
	var tn string
	var err error

	if tn = res.Template; tn == "" {
		return errors.New("No template name set on response. Does your logic component implement ws.Templated?")
	}

	if t, err = rw.loadTemplate(tn); err != nil {
		m := fmt.Sprintf("Problem loading XML template %s: %s", tn, err.Error())
		return errors.New(m)
	}

	return rw.write(ctx, res, w, ch, t)
}

func (rw *StandardXMLResponseWriter) write(ctx context.Context, res *ws.WsResponse, w *httpendpoint.HTTPResponseWriter, ch map[string]string, t *template.Template) error {

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

func (rw *StandardXMLResponseWriter) WriteAbnormalStatus(ctx context.Context, state *ws.WsProcessState) error {

	return rw.Write(ctx, state, ws.Abnormal)
}

func (rw *StandardXMLResponseWriter) writeAbnormalStatus(ctx context.Context, status int, w *httpendpoint.HTTPResponseWriter, ch map[string]string) error {

	var t *template.Template
	var tn string
	var err error

	fmt.Printf("Handling %d\n", status)

	if tn = rw.StatusTemplates[strconv.Itoa(status)]; tn == "" {
		tn = rw.AbnormalTemplate
	}

	if t, err = rw.loadTemplate(tn); err != nil {
		m := fmt.Sprintf("Problem loading XML template %s: %s", tn, err.Error())
		return errors.New(m)
	}

	res := new(ws.WsResponse)
	res.HttpStatus = status
	var errors ws.ServiceErrors

	e := rw.FrameworkErrors.HttpError(status)
	errors.AddError(e)

	res.Errors = &errors

	return rw.write(ctx, res, w, ch, t)

}

func (rw *StandardXMLResponseWriter) loadTemplate(n string) (*template.Template, error) {

	if t := rw.cachedTemplates[n]; rw.CacheTemplates && t != nil {
		return t, nil
	}

	if t, err := template.ParseFiles(rw.TemplateDir + n); err != nil {
		return nil, err
	} else {

		if rw.CacheTemplates {
			rw.cachedTemplates[n] = t
		}

		return t, nil
	}
}

func (rw *StandardXMLResponseWriter) StartComponent() error {
	rw.cachedTemplates = make(map[string]*template.Template)

	if rw.StatusTemplates == nil {
		rw.StatusTemplates = make(map[string]string)
	}

	return nil
}

type StandardXmlUnmarshaller struct {
}

func (um *StandardXmlUnmarshaller) Unmarshall(ctx context.Context, req *http.Request, wsReq *ws.WsRequest) error {

	return nil
}
