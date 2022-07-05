package w3cdid

type Document struct {
	Context              []string             `json:"@context"`
	ID                   string               `json:"id"`
	AlsoKnownAs          []string             `json:"alsoKnownAs,omitempty"`
	Controller           []string             `json:"controller,omitempty"`
	VerificationMethod   []VerificationMethod `json:"verificationMethod,omitempty"`
	Authentication       []VerificationMethod `json:"authentication,omitempty"`
	AssertionMethod      []VerificationMethod `json:"assertionMethod,omitempty"`
	KeyAgreement         []VerificationMethod `json:"keyAgreement,omitempty"`
	CapabilityInvocation []VerificationMethod `json:"capabilityInvocation,omitempty"`
	CapabilityDelegation []VerificationMethod `json:"capabilityDelegation,omitempty"`
	Service              []Service            `json:"service,omitempty"`
}

type VerificationMethodTypes string

const (
	Bls12381G1Key2020                 VerificationMethodTypes = "Bls12381G1Key2020"
	Bls12381G2Key2020                 VerificationMethodTypes = "Bls12381G2Key2020"
	EcdsaSecp256k1RecoveryMethod2020  VerificationMethodTypes = "EcdsaSecp256k1RecoveryMethod2020"
	EcdsaSecp256k1VerificationKey2019 VerificationMethodTypes = "EcdsaSecp256k1VerificationKey2019"
	Ed25519VerificationKey2018        VerificationMethodTypes = "Ed25519VerificationKey2018"
	JsonWebKey2020                    VerificationMethodTypes = "JsonWebKey2020"
	PgpVerificationkey2021            VerificationMethodTypes = "PgpVerificationkey2021"
	RsaVerificationKey2018            VerificationMethodTypes = "RsaVerificationKey2018"
	Verificationcondition2021         VerificationMethodTypes = "Verificationcondition2021"
	X25519KeyAgreementKey2019         VerificationMethodTypes = "X25519KeyAgreementKey2019"
)

type VerificationMethod struct {
	ID                 string                  `json:"id"`
	Type               VerificationMethodTypes `json:"type"`
	Controller         string                  `json:"controller"`
	PublicKeyJwk       []interface{}           `json:"publicKeyJwk,omitempty"`
	PublicKeyMultibase string                  `json:"publicKeyMultibase,omitempty"`
}

type Service struct {
	ID              string `json:"id"`
	Type            string `json:"type"`
	ServiceEndpoint string `json:"serviceEndpoint"`
}
