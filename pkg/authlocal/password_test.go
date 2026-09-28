package authlocal

import (
	"strings"
	"testing"
)

func TestHashIsArgon2idAndNotPlaintext(t *testing.T) {
	const password = "s3cret-horse-battery"
	encoded, err := Hash(password)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(encoded, password) {
		t.Fatalf("hash contains plaintext: %s", encoded)
	}
	if !strings.HasPrefix(encoded, "$argon2id$v=19$") {
		t.Fatalf("hash = %s", encoded)
	}
	if !Verify(password, encoded) {
		t.Fatal("verify rejected the password that was hashed")
	}
	if Verify("other-password-value", encoded) {
		t.Fatal("verify accepted a different password")
	}
	if Verify(password, "not-a-hash") {
		t.Fatal("verify accepted a malformed hash")
	}
}
