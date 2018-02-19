package client

import (
	"errors"
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/protocol"
)

func validateTx(block *protocol.Block, tx protocol.Transaction, txHash [32]byte) error {
	valid := true

	nodes := reqIntermediateNodes(block.Hash, txHash)

	if txHash != tx.Hash() {
		valid = false
	}

	leafHash := txHash
	for i := 0; i < len(nodes); i += 2 {
		var parentHash [32]byte
		concatHash := append(leafHash[:], nodes[i][:]...)
		if parentHash = protocol.SerializeHashContent(concatHash); parentHash != nodes[i+1] {
			concatHash = append(nodes[i][:], leafHash[:]...)
			if parentHash = protocol.SerializeHashContent(concatHash); parentHash != nodes[i+1] {
				valid = false
			}
		}
		leafHash = parentHash
	}

	if !valid {
		return errors.New(fmt.Sprintf("Tx validation failed for %x\n", txHash))
	}

	return nil
}