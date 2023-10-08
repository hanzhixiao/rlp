package core

import (
	"awesomeProject/common"
	"math/big"
)

type Genesis struct {
	Timestamp  uint64         `json:"timestamp"`
	Nonce      uint64         `json:"nonce"`
	GasLimit   uint64         `json:"gasLimit"`
	ExtraData  []byte         `json:"extraData"`
	Coinbase   common.Address `json:"coinbase"`
	Difficulty *big.Int       `json:"difficulty"`
}
