package storage

import "github.com/pkg/errors"

var (
	ErrNotFound = errors.New("not found")

	ErrDIDAlreadyExists    = errors.New("DID already exists")
	ErrDIDInvalid          = errors.New("DID is invalid")
	ErrDIDInvalidSignature = errors.New("tx signature not signed by key in DID")

	ErrOpNotSupported = errors.New("operation not supported on tx type")

	ErrTxVersionNotSupported = errors.New("tx version not supported")
	ErrTxMissingTimestamp    = errors.New("tx missing timestamp")
	ErrTxMissingFrom         = errors.New("tx missing from did")
	ErrTxUnsupportedType     = errors.New("unsupported tx type")
	ErrTxMissingData         = errors.New("tx missing data")
)
