package forms_test

import (
	"encoding/json"
	"errors"
	"testing"

	userforms "github.com/pafthang/servicebase/services/user/forms"

	baseforms "github.com/pafthang/servicebase/services/base/forms"

	recordmodels "github.com/pafthang/servicebase/services/record/models"

	"github.com/pafthang/servicebase/tests"
	"github.com/pafthang/servicebase/tools/security"
)

func TestRecordVerificationConfirmValidateAndSubmit(t *testing.T) {
	t.Parallel()

	testApp, _ := tests.NewTestApp()
	defer testApp.Cleanup()

	authCollection, err := testApp.Dao().FindCollectionByNameOrId("users")
	if err != nil {
		t.Fatal(err)
	}

	scenarios := []struct {
		jsonData    string
		expectError bool
	}{
		// empty data (Validate call check)
		{
			`{}`,
			true,
		},
		// expired token (Validate call check)
		{
			`{"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjRxMXhsY2xtZmxva3UzMyIsImVtYWlsIjoidGVzdEBleGFtcGxlLmNvbSIsImNvbGxlY3Rpb25JZCI6Il9wYl91c2Vyc19hdXRoXyIsInR5cGUiOiJhdXRoUmVjb3JkIiwiZXhwIjoxNjQwOTkxNjYxfQ.Avbt9IP8sBisVz_2AGrlxLDvangVq4PhL2zqQVYLKlE"}`,
			true,
		},
		// valid token (already verified record)
		{
			`{"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6Im9hcDY0MGNvdDR5cnUycyIsImVtYWlsIjoidGVzdDJAZXhhbXBsZS5jb20iLCJjb2xsZWN0aW9uSWQiOiJfcGJfdXNlcnNfYXV0aF8iLCJ0eXBlIjoiYXV0aFJlY29yZCIsImV4cCI6MjIwODk4NTI2MX0.PsOABmYUzGbd088g8iIBL4-pf7DUZm0W5Ju6lL5JVRg"}`,
			false,
		},
		// valid token (unverified record)
		{
			`{"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjRxMXhsY2xtZmxva3UzMyIsImVtYWlsIjoidGVzdEBleGFtcGxlLmNvbSIsImNvbGxlY3Rpb25JZCI6Il9wYl91c2Vyc19hdXRoXyIsInR5cGUiOiJhdXRoUmVjb3JkIiwiZXhwIjoyMjA4OTg1MjYxfQ.hL16TVmStHFdHLc4a860bRqJ3sFfzjv0_NRNzwsvsrc"}`,
			false,
		},
	}

	for i, s := range scenarios {
		form := userforms.NewRecordVerificationConfirm(testApp, authCollection)

		// load data
		loadErr := json.Unmarshal([]byte(s.jsonData), form)
		if loadErr != nil {
			t.Errorf("(%d) Failed to load form data: %v", i, loadErr)
			continue
		}

		interceptorCalls := 0
		interceptor := func(next baseforms.InterceptorNextFunc[*recordmodels.Record]) baseforms.InterceptorNextFunc[*recordmodels.Record] {
			return func(r *recordmodels.Record) error {
				interceptorCalls++
				return next(r)
			}
		}

		record, err := form.Submit(interceptor)

		// check interceptor calls
		expectInterceptorCalls := 1
		if s.expectError {
			expectInterceptorCalls = 0
		}
		if interceptorCalls != expectInterceptorCalls {
			t.Errorf("[%d] Expected interceptor to be called %d, got %d", i, expectInterceptorCalls, interceptorCalls)
		}

		hasErr := err != nil
		if hasErr != s.expectError {
			t.Errorf("(%d) Expected hasErr to be %v, got %v (%v)", i, s.expectError, hasErr, err)
		}

		if hasErr {
			continue
		}

		claims, _ := security.ParseUnverifiedJWT(form.Token)
		tokenRecordId := claims["id"]

		if record.Id != tokenRecordId {
			t.Errorf("(%d) Expected record.Id %q, got %q", i, tokenRecordId, record.Id)
		}

		if !record.Verified() {
			t.Errorf("(%d) Expected record.Verified() to be true, got false", i)
		}
	}
}

func TestRecordVerificationConfirmInterceptors(t *testing.T) {
	t.Parallel()

	testApp, _ := tests.NewTestApp()
	defer testApp.Cleanup()

	authCollection, err := testApp.Dao().FindCollectionByNameOrId("users")
	if err != nil {
		t.Fatal(err)
	}

	authRecord, err := testApp.Dao().FindUserRecordByEmail("users", "test@example.com")
	if err != nil {
		t.Fatal(err)
	}

	form := userforms.NewRecordVerificationConfirm(testApp, authCollection)
	form.Token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjRxMXhsY2xtZmxva3UzMyIsImVtYWlsIjoidGVzdEBleGFtcGxlLmNvbSIsImNvbGxlY3Rpb25JZCI6Il9wYl91c2Vyc19hdXRoXyIsInR5cGUiOiJhdXRoUmVjb3JkIiwiZXhwIjoyMjA4OTg1MjYxfQ.hL16TVmStHFdHLc4a860bRqJ3sFfzjv0_NRNzwsvsrc"
	interceptorVerified := authRecord.Verified()
	testErr := errors.New("test_error")

	interceptor1Called := false
	interceptor1 := func(next baseforms.InterceptorNextFunc[*recordmodels.Record]) baseforms.InterceptorNextFunc[*recordmodels.Record] {
		return func(record *recordmodels.Record) error {
			interceptor1Called = true
			return next(record)
		}
	}

	interceptor2Called := false
	interceptor2 := func(next baseforms.InterceptorNextFunc[*recordmodels.Record]) baseforms.InterceptorNextFunc[*recordmodels.Record] {
		return func(record *recordmodels.Record) error {
			interceptorVerified = record.Verified()
			interceptor2Called = true
			return testErr
		}
	}

	_, submitErr := form.Submit(interceptor1, interceptor2)
	if submitErr != testErr {
		t.Fatalf("Expected submitError %v, got %v", testErr, submitErr)
	}

	if !interceptor1Called {
		t.Fatalf("Expected interceptor1 to be called")
	}

	if !interceptor2Called {
		t.Fatalf("Expected interceptor2 to be called")
	}

	if interceptorVerified == authRecord.Verified() {
		t.Fatalf("Expected the form model to be filled before calling the interceptors")
	}
}
