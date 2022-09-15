package cryptography

import "errors"

type VerificationMethodType string

var (
	ErrInvalidPublicKey         = errors.New("invalid public key")
	ErrInvalidPublicKeyLength   = errors.New("invalid public key length")
	ErrInvalidPublicKeyType     = errors.New("invalid public key type")
	ErrUnsupportedPublicKeyType = errors.New("unsupported public key type")
)

const (
	Bls12381G1Key2020                 VerificationMethodType = "Bls12381G1Key2020"
	Bls12381G2Key2020                 VerificationMethodType = "Bls12381G2Key2020"
	EcdsaSecp256k1RecoveryMethod2020  VerificationMethodType = "EcdsaSecp256k1RecoveryMethod2020"
	EcdsaSecp256k1VerificationKey2019 VerificationMethodType = "EcdsaSecp256k1VerificationKey2019"
	Ed25519VerificationKey2018        VerificationMethodType = "Ed25519VerificationKey2018"
	JsonWebKey2020                    VerificationMethodType = "JsonWebKey2020"
	PgpVerificationkey2021            VerificationMethodType = "PgpVerificationkey2021"
	RsaVerificationKey2018            VerificationMethodType = "RsaVerificationKey2018"
	Verificationcondition2021         VerificationMethodType = "Verificationcondition2021"
	X25519KeyAgreementKey2019         VerificationMethodType = "X25519KeyAgreementKey2019"
)

type VerificationMethod struct {
	ID                 string                 `json:"id"`
	Type               VerificationMethodType `json:"type"`
	Controller         string                 `json:"controller"`
	PublicKeyJwk       []interface{}          `json:"publicKeyJwk,omitempty"`
	PublicKeyMultibase string                 `json:"publicKeyMultibase,omitempty"`
}
