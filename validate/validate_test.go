package validate

import (
	"fmt"
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/facility/jsonmerger"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/test"
	"testing"
)

func LoadTestConfig() *config.ConfigAccessor {

	cFile := test.TestFilePath("validate/validation.json")
	jsonMerger := new(jsonmerger.JsonMerger)
	jsonMerger.Logger = new(logging.ConsoleErrorLogger)

	mergedJson := jsonMerger.LoadAndMergeConfig([]string{cFile})

	return &config.ConfigAccessor{mergedJson, new(logging.ConsoleErrorLogger)}
}

func TestConfigParsing(t *testing.T) {

	ca := LoadTestConfig()

	test.ExpectBool(t, ca.PathExists("profileValidator"), true)

	rm := new(UnparsedRuleRuleManager)
	ca.Populate("ruleManager", rm)

	ov := new(ObjectValidator)
	ov.RuleManager = rm

	ca.Populate("profileValidator", ov)

	err := ov.StartComponent()

	test.ExpectNil(t, err)

	if err != nil {
		fmt.Println(err.Error())
	}

}
