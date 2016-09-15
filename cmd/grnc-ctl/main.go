/*
The grnc-ctl tool, used to control running instances of Granitic.
*/
package main

import (
//"fmt"
//"github.com/graniticio/granitic/initiation"
//"github.com/graniticio/granitic/ioc"
//"github.com/graniticio/granitic/initiation"
)
import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const (
	toolArgPrefix    = "--"
	commandArgPrefix = "-"
	defaultHost      = "localhost"
	defaultPort      = 9099
	minPort          = 1
	maxPort          = 65535
)

func main() {

	cr, ta := parseArgs()
	runCommand(ta, cr)

}

func parseArgs() (*ctlCommandRequest, *toolArgs) {
	args := os.Args

	al := len(args)

	if al <= 1 {
		usageExit()
	}

	ta, r := extractToolArgs(args[1:])
	ca, r := extractCommandArgs(r)

	if len(r) == 0 {
		exitError("No command specified")
	}

	cr := new(ctlCommandRequest)
	cr.Command = r[0]
	cr.Arguments = ca

	if len(r) > 1 {
		cr.Qualifiers = r[1:]
	}

	return cr, ta

}

func runCommand(ta *toolArgs, cr *ctlCommandRequest) {

	url := fmt.Sprintf("http://%s:%d/command", ta.Host, ta.Port)

	var b []byte
	var err error

	if b, err = json.Marshal(cr); err != nil {
		exitError("Problem creating web service request from tool arguments %s", err.Error())
	}

	var r *http.Response

	if r, err = http.Post(url, "application/json; charset=utf-8", bytes.NewReader(b)); err != nil {
		exitError("Problem executing web service call: %s", err.Error())
	}

	assessStatusCode(r.StatusCode)

	var res ctlResponse

	if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
		exitError("Unable to parse the response from the application")
	}

	processErrors(&res)
	renderOutput(&res)

}

func processErrors(res *ctlResponse) {

	if res.Errors == nil || (res.Errors.ByField == nil && res.Errors.General == nil) {
		return
	}

	messages := make([]string, 0)

	if res.Errors.ByField != nil {

		for _, es := range res.Errors.ByField {

			for _, e := range es {
				messages = append(messages, e.Message)
			}

		}

	}

	if res.Errors.General != nil {
		for _, e := range res.Errors.General {
			messages = append(messages, e.Message)
		}
	}

	printErrors(messages)

}

func renderOutput(res *ctlResponse) {

	co := res.Response

	if co == nil {
		return
	}

	if co.OutputHeader != "" {
		fmt.Println(co.OutputHeader)
	}

}

func assessStatusCode(s int) {
	switch s {
	case http.StatusServiceUnavailable:
		exitError("Server is busy")
	case http.StatusOK, http.StatusBadRequest, http.StatusConflict:
		return
	default:
		exitError("Unexpected HTTP %d response from server", s)
	}
}

func extractCommandArgs(args []string) (map[string]string, []string) {

	ca := make(map[string]string)

	remain := make([]string, 0)
	al := len(args)

	for i := 0; i < al; i++ {

		a := args[i]

		if strings.HasPrefix(a, commandArgPrefix) {

			k := strings.Replace(a, commandArgPrefix, "", -1)

			if i+1 < al {
				v := args[i+1]

				if strings.HasPrefix(v, commandArgPrefix) {
					exitError("%s is not a valid value for command argument %s", v, a)
				}

				if ca[k] != "" {
					exitError("Duplicate command argument %s", a)
				}

				ca[k] = v
				i++

				continue

			} else {
				exitError("Command argument %s does not have a value", a)
			}

		} else {
			remain = append(remain, a)
		}

	}

	if len(ca) == 0 {
		ca = nil
	}

	return ca, remain
}

func extractToolArgs(args []string) (*toolArgs, []string) {

	ta := NewToolArgs()

	remain := make([]string, 0)
	al := len(args)

	for i := 0; i < al; i++ {

		a := args[i]

		if strings.HasPrefix(a, toolArgPrefix) {

			if isHelp(a) {
				usageExit()
			} else if isHost(a) {

				if i+1 < al {
					i++
					ta.Host = args[i]
					continue

				} else {
					exitError("Host option specified with no value.")
				}
			} else if isPort(a) {

				if i+1 < al {
					i++
					p, err := strconv.Atoi(args[i])

					if err != nil || p < minPort || p > maxPort {
						exitError("Port option is invalid. Allowed ports are in the range %d-%d", minPort, maxPort)
					}

					ta.Port = p

					continue

				} else {
					exitError("Port option specified with no value.")
				}

			} else {
				exitError("Unsupported option %s", a)
			}

		} else {
			remain = append(remain, a)
		}

	}

	return ta, remain
}

func isHelp(a string) bool {
	return a == "--help"
}

func isPort(a string) bool {
	return a == "--port" || a == "--p"
}

func isHost(a string) bool {
	return a == "--host" || a == "--h"
}

func usageExit() {

	tabPrint("\nIssues commands to a running instance of a Granitic application\n", 0)
	tabPrint("usage: grnc-ctl [options] command [qualifiers] [command_args]", 0)
	tabPrint("options:", 1)
	tabPrint("--help       Prints this usage message", 2)
	tabPrint("--port, --p  The port on which the application is listening for control messages (default 9099).", 2)
	tabPrint("--host, --h  The host on which the application is running (default localhost).\n", 2)

	exitNormal()

}

type ctlCommandRequest struct {
	Command    string
	Qualifiers []string
	Arguments  map[string]string
}

func NewToolArgs() *toolArgs {
	ta := new(toolArgs)

	ta.Host = defaultHost
	ta.Port = defaultPort

	return ta
}

type toolArgs struct {
	Host string
	Port int
}

func tabPrint(s string, t int) {

	for i := 0; i < t; i++ {
		s = "  " + s
	}

	fmt.Println(s)
}

func printErrors(messages []string) {

	for _, message := range messages {
		fmt.Printf("grnc-ctl: %s\n", message)
	}
	os.Exit(1)
}

func exitError(message string, a ...interface{}) {

	m := "grnc-ctl: " + message + "\n"

	fmt.Printf(m, a...)
	os.Exit(1)
}

func exitNormal() {
	os.Exit(0)
}

type ctlResponse struct {
	Errors   *ctlErrors
	Response *commandOutcome
}

type ctlErrors struct {
	ByField map[string][]errorWrapper
	General []errorWrapper
}

type errorWrapper struct {
	Code    string
	Message string
}

type commandOutcome struct {
	OutputHeader string
	OutputBody   [][]string
	RenderHint   string
}
