package main

import (
	"awesomeProject/core/types"
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	block := types.Newblock()
	f, err := os.OpenFile("rlp.dat", os.O_WRONLY, 0666)
	defer f.Close()
	if err != nil {
		fmt.Println(err)
	}
	err = block.EncodeRLP(f)
	if err != nil {
		fmt.Println(err)
	}
	extblock := types.Newextblock(block)
	jsonfile, err := os.Create("json.dat")
	if err != nil {
		fmt.Println(err)
		return
	}
	encoder := json.NewEncoder(jsonfile)
	err = encoder.Encode(extblock)
	if err != nil {
		fmt.Println(err)
		return
	}

	fi, err := os.Stat("rlp.dat")
	if err == nil {
		fmt.Println("rlp file size is ", fi.Size(), "bytes")
	}

	fi, err = os.Stat("json.dat")
	if err == nil {
		fmt.Println("json file size is ", fi.Size(), "bytes")
	}
	return
}
