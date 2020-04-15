package errors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/labstack/gommon/log"
)

const TemplatePath = "$GOPATH/src/github.com/cambridge-blockchain/emf/emf/context/errors/testdata/errors.yaml"

func getLogger() (l *log.Logger) {
	l = log.New("test")
	l.EnableColor()
	l.SetHeader(`[${time_rfc3339}] ${level} ${prefix} @ ${short_file}:${line} |`)
	l.SetLevel(log.DEBUG)
	return
}

func TestErrorHandlerNewError(t *testing.T) {
	eh := &EMFErrorHandlerType{
		DebugMode: false,
	}

	WithLogger(getLogger())(eh)
	WithTemplate(TemplatePath)(eh)

	err1 := eh.NewError("emf.400.QueryParameterInvalid", map[string]interface{}{
		"Error": fmt.Errorf("blah"),
	})
	cbErr1 := err1.(*EMFErrorType)
	err2 := eh.NewError("emf.400.QueryParameterInvalid", map[string]interface{}{
		"Error": fmt.Errorf("blah"),
	})

	if !errors.Is(err1, err2) {
		t.Fail()
	}

	err3 := eh.NewError("emf.400.TokenMissing", map[string]interface{}{
		"Error": fmt.Errorf("blah"),
	})
	cbErr3 := err3.(*EMFErrorType)

	if !errors.Is(err1, err2) {
		t.Log("FAIL: 400.QueryParameterInvalid Errors are NOT equal (errors.Is)")
		t.Fail()
	}
	if !errors.Is(cbErr1, err2) {
		t.Log("FAIL: 400.QueryParameterInvalid Errors are NOT equal (errors.Is)")
		t.Fail()
	}

	if errors.Is(err1, err3) {
		t.Log("FAIL: TokenMissing and QueryParameterInvalid Errors ARE equal (errors.Is)")
		t.Fail()
	}
	if errors.Is(err1, cbErr3) {
		t.Log("FAIL: TokenMissing and QueryParameterInvalid Errors ARE equal (errors.Is)")
		t.Fail()
	}
	if errors.Is(cbErr1, cbErr3) {
		t.Log("FAIL: TokenMissing and QueryParameterInvalid Errors ARE equal (errors.Is)")
		t.Fail()
	}

	errTokenMissing := ErrorType("emf.400.TokenMissing")
	if errors.Is(err1, errTokenMissing) {
		t.Log("400.QueryParameterInvalid Error is equal to ErrorType(emf.400.TokenMissing)")
		t.Fail()
	}
	if errors.Is(cbErr1, errTokenMissing) {
		t.Log("EMF 400.QueryParameterInvalid Error is equal to ErrorType(emf.400.TokenMissing)")
		t.Fail()
	}

	if !errors.Is(err3, errTokenMissing) {
		t.Log("400.TokenMissing Error is NOT equal to ErrorResponseCode(emf.400.TokenMissing)")
		t.Fail()
	}
	if !errors.Is(cbErr3, errTokenMissing) {
		t.Log("EMF 400.TokenMissing Error is NOT equal to ErrorResponseCode(emf.400.TokenMissing)")
		t.Fail()
	}

	errEMF := ErrorFromComponent("emf")
	if !errors.Is(err1, errEMF) {
		t.Log("emf.XXX.XXX Error #1 is NOT equal to ErrorFromComponent(emf)")
		t.Fail()
	}
	if !errors.Is(cbErr1, errEMF) {
		t.Log("EMF emf.XXX.XXX Error #1 is NOT equal to ErrorFromComponent(emf)")
		t.Fail()
	}

	if !errors.Is(err2, errEMF) {
		t.Log("emf.XXX.XXX Error #2 is NOT equal to ErrorFromComponent(emf)")
		t.Fail()
	}

	if !errors.Is(err3, errEMF) {
		t.Log("emf.XXX.XXX Error #3 is NOT equal to ErrorFromComponent(emf)")
		t.Fail()
	}
	if !errors.Is(cbErr3, errEMF) {
		t.Log("EMF emf.XXX.XXX Error #3 is NOT equal to ErrorFromComponent(emf)")
		t.Fail()
	}

	err400 := ErrorResponseCode(400)
	if !errors.Is(err1, err400) {
		t.Log("400 Error is NOT equal to ErrorResponseCode(400)")
		t.Fail()
	}
	if !errors.Is(cbErr1, err400) {
		t.Log("EMF 400 Error is NOT equal to ErrorResponseCode(400)")
		t.Fail()
	}

	if !errors.Is(err3, err400) {
		t.Log("400 Error is NOT equal to ErrorResponseCode(400)")
		t.Fail()
	}
	if !errors.Is(cbErr3, err400) {
		t.Log("EMF 400 Error is NOT equal to ErrorResponseCode(400)")
		t.Fail()
	}

	err401 := ErrorResponseCode(401)
	if errors.Is(err1, err401) || errors.Is(err3, err401) {
		t.Log("400 Error is equal to ErrorResponseCode(401)")
		t.Fail()
	}
	if errors.Is(cbErr1, err401) || errors.Is(err3, err401) {
		t.Log(" EMF 400 Error is equal to ErrorResponseCode(401)")
		t.Fail()
	}
}
