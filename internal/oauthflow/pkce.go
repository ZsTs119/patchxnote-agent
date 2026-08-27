package oauthflow

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"regexp"
)

var verifierPattern = regexp.MustCompile(`^[A-Za-z0-9._~-]{43,128}$`)

type PKCEPair struct {
	Verifier  string
	Challenge string
}

func GeneratePKCE() (PKCEPair, error) {
	return GeneratePKCEWithReader(rand.Reader)
}

func GeneratePKCEWithReader(reader io.Reader) (PKCEPair, error) {
	if reader == nil {
		reader = rand.Reader
	}
	var raw [32]byte
	if _, err := io.ReadFull(reader, raw[:]); err != nil {
		return PKCEPair{}, fmt.Errorf("generate pkce verifier: %w", err)
	}
	verifier := base64.RawURLEncoding.EncodeToString(raw[:])
	if !verifierPattern.MatchString(verifier) {
		return PKCEPair{}, fmt.Errorf("generated pkce verifier is invalid")
	}
	return PKCEPair{
		Verifier:  verifier,
		Challenge: ChallengeS256(verifier),
	}, nil
}

func ChallengeS256(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
