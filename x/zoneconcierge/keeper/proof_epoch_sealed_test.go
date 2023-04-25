package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/crypto/bls12381"
	"github.com/babylonchain/babylon/testutil/datagen"
	testkeeper "github.com/babylonchain/babylon/testutil/keeper"
	checkpointingtypes "github.com/babylonchain/babylon/x/checkpointing/types"
	zckeeper "github.com/babylonchain/babylon/x/zoneconcierge/keeper"
	zctypes "github.com/babylonchain/babylon/x/zoneconcierge/types"
	"github.com/boljen/go-bitmap"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func signBLSWithBitmap(blsSKs []bls12381.PrivateKey, bm bitmap.Bitmap, msg []byte) (bls12381.Signature, error) {
	sigs := []bls12381.Signature{}
	for i := 0; i < len(blsSKs); i++ {
		if bitmap.Get(bm, i) {
			sig := bls12381.Sign(blsSKs[i], msg)
			sigs = append(sigs, sig)
		}
	}
	return bls12381.AggrSigList(sigs)
}

// FuzzProofEpochSealed fuzz tests the prover and verifier of ProofEpochSealed
// Process:
// 1. Generate a random epoch that has a legitimate-looking SealerHeader
// 2. Generate a random validator set with BLS PKs
// 3. Generate a BLS multisig with >1/3 random validators of the validator set
// 4. Generate a checkpoint based on the above validator subset and the above sealer header
// 5. Execute ProveEpochSealed where the mocked checkpointing keeper produces the above validator set
// 6. Execute VerifyEpochSealed with above epoch, checkpoint and proof, and assert the outcome to be true
//
// Tested property: proof is valid only when
// - BLS sig in proof is valid
func FuzzProofEpochSealed_BLSSig(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		r := rand.New(rand.NewSource(seed))

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// generate a random epoch
		epoch := datagen.GenRandomEpoch(r)

		// generate a random validator set with 100 validators
		numVals := 100
		valSet, blsSKs := datagen.GenerateValidatorSetWithBLSPrivKeys(numVals)

		// sample a validator subset, which may or may not reach a quorum
		bm, numSubSet := datagen.GenRandomBitmap(r)
		_, subsetPower, err := valSet.FindSubsetWithPowerSum(bm)
		require.NoError(t, err)

		// construct the rawCkpt
		// Note that the BlsMultiSig will be generated and assigned later
		lch := checkpointingtypes.LastCommitHash(epoch.SealerHeader.LastCommitHash)
		rawCkpt := &checkpointingtypes.RawCheckpoint{
			EpochNum:       epoch.EpochNumber,
			LastCommitHash: &lch,
			Bitmap:         bm,
			BlsMultiSig:    nil,
		}

		// let the subset generate a BLS multisig over sealer header's last_commit_hash
		multiSig, err := signBLSWithBitmap(blsSKs, bm, rawCkpt.SignedMsg())
		require.NoError(t, err)
		// assign multiSig to rawCkpt
		rawCkpt.BlsMultiSig = &multiSig

		// mock checkpointing keeper that produces the expected validator set
		checkpointingKeeper := zctypes.NewMockCheckpointingKeeper(ctrl)
		checkpointingKeeper.EXPECT().GetBLSPubKeySet(gomock.Any(), gomock.Eq(epoch.EpochNumber)).Return(valSet.ValSet, nil).AnyTimes()
		// mock epoching keeper
		epochingKeeper := zctypes.NewMockEpochingKeeper(ctrl)
		epochingKeeper.EXPECT().GetEpoch(gomock.Any()).Return(epoch).AnyTimes()
		epochingKeeper.EXPECT().GetHistoricalEpoch(gomock.Any(), gomock.Eq(epoch.EpochNumber)).Return(epoch, nil).AnyTimes()
		// create zcKeeper and ctx
		zcKeeper, ctx := testkeeper.ZoneConciergeKeeper(t, checkpointingKeeper, nil, epochingKeeper, nil)

		// prove
		proof, err := zcKeeper.ProveEpochSealed(ctx, epoch.EpochNumber)
		require.NoError(t, err)
		// verify
		err = zckeeper.VerifyEpochSealed(epoch, rawCkpt, proof)

		if subsetPower <= valSet.GetTotalPower()*1/3 { // BLS sig does not reach a quorum
			require.LessOrEqual(t, numSubSet, numVals*1/3)
			require.Error(t, err)
			require.NotErrorIs(t, err, zctypes.ErrInvalidMerkleProof)
		} else { // BLS sig has a valid quorum
			require.Greater(t, numSubSet, numVals*1/3)
			require.Error(t, err)
			require.ErrorIs(t, err, zctypes.ErrInvalidMerkleProof)
		}
	})
}

// - TODO: epoch metadata has a valid inclusion proof
func FuzzProofEpochSealed_Epoch(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		// r := rand.New(rand.NewSource(seed))
	})
}

// - TODO: BLS val set has a valid inclusion proof
func FuzzProofEpochSealed_ValSet(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		// r := rand.New(rand.NewSource(seed))
	})
}
