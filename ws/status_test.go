package ws

import "testing"

func TestCodeMapping(t *testing.T) {
	scd := NewGraniticHTTPStatusCodeDeterminer()

	r := new(Response)
	r.HTTPStatus = 200

	h := scd.DetermineCode(r)

	if h != 200 {
		t.Fail()
	}
}

func TestDefaultCodeBehaviour(t *testing.T) {
	scd := NewGraniticHTTPStatusCodeDeterminer()

	r := new(Response)

	se := new(ServiceErrors)
	r.Errors = se

	h := scd.DetermineCode(r)

	if h != 200 {
		t.Fail()
	}

	he := new(CategorisedError)
	he.Category = HTTP
	he.Code = "999"

	se.AddNewError(Unexpected, "", "")
	se.AddError(he)
	se.AddNewError(Security, "", "")
	se.AddNewError(Client, "", "")
	se.AddNewError(Logic, "", "")

	h = scd.DetermineCode(r)

	if h != 500 {
		t.Fail()

	}

	se = new(ServiceErrors)
	r.Errors = se

	se.AddError(he)
	se.AddNewError(Security, "", "")
	se.AddNewError(Client, "", "")
	se.AddNewError(Logic, "", "")

	h = scd.DetermineCode(r)

	if h != 999 {
		t.Fail()

	}

	se = new(ServiceErrors)
	r.Errors = se

	se.AddNewError(Security, "", "")
	se.AddNewError(Client, "", "")
	se.AddNewError(Logic, "", "")

	h = scd.DetermineCode(r)

	if h != 401 {
		t.Fail()

	}

	se = new(ServiceErrors)
	r.Errors = se

	se.AddNewError(Client, "", "")
	se.AddNewError(Logic, "", "")

	h = scd.DetermineCode(r)

	if h != 400 {
		t.Fail()

	}

	se = new(ServiceErrors)
	r.Errors = se

	se.AddNewError(Logic, "", "")

	h = scd.DetermineCode(r)

	if h != 409 {
		t.Fail()

	}

	se = new(ServiceErrors)
	r.Errors = se

	se.AddError(he)
	se.AddNewError(Security, "", "")
	se.AddNewError(Client, "", "")
	se.AddNewError(Logic, "", "")

	r.HTTPStatus = 777

	h = scd.DetermineCode(r)

	if h != 777 {
		t.Fail()

	}

	se = new(ServiceErrors)
	r.Errors = se

	se.AddError(he)
	se.AddNewError(Security, "", "")
	se.AddNewError(Client, "", "")
	se.AddNewError(Logic, "", "")

	r.HTTPStatus = 0
	se.HTTPStatus = 888

	h = scd.DetermineCode(r)

	if h != 888 {
		t.Fail()

	}
}
