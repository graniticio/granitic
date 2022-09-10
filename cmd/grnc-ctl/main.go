// Copyright 2016-2022 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

/*
The grnc-ctl tool, used to control running instances of Granitic applications.

Any Granitic application that enables the RuntimeCtl facility can be interacted with at runtime using this tool. Each
type of action that can be performed is referred to as a 'command'. Any component implementing the ctl.Commmand interface
will be automatically made available to grnc-ctl.

A number of built-in commands are available to perform common management tasks such as shutting down or suspending an application
or modifying the log-level of an individual component.

Run

	grnc-ctl help

to see a list of all commands, then

	grnct-ctl help command-name

to see usage and detailed help on an individual command.

Usage of grnc-ctl:

	grnc-ctl [options] command [qualifiers] [command_args]

	--port, --p  The port on which the application is listening for control messages (default 9099).
	--host, --h  The host on which the application is running (default localhost).

Built-in commands:

	components    Show a list of the names of components managed by the IoC container.
	global-level  Views or sets the global logging threshold for application or framework components.
	help          Show a list of all available commands or show help on a specific command.
	log-level     Views or sets a specific logging threshold for application or framework components.
	resume        Resumes one component or all components that have previously been suspended.
	shutdown      Stops all components then exits the application.
	start         Starts one component or all components.
	stop          Stops one component or all components.
	suspend       Suspends one component or all components.
*/
package main

import (
// "fmt"
// "github.com/graniticio/granitic/v3/initiation"
// "github.com/graniticio/granitic/v3/ioc"
// "github.com/graniticio/granitic/v3/initiation"
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
	termWidth        = 120
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

	if co.RenderHint == "COLUMNS" {
		columnOutput(co)
	} else {
		paragraphOutput(co)
	}
}

func columnOutput(co *commandOutcome) {

	tWidth := termWidth
	indent := 2
	minWidth := 20

	if co.OutputHeader != "" {
		fmt.Printf("\n%s\n\n", co.OutputHeader)
	} else {
		fmt.Println()
	}

	if co.OutputBody == nil {
		return
	}

	firstColMax := maxWidth(co.OutputBody, 0)
	firstColWidth := firstColMax + (indent * 2)
	secondColWidth := tWidth - firstColWidth

	if secondColWidth < minWidth {
		secondColWidth = minWidth
	}

	spaceCol := spaces(firstColWidth)

	for _, l := range co.OutputBody {

		ll := len(l)

		if ll > 0 {

			fmt.Printf("  %s  ", padRightTo(l[0], firstColMax))

			if ll == 1 {
				fmt.Printf("\n")
			}
		}

		if ll > 1 {
			sm := splitToMax(l[1], secondColWidth)

			sml := len(sm)

			if sml > 0 {
				fmt.Printf("%s\n", sm[0])
			}

			if sml > 1 {

				for _, m := range sm[1:] {
					fmt.Printf("%s%s\n", spaceCol, m)
				}

			}

		}

	}

	fmt.Println()

}

func spaces(w int) string {

	var b bytes.Buffer

	for i := 0; i < w; i++ {
		b.WriteString(" ")
	}

	return b.String()
}

func splitToMax(original string, max int) []string {

	result := make([]string, 0)

	var b bytes.Buffer

	split := strings.Split(original, " ")

	for i, s := range split {

		if i < (len(split) - 1) {
			s = s + " "
		}

		bs := b.Len()
		sl := len(s)

		if bs+sl > max {

			if bs == 0 {
				result = append(result, s)
			} else {
				result = append(result, b.String())
				b.Reset()
				b.WriteString(s)
			}

		} else {
			b.WriteString(s)
		}
	}

	if b.Len() > 0 {
		result = append(result, b.String())
	}

	return result
}

func maxWidth(c [][]string, i int) int {

	widest := 0

	for _, l := range c {

		if len(l) > i {
			cw := len(l[i])

			if cw > widest {
				widest = cw
			}

		}

	}

	return widest
}

func paragraphOutput(co *commandOutcome) {
	if co.OutputHeader != "" {
		fmt.Printf("\n%s\n", co.OutputHeader)

		if co.OutputBody == nil {
			fmt.Println()
		}
	}

	if co.OutputBody != nil {

		fmt.Println()

		for _, p := range co.OutputBody {

			for _, s := range p {

				cropped := splitToMax(s, termWidth-2)

				for _, c := range cropped {
					fmt.Printf("  %s\n", c)
				}

				fmt.Printf("\n")
			}

		}

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

	ta := newToolArgs()

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

func newToolArgs() *toolArgs {
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

func padRightTo(s string, p int) string {

	l := len(s)

	if l >= p {
		return s
	}

	var b bytes.Buffer

	b.WriteString(s)

	for i := l; i < p; i++ {
		b.WriteString(" ")
	}

	return b.String()
}
