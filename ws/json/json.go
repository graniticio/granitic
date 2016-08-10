package json

import (
	"encoding/json"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/ws"
	"net/http"
)

type DefaultJsonUnmarshaller struct {
	FrameworkLogger logging.Logger
}

func (jdu *DefaultJsonUnmarshaller) Unmarshall(req *http.Request, wsReq *ws.WsRequest) error {

	err := json.NewDecoder(req.Body).Decode(&wsReq.RequestBody)

	return err

}

type DefaultJsonResponseWriter struct {
	FrameworkLogger logging.Logger
}

func (djrw *DefaultJsonResponseWriter) Write(res *ws.WsResponse, w http.ResponseWriter) error {

	if res.Body == nil {
		return nil
	}

	wrapper := wrapJsonResponse(nil, res.Body)

	data, err := json.Marshal(wrapper)

	if err != nil {
		return err
	}

	_, err = w.Write(data)

	return err
}

type DefaultAbnormalResponseWriter struct {
	FrameworkLogger logging.Logger
}

func (djerw *DefaultAbnormalResponseWriter) WriteWithErrors(status int, errors *ws.ServiceErrors, w http.ResponseWriter) error {

	w.WriteHeader(status)
	wrapper := wrapJsonResponse(djerw.formatErrors(errors), nil)

	data, err := json.Marshal(wrapper)

	if err != nil {
		return err
	}

	_, err = w.Write(data)

	return err
}

func (djerw *DefaultAbnormalResponseWriter) formatErrors(errors *ws.ServiceErrors) interface{} {

	f := make(map[string][]string)

	for _, error := range errors.Errors {

		var c string

		switch error.Category {
		default:
			c = "?"
		case ws.Unexpected:
			c = "U"
		case ws.Security:
			c = "S"
		case ws.Logic:
			c = "L"
		case ws.Client:
			c = "C"
		case ws.HTTP:
			c = "H"
		}

		k := c + "-" + error.Label
		a := f[k]

		f[k] = append(a, error.Message)
	}

	return f
}

func wrapJsonResponse(errors interface{}, body interface{}) interface{} {
	f := make(map[string]interface{})

	if errors != nil {
		f["Errors"] = errors
	}

	if body != nil {
		f["Response"] = body
	}

	return f
}

type JsonWrapper struct {
	Response interface{}
	Errors   interface{}
}
