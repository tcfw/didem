package w3cdid

import (
	"github.com/pkg/errors"
	"github.com/tcfw/didem/pkg/cryptography"
)

var (
	ErrInvalid = errors.New("did invalid")
)

func (d *Document) IsValid() error {
	if d.ID == "" {
		return ErrInvalid
	}

	didurl := URL(d.ID)
	if didurl.Id() == "" || didurl.Method() == "" {
		return ErrInvalid
	}

	vms := []cryptography.VerificationMethod{}
	vms = append(vms, d.VerificationMethod...)
	vms = append(vms, d.Authentication...)
	vms = append(vms, d.AssertionMethod...)
	vms = append(vms, d.KeyAgreement...)
	vms = append(vms, d.CapabilityInvocation...)
	vms = append(vms, d.CapabilityDelegation...)

	for _, v := range vms {
		if v.Type == "" {
			return ErrInvalid
		}
		idurl := URL(v.ID)
		controllerurl := URL(v.Controller)
		if idurl.Id() == "" || idurl.Method() == "" {
			return ErrInvalid
		}
		if controllerurl.Id() == "" || controllerurl.Method() == "" {
			return ErrInvalid
		}
	}

	for _, s := range d.Service {
		if s.ID == "" || s.Type == "" || s.ServiceEndpoint == "" {
			return ErrInvalid
		}
	}

	return nil
}
