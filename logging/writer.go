package logging

import "fmt"

type LogWriter interface {
	WriteMessage(string)
}

type ConsoleWriter struct {
}

func (cw *ConsoleWriter) WriteMessage(m string) {
	fmt.Print(m)
}
