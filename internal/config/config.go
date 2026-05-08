package config

import (
	"errors"
	"os"
)

const (
	EnvToken   = "IBKR_FLEX_TOKEN"
	EnvQueryID = "IBKR_FLEX_QUERY_ID"
)

type Config struct {
	Token   string
	QueryID string
}

var ErrMissing = errors.New("config: required env var not set")

// Load reads IBKR_FLEX_TOKEN and IBKR_FLEX_QUERY_ID from the environment.
// Returns ErrMissing wrapped with the offending var name if either is empty.
func Load() (*Config, error) {
	tok := os.Getenv(EnvToken)
	qid := os.Getenv(EnvQueryID)
	if tok == "" {
		return nil, &MissingError{Var: EnvToken}
	}
	if qid == "" {
		return nil, &MissingError{Var: EnvQueryID}
	}
	return &Config{Token: tok, QueryID: qid}, nil
}

type MissingError struct {
	Var string
}

func (e *MissingError) Error() string {
	return "config: required env var not set: " + e.Var
}

func (e *MissingError) Is(target error) bool {
	return target == ErrMissing
}
