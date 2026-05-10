package tokens_test

import (
	"testing"

	"github.com/pafthang/servicebase/tests"
	"github.com/pafthang/servicebase/tools/tokens"
)

func TestNewRecordAuthToken(t *testing.T) {
	t.Parallel()

	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	user, err := app.Dao().FindUserRecordByEmail("users", "test@example.com")
	if err != nil {
		t.Fatal(err)
	}

	token, err := tokens.NewRecordAuthToken(app, user)
	if err != nil {
		t.Fatal(err)
	}

	tokenRecord, _ := app.Dao().FindUserRecordByToken(
		token,
		app.Settings().RecordAuthToken.Secret,
	)
	if tokenRecord == nil || tokenRecord.Id != user.Id {
		t.Fatalf("Expected users record %v, got %v", user, tokenRecord)
	}
}

func TestNewRecordVerifyToken(t *testing.T) {
	t.Parallel()

	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	user, err := app.Dao().FindUserRecordByEmail("users", "test@example.com")
	if err != nil {
		t.Fatal(err)
	}

	token, err := tokens.NewRecordVerifyToken(app, user)
	if err != nil {
		t.Fatal(err)
	}

	tokenRecord, _ := app.Dao().FindUserRecordByToken(
		token,
		app.Settings().RecordVerificationToken.Secret,
	)
	if tokenRecord == nil || tokenRecord.Id != user.Id {
		t.Fatalf("Expected users record %v, got %v", user, tokenRecord)
	}
}

func TestNewRecordResetPasswordToken(t *testing.T) {
	t.Parallel()

	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	user, err := app.Dao().FindUserRecordByEmail("users", "test@example.com")
	if err != nil {
		t.Fatal(err)
	}

	token, err := tokens.NewRecordResetPasswordToken(app, user)
	if err != nil {
		t.Fatal(err)
	}

	tokenRecord, _ := app.Dao().FindUserRecordByToken(
		token,
		app.Settings().RecordPasswordResetToken.Secret,
	)
	if tokenRecord == nil || tokenRecord.Id != user.Id {
		t.Fatalf("Expected users record %v, got %v", user, tokenRecord)
	}
}

func TestNewRecordChangeEmailToken(t *testing.T) {
	t.Parallel()

	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	user, err := app.Dao().FindUserRecordByEmail("users", "test@example.com")
	if err != nil {
		t.Fatal(err)
	}

	token, err := tokens.NewRecordChangeEmailToken(app, user, "test_new@example.com")
	if err != nil {
		t.Fatal(err)
	}

	tokenRecord, _ := app.Dao().FindUserRecordByToken(
		token,
		app.Settings().RecordEmailChangeToken.Secret,
	)
	if tokenRecord == nil || tokenRecord.Id != user.Id {
		t.Fatalf("Expected users record %v, got %v", user, tokenRecord)
	}
}

func TestNewRecordFileToken(t *testing.T) {
	t.Parallel()

	app, _ := tests.NewTestApp()
	defer app.Cleanup()

	user, err := app.Dao().FindUserRecordByEmail("users", "test@example.com")
	if err != nil {
		t.Fatal(err)
	}

	token, err := tokens.NewRecordFileToken(app, user)
	if err != nil {
		t.Fatal(err)
	}

	tokenRecord, _ := app.Dao().FindUserRecordByToken(
		token,
		app.Settings().RecordFileToken.Secret,
	)
	if tokenRecord == nil || tokenRecord.Id != user.Id {
		t.Fatalf("Expected users record %v, got %v", user, tokenRecord)
	}
}
