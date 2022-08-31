package w3cdid

import "github.com/tcfw/didem/pkg/did/w3cdid/cryptography"

type Document struct {
	Context              []string                          `json:"@context"`
	ID                   string                            `json:"id"`
	AlsoKnownAs          []string                          `json:"alsoKnownAs,omitempty"`
	Controller           []string                          `json:"controller,omitempty"`
	VerificationMethod   []cryptography.VerificationMethod `json:"verificationMethod,omitempty"`
	Authentication       []cryptography.VerificationMethod `json:"authentication,omitempty"`
	AssertionMethod      []cryptography.VerificationMethod `json:"assertionMethod,omitempty"`
	KeyAgreement         []cryptography.VerificationMethod `json:"keyAgreement,omitempty"`
	CapabilityInvocation []cryptography.VerificationMethod `json:"capabilityInvocation,omitempty"`
	CapabilityDelegation []cryptography.VerificationMethod `json:"capabilityDelegation,omitempty"`
	Service              []Service                         `json:"service,omitempty"`
}

type Service struct {
	ID              string `json:"id"`
	Type            string `json:"type"`
	ServiceEndpoint string `json:"serviceEndpoint"`
}
