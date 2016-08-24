package validate

import (
	"fmt"
	"github.com/graniticio/granitic/config"
	"github.com/graniticio/granitic/facility/jsonmerger"
	"github.com/graniticio/granitic/ioc"
	"github.com/graniticio/granitic/logging"
	"github.com/graniticio/granitic/test"
	"github.com/graniticio/granitic/ws/nillable"
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

	ov, u := validatorAndUser(t)

	fe, err := ov.Validate(u)

	test.ExpectInt(t, len(fe), 0)

	test.ExpectNil(t, err)

	if err != nil {
		fmt.Println(err.Error())
	}

}

func validatorAndUser(t *testing.T) (*ObjectValidator, *User) {
	ca := LoadTestConfig()

	test.ExpectBool(t, ca.PathExists("profileValidator"), true)

	rm := new(UnparsedRuleRuleManager)
	ca.Populate("ruleManager", rm)

	ov := new(ObjectValidator)
	ov.RuleManager = rm
	ov.ComponentFinder = new(TestComponentFinder)

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

	u.Profile = p

	u.UserName = "Valid User"
	u.Role = nillable.NewNillableString("ADMIN")
	u.Password = "sadas*dasd1"
	u.Hint = " Sad "
	u.SecurityPhrase = "Is this your account?"
	p.Email = "email@example.com"
	p.Website = nillable.NewNillableString("  http://www.example.com ")

	return u
}

type User struct {
	UserName       string
	Role           *nillable.NillableString
	Password       string
	Hint           string
	SecurityPhrase string
	Profile        *Profile
}

type Profile struct {
	Email   string
	Website *nillable.NillableString
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
