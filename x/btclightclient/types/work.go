package types

import (
	sdkmath "cosmossdk.io/math"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/btcsuite/btcd/blockchain"
)

func CalcWork(header *bbn.BTCHeaderBytes) sdkmath.Uint {
	return sdkmath.NewUintFromBigInt(blockchain.CalcWork(header.Bits()))
}

func CumulativeWork(childWork sdkmath.Uint, parentWork sdkmath.Uint) sdkmath.Uint {
	sum := sdkmath.NewUint(0)
	sum = sum.Add(childWork)
	sum = sum.Add(parentWork)
	return sum
}
