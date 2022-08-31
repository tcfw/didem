package storage

import "github.com/pkg/errors"

var (
	ErrNotFound = errors.New("not found")

	ErrDIDAlreadyExists    = errors.New("DID already exists")
	ErrDIDInvalid          = errors.New("DID is invalid")
	ErrDIDInvalidSignature = errors.New("tx signature not signed by key in DID")

	ErrOpNotSupported = errors.New("tx operation not supported on tx type")
)
