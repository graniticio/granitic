package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSomething(t *testing.T) {

	_, err := findGoModCache()

	if !assert.Nil(t, err) {

		assert.Fail(t, err.Error())

	}

}
