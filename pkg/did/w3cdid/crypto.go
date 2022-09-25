package w3cdid

import (
	"github.com/pkg/errors"
	"github.com/tcfw/didem/internal/utils/logging"
	"github.com/tcfw/didem/pkg/cryptography"
)

type SignatureValidator func(vm cryptography.VerificationMethod, sig []byte, msg []byte) (bool, error)

var (
	ErrNoValidSignatures = errors.New("no valid signatures")

	validators = map[cryptography.VerificationMethodType]SignatureValidator{
		cryptography.Ed25519VerificationKey2018:        cryptography.ValidateEd25519,
		cryptography.Bls12381G2Key2020:                 cryptography.ValidateBls12381,
		cryptography.EcdsaSecp256k1VerificationKey2019: cryptography.ValidateEcdsaSecp256k1,
	}
)

// Signed checks if the signature provided was signed
// by a key in the Document. If prev is provided, the signature
// is compared to keys in the previous Document rather than the current
func (d *Document) Signed(signature []byte, msg []byte) error {
	if len(d.VerificationMethod) == 0 {
		return errors.New("no verification method specified")
	}

	for _, vm := range d.VerificationMethod {
		validator, ok := validators[vm.Type]
		if !ok {
			logging.Entry().Debugf("unsupported verification type: %s", vm.Type)
			continue
		}

		ok, err := validator(vm, signature, msg)
		if err != nil {
			logging.Entry().WithField("type", vm.Type).WithError(err).Debug("validating signature")
			continue
		}

		if ok {
			return nil
		}
	}

	return ErrNoValidSignatures
}
