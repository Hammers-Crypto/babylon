package keeper

import (
	"context"
	"fmt"

	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	"github.com/babylonchain/babylon/x/zoneconcierge/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) ChainList(c context.Context, req *types.QueryChainListRequest) (*types.QueryChainListResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	chainIDs := k.GetAllChainIDs(ctx)
	// TODO: pagination for this API
	resp := &types.QueryChainListResponse{ChainIds: chainIDs}
	return resp, nil
}

// ChainInfo returns the latest info of a chain with given ID
func (k Keeper) ChainInfo(c context.Context, req *types.QueryChainInfoRequest) (*types.QueryChainInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if len(req.ChainId) == 0 {
		return nil, status.Error(codes.InvalidArgument, "chain ID cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(c)

	// find the chain info of this epoch
	chainInfo := k.GetChainInfo(ctx, req.ChainId)
	resp := &types.QueryChainInfoResponse{ChainInfo: chainInfo}
	return resp, nil
}

func (k Keeper) FinalizedChainInfo(c context.Context, req *types.QueryFinalizedChainInfoRequest) (*types.QueryFinalizedChainInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if len(req.ChainId) == 0 {
		return nil, status.Error(codes.InvalidArgument, "chain ID cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(c)

	// find the last finalised epoch
	finalizedEpoch, err := k.GetFinalizedEpoch(ctx)
	if err != nil {
		return nil, err
	}

	// find the chain info of this epoch
	chainInfo, err := k.GetEpochChainInfo(ctx, req.ChainId, finalizedEpoch)
	if err != nil {
		return nil, err
	}

	// It's possible that the chain info's epoch is way before the last finalised epoch
	// e.g., when there is no relayer for many epochs
	// NOTE: if an epoch is finalisedm then all of its previous epochs are also finalised
	if chainInfo.LatestHeader.BabylonEpoch < finalizedEpoch {
		finalizedEpoch = chainInfo.LatestHeader.BabylonEpoch
	}

	// find the metadata of this epoch
	epochInfo, err := k.epochingKeeper.GetHistoricalEpoch(ctx, finalizedEpoch)
	if err != nil {
		return nil, err
	}

	// find the btc checkpoint info of this epoch
	ed := k.btccKeeper.GetEpochData(ctx, finalizedEpoch)
	if ed.Status != btcctypes.Finalized {
		err := fmt.Errorf("epoch %d should have been finalized, but is in status %s", finalizedEpoch, ed.Status.String())
		panic(err)
	}
	if len(ed.Key) == 0 {
		err := fmt.Errorf("finalized epoch %d should have at least 1 checkpoint submission", finalizedEpoch)
		panic(err)
	}
	bestSubmissionKey := ed.Key[0]

	// TODO: construct inclusion proofs

	resp := &types.QueryFinalizedChainInfoResponse{
		FinalizedChainInfo: chainInfo,
		EpochInfo:          epochInfo,
		BtcCheckpointInfo:  bestSubmissionKey,
	}
	return resp, nil
}
