package xml

import (
	"github.com/graniticio/granitic/v3/logging"
	"github.com/graniticio/granitic/v3/test"
	"testing"
)

func TestAbnormalStatusWriting(t *testing.T) {

	xm := new(TemplatedXMLResponseWriter)

	xm.FrameworkLogger = new(logging.ConsoleErrorLogger)
	xm.TemplateDir = test.FilePath("xml-template")
	xm.AbnormalTemplate = "abnormal"

	err := xm.StartComponent()

	if err != nil {
		t.Errorf(err.Error())
	}

}
