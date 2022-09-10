// Copyright 2016-2022 Granitic. All rights reserved.
// Use of this source code is governed by an Apache 2.0 license that can be found in the LICENSE file at the root of this project.

package validate

import (
	"context"
	"fmt"
	"github.com/graniticio/granitic/v3/config"
	"github.com/graniticio/granitic/v3/ioc"
	"github.com/graniticio/granitic/v3/logging"
	"github.com/graniticio/granitic/v3/test"
	"github.com/graniticio/granitic/v3/types"
	"path/filepath"
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

func LoadTestConfig() *config.Accessor {

	cfp := filepath.Join("validate", "validation.json")

	cFile := test.FilePath(cfp)
	jsonMerger := config.NewJSONMergerWithDirectLogging(new(logging.ConsoleErrorLogger), new(config.JSONContentParser))

	mergedJSON, _ := jsonMerger.LoadAndMergeConfig([]string{cFile})

	return &config.Accessor{JSONData: mergedJSON, FrameworkLogger: new(logging.ConsoleErrorLogger)}
}

func TestConfigParsing(t *testing.T) {

	ov, u := validatorAndUser(t)

	sc := new(SubjectContext)
	sc.Subject = u

	fe, err := ov.Validate(context.Background(), sc)

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

func (cf *TestComponentFinder) AllComponents() []*ioc.Component {
	return []*ioc.Component{}
}

type PasswordValidator struct {
}

func (pv *PasswordValidator) ValidString(p string) (bool, error) {

	if p == "password" {
		return false, nil
	}

	return true, nil
}

type ExtStringChecker struct {
}

func (ec *ExtStringChecker) ValidString(s string) (bool, error) {
	return s == "valid", nil
}

type ExtIntChecker struct {
}

func (ec *ExtIntChecker) ValidInt64(i int64) (bool, error) {

	return i == 64, nil
}

type ExtFloatChecker struct {
}

func (ec *ExtFloatChecker) ValidFloat64(f float64) (bool, error) {

	return f == 64.21019, nil
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

func (cf *CompFinder) AllComponents() []*ioc.Component {
	return []*ioc.Component{}
}
