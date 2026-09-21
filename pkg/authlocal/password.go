package authlocal

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// OWASP Argon2id parameters. Memory is KiB (64 MiB).
const (
	argonMemory  uint32 = 64 * 1024
	argonTime    uint32 = 3
	argonThreads uint8  = 2
	argonKeyLen  uint32 = 32
	argonSaltLen        = 16
)

var errBadHash = errors.New("bad password hash")

// Hash returns an Argon2id PHC string. The plaintext password is not retained.
func Hash(password string) (string, error) {
	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	sum := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return encode(salt, sum, argonTime, argonMemory, argonThreads), nil
}

// Verify compares password to an Argon2id PHC string in constant time.
func Verify(password, encoded string) bool {
	salt, want, t, mem, threads, err := decode(encoded)
	if err != nil {
		return false
	}
	sum := argon2.IDKey([]byte(password), salt, t, mem, threads, uint32(len(want)))
	return subtle.ConstantTimeCompare(sum, want) == 1
}

func encode(salt, sum []byte, t, mem uint32, threads uint8) string {
	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		mem, t, threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(sum),
	)
}

func decode(encoded string) (salt, sum []byte, t, mem uint32, threads uint8, err error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return nil, nil, 0, 0, 0, errBadHash
	}
	var version int
	if _, err = fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != 19 {
		return nil, nil, 0, 0, 0, errBadHash
	}
	var threadsInt int
	if _, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &mem, &t, &threadsInt); err != nil {
		return nil, nil, 0, 0, 0, errBadHash
	}
	if mem == 0 || mem > 256*1024 || t == 0 || t > 10 || threadsInt < 1 || threadsInt > 8 {
		return nil, nil, 0, 0, 0, errBadHash
	}
	threads = uint8(threadsInt)
	salt, err = base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) < 8 || len(salt) > 64 {
		return nil, nil, 0, 0, 0, errBadHash
	}
	sum, err = base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(sum) < 16 || len(sum) > 64 {
		return nil, nil, 0, 0, 0, errBadHash
	}
	return salt, sum, t, mem, threads, nil
}
