package redirect

import (
	"testing"

	"github.com/go-playground/validator"
)

func assertNumberOfErrors(t *testing.T, err error, expectedNumberOfErrors int) {
	if err == nil {
		if expectedNumberOfErrors != 0 {
			t.Errorf("Excpected %d error(s), got no error(s)", expectedNumberOfErrors)
		}
		return
	}

	validationErrors := err.(validator.ValidationErrors)
	validationErrorsLen := len(validationErrors)
	if validationErrorsLen != expectedNumberOfErrors {
		t.Errorf("Expected %d error(s), got %d", expectedNumberOfErrors, validationErrorsLen)

	}

}

func TestRedirectStructInitialization(t *testing.T) {

	var err error

	// Pretty much everything wrong
	_, err = NewRedirect("", -1, "", -1, -10)
	assertNumberOfErrors(t, err, 5)

	// Invalid DestIP
	_, err = NewRedirect("192.168.22.500", 30, "1.1.1.1", 70, 900)
	assertNumberOfErrors(t, err, 1)

	// Invalid Port bellow limits
	_, err = NewRedirect("192.168.22.1", 0, "1.1.1.1", 70, 900)
	assertNumberOfErrors(t, err, 1)

	// Invalid Port above limits
	_, err = NewRedirect("192.168.22.1", 9990889, "1.1.1.1", 70, 900)
	assertNumberOfErrors(t, err, 1)

	// Invalid TTs
	_, err = NewRedirect("192.168.22.1", 30, "1.1.1.1", 70, -200)
	assertNumberOfErrors(t, err, 1)

	// Valid options
	_, err = NewRedirect("192.168.22.1", 30, "1.1.1.1", 70, 200)
	assertNumberOfErrors(t, err, 0)

	_, err = NewRedirect("192.168.22.1", 30, "1.1.1.1", 70, -1)
	assertNumberOfErrors(t, err, 0)

}

func TestRuleAdding(t *testing.T) {
	// TODO
}

func TestRuleRemoval(t *testing.T) {
	// TODO
}
