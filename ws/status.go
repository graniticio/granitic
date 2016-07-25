package ws

import "net/http"

type HttpStatusCodeDeterminer interface {
	DetermineCode(response *WsResponse) int
	DetermineCodeFromErrors(errors *ServiceErrors) int
}

type DefaultHttpStatusCodeDeterminer struct {
}

func (dhscd *DefaultHttpStatusCodeDeterminer) DetermineCode(response *WsResponse) int {
	if response.HttpStatus != 0 {
		return response.HttpStatus

	} else {
		return http.StatusOK
	}
}

func (dhscd *DefaultHttpStatusCodeDeterminer) DetermineCodeFromErrors(errors *ServiceErrors) int {

	if errors.HttpStatus != 0 {
		return errors.HttpStatus
	}

	sCount := 0
	cCount := 0
	lCount := 0

	for _, error := range errors.Errors {

		switch error.Category {
		case Unexpected:
			return http.StatusInternalServerError
		case Security:
			sCount += 1
		case Logic:
			lCount += 1
		case Client:
			cCount += 1
		}
	}

	if sCount > 0 {
		return http.StatusUnauthorized
	}

	if cCount > 0 {
		return http.StatusBadRequest
	}

	if lCount > 0 {
		return http.StatusConflict
	}

	return http.StatusOK
}
