package crypto

import "hash"

type KeccakState interface {
	hash.Hash
	Read([]byte) (int, error)
}
