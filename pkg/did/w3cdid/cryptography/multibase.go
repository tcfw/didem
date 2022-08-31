package cryptography

import "github.com/multiformats/go-multibase"

func decodeMultibase(mb string) ([]byte, error) {
	_, d, err := multibase.Decode(mb)
	return d, err
}
