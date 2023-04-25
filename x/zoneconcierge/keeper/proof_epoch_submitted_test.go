package keeper_test

import (
	"encoding/hex"
	"math/rand"
	"testing"

	txformat "github.com/babylonchain/babylon/btctxformatter"
	"github.com/babylonchain/babylon/testutil/datagen"
	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"
	zckeeper "github.com/babylonchain/babylon/x/zoneconcierge/keeper"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
	"github.com/stretchr/testify/require"
)

func FuzzProofEpochSubmitted(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		// generate random epoch, random rawBtcCkpt and random rawCkpt
		epoch := datagen.GenRandomEpoch(r)
		rawBtcCkpt := datagen.GetRandomRawBtcCheckpoint(r)
		rawBtcCkpt.Epoch = epoch.EpochNumber
		rawCkpt, err := checkpointingtypes.FromBTCCkptToRawCkpt(rawBtcCkpt)
		require.NoError(t, err)

		// encode ckpt to BTC txs in BTC blocks
		testRawCkptData := datagen.EncodeRawCkptToTestData(rawBtcCkpt)
		idxs := []uint64{datagen.RandomInt(r, 5) + 1, datagen.RandomInt(r, 5) + 1}
		offsets := []uint64{datagen.RandomInt(r, 5) + 1, datagen.RandomInt(r, 5) + 1}
		btcBlocks := []*datagen.BlockCreationResult{
			datagen.CreateBlock(r, 1, uint32(idxs[0]+offsets[0]), uint32(idxs[0]), testRawCkptData.FirstPart),
			datagen.CreateBlock(r, 2, uint32(idxs[1]+offsets[1]), uint32(idxs[1]), testRawCkptData.SecondPart),
		}
		// create MsgInsertBtcSpvProof for the rawCkpt
		msgInsertBtcSpvProof := datagen.GenerateMessageWithRandomSubmitter([]*datagen.BlockCreationResult{btcBlocks[0], btcBlocks[1]})

		// get headers for verification
		btcHeaders := []*wire.BlockHeader{
			btcBlocks[0].HeaderBytes.ToBlockHeader(),
			btcBlocks[1].HeaderBytes.ToBlockHeader(),
		}

		// get 2 tx info for the ckpt parts
		txsInfo := []*btcctypes.TransactionInfo{
			{
				Key:         &btcctypes.TransactionKey{Index: uint32(idxs[0]), Hash: btcBlocks[0].HeaderBytes.Hash()},
				Transaction: msgInsertBtcSpvProof.Proofs[0].BtcTransaction,
				Proof:       msgInsertBtcSpvProof.Proofs[0].MerkleNodes,
			},
			{
				Key:         &btcctypes.TransactionKey{Index: uint32(idxs[1]), Hash: btcBlocks[1].HeaderBytes.Hash()},
				Transaction: msgInsertBtcSpvProof.Proofs[1].BtcTransaction,
				Proof:       msgInsertBtcSpvProof.Proofs[1].MerkleNodes,
			},
		}

		// net param, babylonTag
		powLimit := chaincfg.SimNetParams.PowLimit
		babylonTag := btcctypes.DefaultCheckpointTag
		tagAsBytes, _ := hex.DecodeString(babylonTag)

		// verify
		err = zckeeper.VerifyEpochSubmitted(rawCkpt, txsInfo, btcHeaders, powLimit, txformat.BabylonTag(tagAsBytes))
		require.NoError(t, err)
	})
}
