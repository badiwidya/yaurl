// Implementations based on Alex Edward's blog post:
// https://www.alexedwards.net/blog/how-to-hash-and-verify-passwords-with-argon2-in-go

package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type argon2Params struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

var defaultParams = &argon2Params{
	memory:      64 * 1024,
	iterations:  2,
	parallelism: 1,
	saltLength:  16,
	keyLength:   32,
}

var (
	ErrInvalidHash          = errors.New("format of encoded hash is invalid")
	ErrVersionNotCompatible = errors.New("incompatible version of argon2")
)

func hashPassword(password string, p *argon2Params) (string, error) {
	salt := make([]byte, p.saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		p.iterations,
		p.memory,
		p.parallelism,
		p.keyLength,
	)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		p.memory,
		p.iterations,
		p.parallelism,
		b64Salt,
		b64Hash,
	)

	return encodedHash, nil
}

func comparePassAndHash(password, encodedHash string) (bool, error) {
	params, salt, hash, err := decodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	otherHash := argon2.IDKey(
		[]byte(password),
		salt,
		params.iterations,
		params.memory,
		params.parallelism,
		params.keyLength,
	)

	return subtle.ConstantTimeCompare(hash, otherHash) == 1, nil
}

func decodeHash(encodedHash string) (*argon2Params, []byte, []byte, error) {
	values := strings.Split(encodedHash, "$")
	if len(values) != 6 {
		return nil, nil, nil, ErrInvalidHash
	}

	var version int
	if _, err := fmt.Sscanf(values[2], "v=%d", &version); err != nil {
		return nil, nil, nil, err
	}
	if version != argon2.Version {
		return nil, nil, nil, ErrVersionNotCompatible
	}

	params := &argon2Params{}
	if _, err := fmt.Sscanf(
		values[3],
		"m=%d,t=%d,p=%d",
		&params.memory,
		&params.iterations,
		&params.parallelism,
	); err != nil {
		return nil, nil, nil, err
	}

	salt, err := base64.RawStdEncoding.Strict().DecodeString(values[4])
	if err != nil {
		return nil, nil, nil, err
	}
	params.saltLength = uint32(len(salt))

	hash, err := base64.RawStdEncoding.Strict().DecodeString(values[5])
	if err != nil {
		return nil, nil, nil, err
	}
	params.keyLength = uint32(len(hash))

	return params, salt, hash, nil
}
