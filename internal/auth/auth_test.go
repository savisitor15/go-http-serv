package auth_test

import (
	"testing"

	"github.com/savisitor15/go-http-serv/internal/auth"
)

// Test hashing of a password
func TestHashPassword(t *testing.T) {
	password := "PASSW0RD!"
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Errorf(`HashPassword("PASSW0RD!") resulted in an error %v`, err)
	}
	if err := auth.CheckPasswordHash(hash, password); err != nil{
		t.Errorf("Hash is unusable")
	}
}
