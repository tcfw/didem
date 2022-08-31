package cryptography

import (
	"crypto/ed25519"

	"github.com/pkg/errors"
)

func ValidateEd25519(vm VerificationMethod, sig []byte, msg []byte) (bool, error) {
	pkbytes, err := decodeMultibase(vm.PublicKeyMultibase)
	if err != nil {
		return false, errors.Wrap(err, "decoding multibase")
	}

	pk := ed25519.PublicKey(pkbytes)

	return ed25519.Verify(pk, msg, sig), nil
}
