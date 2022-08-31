package w3cdid

import "github.com/pkg/errors"

func (d *Document) IsValid() error {
	return errors.New("not implemented")
}

//Signed checks if the signature provided was signed
//by a key in the Document. If prev is provided, the signature
//is compared to keys in the previous Document rather than the current
func (d *Document) Signed(signature []byte, msg []byte) error {
	return errors.New("not implemented")
}
