package consensus

func signatureData(msg *Msg) ([]byte, error) {
	sig := msg.Signature
	msg.Signature = nil
	d, err := msg.Marshal()
	if err != nil {
		return nil, err
	}

	msg.Signature = sig
	return d, nil
}
