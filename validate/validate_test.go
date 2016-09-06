package validate

import (
	"fmt"
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/test"
	"github.com/graniticio/granitic/types"
	"testing"
)

func TestPathParsing(t *testing.T) {

	p := "a.b.c.d"

	s := determinePathFields(p)

	test.ExpectInt(t, s.Size(), 3)
	test.ExpectBool(t, s.Contains("a"), true)
	test.ExpectBool(t, s.Contains("a.b"), true)
	test.ExpectBool(t, s.Contains("a.b.c"), true)
	test.ExpectBool(t, s.Contains("a.b.c.d"), false)

	p = "a"

	s = determinePathFields(p)
	test.ExpectInt(t, s.Size(), 0)

}

func LoadTestConfig() *config.ConfigAccessor {

	cFile := test.TestFilePath("validate/validation.json")
	jsonMerger := new(config.JSONMerger)
	jsonMerger.Logger = new(logging.ConsoleErrorLogger)

	mergedJson, _ := jsonMerger.LoadAndMergeConfig([]string{cFile})

	return &config.ConfigAccessor{mergedJson, new(logging.ConsoleErrorLogger)}
}

func TestConfigParsing(t *testing.T) {

	ov, u := validatorAndUser(t)

	sc := new(SubjectContext)
	sc.Subject = u

	fe, err := ov.Validate(sc)

	for _, e := range fe {
		fmt.Printf("%s %q", e.Field, e.ErrorCodes)
	}

	test.ExpectInt(t, len(fe), 0)

	test.ExpectNil(t, err)

	if err != nil {
		fmt.Println(err.Error())
	}

}

func validatorAndUser(t *testing.T) (*RuleValidator, *User) {
	ca := LoadTestConfig()

	test.ExpectBool(t, ca.PathExists("profileValidator"), true)

	rm := new(UnparsedRuleManager)
	ca.Populate("ruleManager", rm)

	ov := new(RuleValidator)
	ov.RuleManager = rm
	ov.ComponentFinder = new(TestComponentFinder)
	ov.DefaultErrorCode = "DEFAULT"
	ov.Log = new(logging.ConsoleErrorLogger)

	ca.Populate("profileValidator", ov)

	err := ov.StartComponent()

	test.ExpectNil(t, err)

	if err != nil {
		fmt.Println(err.Error())
	}

	return ov, validUser()

}

func validUser() *User {
	u := new(User)
	p := new(Profile)
	pr := new(Preferences)

	u.Profile = p
	u.FailuresAllowed = 1
	u.UserName = "Valid User"
	u.Role = types.NewNilableString("ADMIN")
	u.Password = "sadas*dasd1"
	u.Hint = " Sad "
	u.SecurityPhrase = "Is this your account?"
	p.Email = "email@example.com"
	p.Website = types.NewNilableString("  http://www.example.com ")
	p.MarketTo = types.NewNilableBool(true)
	u.Prefs = pr
	pr.ResultsPer = types.NewNilableInt64(10)

	return u
}

type User struct {
	UserName        string
	Role            *types.NilableString
	Password        string
	Hint            string
	SecurityPhrase  string
	Profile         *Profile
	FailuresAllowed int8
	Prefs           *Preferences
	Salt            float64
}

type Profile struct {
	Email    string
	Website  *types.NilableString
	MarketTo *types.NilableBool
}

type Preferences struct {
	ResultsPer *types.NilableInt64
}

type TestComponentFinder struct {
}

func (cf *TestComponentFinder) ComponentByName(name string) *ioc.Component {

	if name == "Password" {
		return ioc.NewComponent(name, new(PasswordValidator))
	}

	return nil
}

type PasswordValidator struct {
}

func (pv *PasswordValidator) ValidString(p string) bool {

	if p == "password" {
		return false
	}

	return true
}

type ExtStringChecker struct {
}

func (ec *ExtStringChecker) ValidString(s string) bool {
	return s == "valid"
}

type ExtIntChecker struct {
}

func (ec *ExtIntChecker) ValidInt64(i int64) bool {

	return i == 64
}

type ExtFloatChecker struct {
}

func (ec *ExtFloatChecker) ValidFloat64(f float64) bool {

	return f == 64.21019
}

type CompFinder struct {
}

func (cf *CompFinder) ComponentByName(n string) *ioc.Component {

	if n == "extChecker" {
		return ioc.NewComponent(n, new(ExtStringChecker))
	} else if n == "extInt64Checker" {
		return ioc.NewComponent(n, new(ExtIntChecker))
	} else if n == "extFloat64Checker" {
		return ioc.NewComponent(n, new(ExtFloatChecker))
	} else {
		return ioc.NewComponent(n, new(types.NilableString))
	}

}
