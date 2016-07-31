package test

import "testing"

func ExpectString(t *testing.T, check, actual string) bool {
	if actual != check {
		t.Errorf("Expected %s, actual %s", check, actual)
		return false
	} else {
		return true
	}
}

func ExpectBool(t *testing.T, check, actual bool) bool {
	if actual != check {
		t.Errorf("Expected %t, actual %t", check, actual)
		return false
	} else {
		return true
	}
}

func ExpectInt(t *testing.T, check, actual int) bool {
	if actual != check {
		t.Errorf("Expected %d, actual %d", check, actual)
		return false
	} else {
		return true
	}
}

func ExpectFloat(t *testing.T, check, actual float64) bool {
	if actual != check {
		t.Errorf("Expected %e, actual %e", check, actual)
		return false
	} else {
		return true
	}
}

func ExpectNil(t *testing.T, check interface{}) bool {
	return check == nil
}

func ExpectNotNil(t *testing.T, check interface{}) bool {
	return check != nil
}
