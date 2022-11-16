package keeper_test

import (
	"math/rand"
	"testing"

	"github.com/babylonchain/babylon/testutil/datagen"
	ibctmtypes "github.com/cosmos/ibc-go/v5/modules/light-clients/07-tendermint/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func FuzzEpochChainInfoIndexer(f *testing.F) {
	datagen.AddRandomSeedsToFuzzer(f, 10)

	f.Fuzz(func(t *testing.T, seed int64) {
		rand.Seed(seed)

		_, babylonChain, czChain, zcKeeper := SetupTest(t)

		ctx := babylonChain.GetContext()
		hooks := zcKeeper.Hooks()

		// invoke the hook a random number of times to simulate a random number of blocks
		numHeaders := datagen.RandomInt(100)
		for i := uint64(0); i < numHeaders; i++ {
			header := &ibctmtypes.Header{
				SignedHeader: &tmproto.SignedHeader{
					Header: &tmproto.Header{
						ChainID:        czChain.ChainID,
						Height:         int64(i),
						LastCommitHash: datagen.GenRandomByteArray(32),
					},
				},
			}
			hooks.AfterHeaderWithValidCommit(ctx, datagen.GenRandomByteArray(32), header, false)
		}

		// generate a number of fork headers
		numForkHeaders := int(datagen.RandomInt(10))
		for i := 0; i < numForkHeaders; i++ {
			header := &ibctmtypes.Header{
				SignedHeader: &tmproto.SignedHeader{
					Header: &tmproto.Header{
						ChainID:        czChain.ChainID,
						Height:         int64(numHeaders - 1),
						LastCommitHash: datagen.GenRandomByteArray(32),
					},
				},
			}
			hooks.AfterHeaderWithValidCommit(ctx, datagen.GenRandomByteArray(32), header, true)
		}

		// simulate the scenario that a random epoch has ended
		epochNum := datagen.RandomInt(10)
		hooks.AfterEpochEnds(ctx, epochNum)

		// check if the chain info of this epoch is recorded or not
		chainInfo, err := zcKeeper.GetEpochChainInfo(ctx, czChain.ChainID, epochNum)
		require.NoError(t, err)
		require.Equal(t, numHeaders-1, chainInfo.LatestHeader.Height)
		require.Equal(t, numForkHeaders, len(chainInfo.LatestForks.Headers))
	})
}
