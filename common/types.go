package common

const (
	AddressLength = 20
	HashLength    = 32
)

type Address [AddressLength]byte
type Hash [HashLength]byte

func (a Address) Bytes() []byte { return a[:] }

func (h *Hash) SetBytes(b []byte) {
	if len(b) > len(h) {
		b = b[len(b)-HashLength:]
	}

	copy(h[HashLength-len(b):], b)
}

func BytesToHash(b []byte) Hash {
	var h Hash
	h.SetBytes(b)
	return h
}
func HexToHash(s string) Hash { return BytesToHash(FromHex(s)) }
