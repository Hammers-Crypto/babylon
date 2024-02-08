package keeper

import (
	"context"
	"fmt"

	corestoretypes "cosmossdk.io/core/store"

	"cosmossdk.io/log"
	"github.com/babylonchain/babylon/x/btcstaking/types"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	Keeper struct {
		cdc          codec.BinaryCodec
		storeService corestoretypes.KVStoreService

		btclcKeeper types.BTCLightClientKeeper
		btccKeeper  types.BtcCheckpointKeeper

		btcNet *chaincfg.Params
		// the address capable of executing a MsgUpdateParams message. Typically, this
		// should be the x/gov module account.
		authority string
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService corestoretypes.KVStoreService,

	btclcKeeper types.BTCLightClientKeeper,
	btccKeeper types.BtcCheckpointKeeper,

	btcNet *chaincfg.Params,
	authority string,
) Keeper {
	return Keeper{
		cdc:          cdc,
		storeService: storeService,

		btclcKeeper: btclcKeeper,
		btccKeeper:  btccKeeper,

		btcNet:    btcNet,
		authority: authority,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// BeginBlocker is invoked upon `BeginBlock` of the system. The function
// iterates over all BTC delegations under non-slashed finality providers
// to 1) record the voting power table for the current height, and 2) record
// the voting power distribution cache used for distributing rewards once
// the block is finalised by finality providers.
func (k Keeper) BeginBlocker(ctx context.Context) error {
	// index BTC height at the current height
	k.IndexBTCHeight(ctx)

	covenantQuorum := k.GetParams(ctx).CovenantQuorum
	btcTipHeight, err := k.GetCurrentBTCHeight(ctx)
	if err != nil {
		panic(err) // only possible upon programming error
	}
	wValue := k.btccKeeper.GetParams(ctx).CheckpointFinalizationTimeout

	// prepare for recording finality providers and their BTC delegations
	// for voting power and rewards
	dc := types.NewVotingPowerDistCache()

	// iterate over all finality providers to find out non-slashed ones that have
	// positive voting power
	k.IterateActiveFPs(
		ctx,
		func(fp *types.FinalityProvider) bool {
			fpDistInfo := types.NewFinalityProviderDistInfo(fp)

			// iterate over all BTC delegations under the finality provider
			// in order to accumulate voting power dist info for it
			k.IterateBTCDelegations(ctx, fp.BtcPk, func(btcDel *types.BTCDelegation) bool {
				// accumulate voting power and voting power distribution cache
				fpDistInfo.AddBTCDel(btcDel, btcTipHeight, wValue, covenantQuorum)
				return true
			})

			if fpDistInfo.TotalVotingPower > 0 {
				dc.AddFinalityProviderDistInfo(fpDistInfo)
			}

			return true
		},
	)

	// filter out top `MaxActiveFinalityProviders` active finality providers in terms of voting power
	maxNumActiveFPs := k.GetParams(ctx).MaxActiveFinalityProviders
	activeFps := types.FilterTopNFinalityProviders(dc.FinalityProviders, maxNumActiveFPs)
	// set voting power table and re-calculate total voting power of top N finality providers
	dc.TotalVotingPower = uint64(0)
	babylonTipHeight := uint64(sdk.UnwrapSDKContext(ctx).HeaderInfo().Height)
	for _, fp := range activeFps {
		k.SetVotingPower(ctx, fp.BtcPk.MustMarshal(), babylonTipHeight, fp.TotalVotingPower)
		dc.TotalVotingPower += fp.TotalVotingPower
	}

	// set the voting power distribution cache of the current height
	k.setVotingPowerDistCache(ctx, uint64(sdk.UnwrapSDKContext(ctx).HeaderInfo().Height), dc)

	return nil
}
