package main

import "testing"

func TestInsecureJWTSecret(t *testing.T) {
	for _, secret := range []string{
		"",
		"rexec-dev-secret-change-me",
		"rexec-dev-secret-change-in-production",
		"your-super-secret-jwt-key-change-this-in-production",
		"your-secure-jwt-secret-here",
		"prefix-REXEC-DEV-SECRET-suffix",
		"custom-change-me-value",
	} {
		if !insecureJWTSecret(secret) {
			t.Fatalf("expected %q to be refused", secret)
		}
	}
	if insecureJWTSecret("a-unique-production-secret-value-32b") {
		t.Fatal("unique secret was treated as a placeholder")
	}
}

func TestJWTSecretErrorOnlyInProduction(t *testing.T) {
	t.Setenv("GIN_MODE", "debug")
	t.Setenv("REXEC_ENV", "")
	t.Setenv("APP_ENV", "")
	t.Setenv("GO_ENV", "")
	t.Setenv("ENV", "")
	if err := jwtSecretError("rexec-dev-secret-change-me"); err != nil {
		t.Fatalf("debug mode refused the example secret: %v", err)
	}

	t.Setenv("GIN_MODE", "release")
	if err := jwtSecretError("rexec-dev-secret-change-me"); err == nil {
		t.Fatal("release mode accepted a placeholder secret")
	}
	if err := jwtSecretError("a-unique-production-secret-value-32b"); err != nil {
		t.Fatalf("release mode refused a unique secret: %v", err)
	}

	t.Setenv("GIN_MODE", "debug")
	t.Setenv("APP_ENV", "production")
	if err := jwtSecretError("your-secure-jwt-secret-here-but-longer"); err == nil {
		t.Fatal("APP_ENV=production accepted a placeholder secret")
	}
}
