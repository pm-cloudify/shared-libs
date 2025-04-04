package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/argon2"
)

// generating JWT for user
func GenerateToken(username, secret_key string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	return token.SignedString([]byte(secret_key))
}

// validating tokens
func ValidateToken(token_str, secret_key string) (jwt.MapClaims, error) {

	token, err := jwt.Parse(token_str, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(secret_key), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

type Argon2Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// TODO: 1- import conf from env
// TODO: 2- read more about effects of this config
// TODO: 3- use a wrapper pkg to use Argon2
var DefaultArgon2Params = Argon2Params{
	Memory:      64 * 1024,
	Iterations:  3,
	Parallelism: 2,
	SaltLength:  16,
	KeyLength:   32,
}

func GenerateHash(password string, p Argon2Params) (string, error) {
	// generate salt
	salt := make([]byte, p.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	// generate hash
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		p.Iterations,
		p.Memory,
		p.Parallelism,
		p.KeyLength,
	)

	// encode salt and hash with b64
	b64_salt := base64.RawStdEncoding.EncodeToString(salt)
	b64_hash := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		p.Memory,
		p.Iterations,
		p.Parallelism,
		b64_salt,
		b64_hash,
	)

	log.Printf("encoded pass: %s", encoded)
	return encoded, nil
}

func VerifyPassword(encoded, password string) (bool, error) {
	// decode stored hash
	parts := strings.Split(encoded, "$")
	log.Println(parts)
	if len(parts) != 6 {
		log.Printf("invalid encoded password format")
		return false, errors.New("invalid hash format")
	}

	// parse params
	var version int
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return false, err
	}
	if version != argon2.Version {
		return false, errors.New("incompatible argon2 version")
	}

	p := Argon2Params{}
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &p.Memory, &p.Iterations, &p.Parallelism)
	if err != nil {
		return false, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}
	p.SaltLength = uint32(len(salt))

	storedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}
	p.KeyLength = uint32(len(storedHash))

	// generate hash with same parameters
	comparisonHash := argon2.IDKey(
		[]byte(password),
		salt,
		p.Iterations,
		p.Memory,
		p.Parallelism,
		p.KeyLength,
	)

	// constant time comparison
	if len(comparisonHash) != len(storedHash) {
		return false, nil
	}
	for i := 0; i < len(comparisonHash); i++ {
		if comparisonHash[i] != storedHash[i] {
			return false, nil
		}
	}

	return true, nil
}
