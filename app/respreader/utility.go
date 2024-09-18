package respreader

func stripTerminator(bs []byte) ([]byte, error) {
	if len(bs) < 2 || bs[len(bs)-2] != '\r' || bs[len(bs)-1] != '\n' {
		return nil, ErrorInvalidTerminator
	}
	return bs[0 : len(bs)-2], nil
}
