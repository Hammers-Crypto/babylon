package e2e

import (
	"encoding/hex"
	"math"
	"math/rand"
	"time"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/wire"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/babylonchain/babylon/crypto/eots"
	"github.com/babylonchain/babylon/test/e2e/configurer"
	"github.com/babylonchain/babylon/test/e2e/configurer/chain"
	"github.com/babylonchain/babylon/test/e2e/initialization"
	"github.com/babylonchain/babylon/testutil/datagen"
	bbn "github.com/babylonchain/babylon/types"
	btcctypes "github.com/babylonchain/babylon/x/btccheckpoint/types"
	bstypes "github.com/babylonchain/babylon/x/btcstaking/types"
	ftypes "github.com/babylonchain/babylon/x/finality/types"
	itypes "github.com/babylonchain/babylon/x/incentive/types"
)

var (
	r   = rand.New(rand.NewSource(time.Now().Unix()))
	net = &chaincfg.SimNetParams
	// finality provider
	fpBTCSK, _, _ = datagen.GenRandomBTCKeyPair(r)
	fp            *bstypes.FinalityProvider
	msr           *eots.MasterSecretRand
	// BTC delegation
	delBTCSK, delBTCPK, _ = datagen.GenRandomBTCKeyPair(r)
	// covenant
	covenantSKs, _, covenantQuorum = bstypes.DefaultCovenantCommittee()

	stakingValue = int64(2 * 10e8)
)

type BTCStakingTestSuite struct {
	suite.Suite

	configurer configurer.Configurer
}

func (s *BTCStakingTestSuite) SetupSuite() {
	s.T().Log("setting up e2e integration test suite...")
	var err error

	// The e2e test flow is as follows:
	//
	// 1. Configure 1 chain with some validator nodes
	// 2. Execute various e2e tests
	s.configurer, err = configurer.NewBTCStakingConfigurer(s.T(), true)
	s.NoError(err)
	err = s.configurer.ConfigureChains()
	s.NoError(err)
	err = s.configurer.RunSetup()
	s.NoError(err)
}

func (s *BTCStakingTestSuite) TearDownSuite() {
	err := s.configurer.ClearResources()
	s.Require().NoError(err)
}

// TestCreateFinalityProviderAndDelegation is an end-to-end test for
// user story 1: user creates finality provider and BTC delegation
func (s *BTCStakingTestSuite) Test1CreateFinalityProviderAndDelegation() {
	chainA := s.configurer.GetChainConfig(0)
	chainA.WaitUntilHeight(1)
	nonValidatorNode, err := chainA.GetNodeAtIndex(2)
	s.NoError(err)

	/*
		create a random finality provider on Babylon
	*/
	// NOTE: we use the node's secret key as Babylon secret key for the finality provider
	msr, _, err = eots.NewMasterRandPair(r)
	s.NoError(err)
	fp, err = datagen.GenRandomCustomFinalityProvider(r, fpBTCSK, nonValidatorNode.SecretKey, msr)
	s.NoError(err)
	nonValidatorNode.CreateFinalityProvider(fp.BabylonPk, fp.BtcPk, fp.Pop, fp.MasterPubRand, fp.Description.Moniker, fp.Description.Identity, fp.Description.Website, fp.Description.SecurityContact, fp.Description.Details, fp.Commission)

	// wait for a block so that above txs take effect
	nonValidatorNode.WaitForNextBlock()

	// query the existence of finality provider and assert equivalence
	actualFps := nonValidatorNode.QueryFinalityProviders()
	s.Len(actualFps, 1)
	fp.RegisteredEpoch = actualFps[0].RegisteredEpoch // remember registered epoch
	s.equalFinalityProviderResp(fp, actualFps[0])

	/*
		create a random BTC delegation under this finality provider
	*/
	// BTC staking params, BTC delegation key pairs and PoP
	params := nonValidatorNode.QueryBTCStakingParams()

	// minimal required unbonding time
	unbondingTime := uint16(initialization.BabylonBtcFinalizationPeriod) + 1

	// get covenant BTC PKs
	covenantBTCPKs := []*btcec.PublicKey{}
	for _, covenantPK := range params.CovenantPks {
		covenantBTCPKs = append(covenantBTCPKs, covenantPK.MustToBTCPK())
	}
	// NOTE: we use the node's secret key as Babylon secret key for the BTC delegation
	delBabylonSK := nonValidatorNode.SecretKey
	pop, err := bstypes.NewPoP(delBabylonSK, delBTCSK)
	s.NoError(err)
	// generate staking tx and slashing tx
	stakingTimeBlocks := uint16(math.MaxUint16)
	testStakingInfo := datagen.GenBTCStakingSlashingInfo(
		r,
		s.T(),
		net,
		delBTCSK,
		[]*btcec.PublicKey{fp.BtcPk.MustToBTCPK()},
		covenantBTCPKs,
		covenantQuorum,
		stakingTimeBlocks,
		stakingValue,
		params.SlashingAddress,
		params.SlashingRate,
		unbondingTime,
	)

	stakingMsgTx := testStakingInfo.StakingTx
	stakingTxHash := stakingMsgTx.TxHash().String()
	stakingSlashingPathInfo, err := testStakingInfo.StakingInfo.SlashingPathSpendInfo()
	s.NoError(err)

	// generate proper delegator sig
	delegatorSig, err := testStakingInfo.SlashingTx.Sign(
		stakingMsgTx,
		datagen.StakingOutIdx,
		stakingSlashingPathInfo.GetPkScriptPath(),
		delBTCSK,
	)
	s.NoError(err)

	// submit staking tx to Bitcoin and get inclusion proof
	currentBtcTipResp, err := nonValidatorNode.QueryTip()
	s.NoError(err)
	currentBtcTip, err := chain.ParseBTCHeaderInfoResponseToInfo(currentBtcTipResp)
	s.NoError(err)

	blockWithStakingTx := datagen.CreateBlockWithTransaction(r, currentBtcTip.Header.ToBlockHeader(), stakingMsgTx)
	nonValidatorNode.InsertHeader(&blockWithStakingTx.HeaderBytes)
	// make block k-deep
	for i := 0; i < initialization.BabylonBtcConfirmationPeriod; i++ {
		nonValidatorNode.InsertNewEmptyBtcHeader(r)
	}
	stakingTxInfo := btcctypes.NewTransactionInfoFromSpvProof(blockWithStakingTx.SpvProof)

	// generate BTC undelegation stuff
	stkTxHash := testStakingInfo.StakingTx.TxHash()
	unbondingValue := stakingValue - datagen.UnbondingTxFee // TODO: parameterise fee
	testUnbondingInfo := datagen.GenBTCUnbondingSlashingInfo(
		r,
		s.T(),
		net,
		delBTCSK,
		[]*btcec.PublicKey{fp.BtcPk.MustToBTCPK()},
		covenantBTCPKs,
		covenantQuorum,
		wire.NewOutPoint(&stkTxHash, datagen.StakingOutIdx),
		stakingTimeBlocks,
		unbondingValue,
		params.SlashingAddress,
		params.SlashingRate,
		unbondingTime,
	)
	delUnbondingSlashingSig, err := testUnbondingInfo.GenDelSlashingTxSig(delBTCSK)
	s.NoError(err)

	// submit the message for creating BTC delegation
	nonValidatorNode.CreateBTCDelegation(
		delBabylonSK.PubKey().(*secp256k1.PubKey),
		bbn.NewBIP340PubKeyFromBTCPK(delBTCPK),
		pop,
		stakingTxInfo,
		fp.BtcPk,
		stakingTimeBlocks,
		btcutil.Amount(stakingValue),
		testStakingInfo.SlashingTx,
		delegatorSig,
		testUnbondingInfo.UnbondingTx,
		testUnbondingInfo.SlashingTx,
		uint16(unbondingTime),
		btcutil.Amount(unbondingValue),
		delUnbondingSlashingSig,
	)

	// wait for a block so that above txs take effect
	nonValidatorNode.WaitForNextBlock()
	nonValidatorNode.WaitForNextBlock()

	pendingDelSet := nonValidatorNode.QueryFinalityProviderDelegations(fp.BtcPk.MarshalHex())
	s.Len(pendingDelSet, 1)
	pendingDels := pendingDelSet[0]
	s.Len(pendingDels.Dels, 1)
	s.Equal(delBTCPK.SerializeCompressed()[1:], pendingDels.Dels[0].BtcPk.MustToBTCPK().SerializeCompressed()[1:])
	s.Len(pendingDels.Dels[0].CovenantSigs, 0)

	// check delegation
	delegation := nonValidatorNode.QueryBtcDelegation(stakingTxHash)
	s.NotNil(delegation)
}

// Test2SubmitCovenantSignature is an end-to-end test for user
// story 2: covenant approves the BTC delegation
func (s *BTCStakingTestSuite) Test2SubmitCovenantSignature() {
	chainA := s.configurer.GetChainConfig(0)
	chainA.WaitUntilHeight(1)
	nonValidatorNode, err := chainA.GetNodeAtIndex(2)
	s.NoError(err)

	// get last BTC delegation
	pendingDelsSet := nonValidatorNode.QueryFinalityProviderDelegations(fp.BtcPk.MarshalHex())
	s.Len(pendingDelsSet, 1)
	pendingDels := pendingDelsSet[0]
	s.Len(pendingDels.Dels, 1)
	pendingDelResp := pendingDels.Dels[0]
	pendingDel, err := ParseRespBTCDelToBTCDel(pendingDelResp)
	s.NoError(err)
	s.Len(pendingDel.CovenantSigs, 0)

	slashingTx := pendingDel.SlashingTx
	stakingTx := pendingDel.StakingTx

	stakingMsgTx, err := bbn.NewBTCTxFromBytes(stakingTx)
	s.NoError(err)
	stakingTxHash := stakingMsgTx.TxHash().String()

	params := nonValidatorNode.QueryBTCStakingParams()

	fpBTCPKs, err := bbn.NewBTCPKsFromBIP340PKs(pendingDel.FpBtcPkList)
	s.NoError(err)

	stakingInfo, err := pendingDel.GetStakingInfo(params, net)
	s.NoError(err)

	stakingSlashingPathInfo, err := stakingInfo.SlashingPathSpendInfo()
	s.NoError(err)

	/*
		generate and insert new covenant signature, in order to activate the BTC delegation
	*/
	// covenant signatures on slashing tx
	covenantSlashingSigs, err := datagen.GenCovenantAdaptorSigs(
		covenantSKs,
		fpBTCPKs,
		stakingMsgTx,
		stakingSlashingPathInfo.GetPkScriptPath(),
		slashingTx,
	)
	s.NoError(err)

	// cov Schnorr sigs on unbonding signature
	unbondingPathInfo, err := stakingInfo.UnbondingPathSpendInfo()
	s.NoError(err)
	unbondingTx, err := bbn.NewBTCTxFromBytes(pendingDel.BtcUndelegation.UnbondingTx)
	s.NoError(err)

	covUnbondingSigs, err := datagen.GenCovenantUnbondingSigs(
		covenantSKs,
		stakingMsgTx,
		pendingDel.StakingOutputIdx,
		unbondingPathInfo.GetPkScriptPath(),
		unbondingTx,
	)
	s.NoError(err)

	unbondingInfo, err := pendingDel.GetUnbondingInfo(params, net)
	s.NoError(err)
	unbondingSlashingPathInfo, err := unbondingInfo.SlashingPathSpendInfo()
	s.NoError(err)
	covenantUnbondingSlashingSigs, err := datagen.GenCovenantAdaptorSigs(
		covenantSKs,
		fpBTCPKs,
		unbondingTx,
		unbondingSlashingPathInfo.GetPkScriptPath(),
		pendingDel.BtcUndelegation.SlashingTx,
	)
	s.NoError(err)

	for i := 0; i < int(covenantQuorum); i++ {
		nonValidatorNode.AddCovenantSigs(
			covenantSlashingSigs[i].CovPk,
			stakingTxHash,
			covenantSlashingSigs[i].AdaptorSigs,
			bbn.NewBIP340SignatureFromBTCSig(covUnbondingSigs[i]),
			covenantUnbondingSlashingSigs[i].AdaptorSigs,
		)
		// wait for a block so that above txs take effect
		nonValidatorNode.WaitForNextBlock()
	}

	// wait for a block so that above txs take effect
	nonValidatorNode.WaitForNextBlock()
	nonValidatorNode.WaitForNextBlock()

	// ensure the BTC delegation has covenant sigs now
	activeDelsSet := nonValidatorNode.QueryFinalityProviderDelegations(fp.BtcPk.MarshalHex())
	s.Len(activeDelsSet, 1)

	activeDels, err := ParseRespsBTCDelToBTCDel(activeDelsSet[0])
	s.NoError(err)
	s.NotNil(activeDels)
	s.Len(activeDels.Dels, 1)

	activeDel := activeDels.Dels[0]
	s.True(activeDel.HasCovenantQuorums(covenantQuorum))

	// wait for a block so that above txs take effect and the voting power table
	// is updated in the next block's BeginBlock
	nonValidatorNode.WaitForNextBlock()

	// ensure BTC staking is activated
	activatedHeight := nonValidatorNode.QueryActivatedHeight()
	s.Positive(activatedHeight)
	// ensure finality provider has voting power at activated height
	currentBtcTip, err := nonValidatorNode.QueryTip()
	s.NoError(err)
	activeFps := nonValidatorNode.QueryActiveFinalityProvidersAtHeight(activatedHeight)
	s.Len(activeFps, 1)
	s.Equal(activeFps[0].VotingPower, activeDels.VotingPower(currentBtcTip.Height, initialization.BabylonBtcFinalizationPeriod, params.CovenantQuorum))
	s.Equal(activeFps[0].VotingPower, activeDel.VotingPower(currentBtcTip.Height, initialization.BabylonBtcFinalizationPeriod, params.CovenantQuorum))
}

// Test3SubmitFinalitySignature is an end-to-end test for user story 3:
// finality provider and submits finality signature, such that blocks can be finalised.
func (s *BTCStakingTestSuite) Test3SubmitFinalitySignature() {
	chainA := s.configurer.GetChainConfig(0)
	chainA.WaitUntilHeight(1)
	nonValidatorNode, err := chainA.GetNodeAtIndex(2)
	s.NoError(err)

	// finalise epochs until the registered epoch of the finality provider
	// so that the finality provider can vote
	var (
		startEpoch = uint64(1)
		endEpoch   = fp.RegisteredEpoch
	)
	nonValidatorNode.FinalizeSealedEpochs(startEpoch, endEpoch)

	// get activated height
	activatedHeight := nonValidatorNode.QueryActivatedHeight()
	s.Positive(activatedHeight)

	// no reward gauge for finality provider yet
	fpBabylonAddr := sdk.AccAddress(nonValidatorNode.SecretKey.PubKey().Address().Bytes())
	fpRewardGauges, err := nonValidatorNode.QueryRewardGauge(fpBabylonAddr)
	s.NoError(err)
	_, ok := fpRewardGauges[itypes.FinalityProviderType.String()]
	s.False(ok)

	// no reward gauge for BTC delegator yet
	delBabylonAddr := sdk.AccAddress(nonValidatorNode.SecretKey.PubKey().Address().Bytes())
	btcDelRewardGauges, err := nonValidatorNode.QueryRewardGauge(delBabylonAddr)
	s.NoError(err)
	_, ok = btcDelRewardGauges[itypes.BTCDelegationType.String()]
	s.False(ok)

	/*
		generate finality signature
	*/
	// get block to vote
	blockToVote, err := nonValidatorNode.QueryBlock(int64(activatedHeight))
	s.NoError(err)
	appHash := blockToVote.AppHash
	msgToSign := append(sdk.Uint64ToBigEndian(activatedHeight), appHash...)
	// generate EOTS signature
	sr, _, err := msr.DeriveRandPair(uint32(activatedHeight))
	s.NoError(err)
	sig, err := eots.Sign(fpBTCSK, sr, msgToSign)
	s.NoError(err)
	eotsSig := bbn.NewSchnorrEOTSSigFromModNScalar(sig)

	/*
		submit finality signature
	*/
	// submit finality signature
	nonValidatorNode.AddFinalitySig(fp.BtcPk, activatedHeight, appHash, eotsSig)

	// ensure vote is eventually cast
	nonValidatorNode.WaitForNextBlock()
	var votes []bbn.BIP340PubKey
	s.Eventually(func() bool {
		votes = nonValidatorNode.QueryVotesAtHeight(activatedHeight)
		return len(votes) > 0
	}, time.Minute, time.Second*5)
	s.Equal(1, len(votes))
	s.Equal(votes[0].MarshalHex(), fp.BtcPk.MarshalHex())
	// once the vote is cast, ensure block is finalised
	finalizedBlock := nonValidatorNode.QueryIndexedBlock(activatedHeight)
	s.NotEmpty(finalizedBlock)
	s.Equal(appHash.Bytes(), finalizedBlock.AppHash)
	finalizedBlocks := nonValidatorNode.QueryListBlocks(ftypes.QueriedBlockStatus_FINALIZED)
	s.NotEmpty(finalizedBlocks)
	s.Equal(appHash.Bytes(), finalizedBlocks[0].AppHash)

	// ensure finality provider has received rewards after the block is finalised
	fpRewardGauges, err = nonValidatorNode.QueryRewardGauge(fpBabylonAddr)
	s.NoError(err)
	fpRewardGauge, ok := fpRewardGauges[itypes.FinalityProviderType.String()]
	s.True(ok)
	s.True(fpRewardGauge.Coins.IsAllPositive())

	// ensure BTC delegation has received rewards after the block is finalised
	btcDelRewardGauges, err = nonValidatorNode.QueryRewardGauge(delBabylonAddr)
	s.NoError(err)
	btcDelRewardGauge, ok := btcDelRewardGauges[itypes.BTCDelegationType.String()]
	s.True(ok)
	s.True(btcDelRewardGauge.Coins.IsAllPositive())
}

func (s *BTCStakingTestSuite) Test4WithdrawReward() {
	chainA := s.configurer.GetChainConfig(0)
	nonValidatorNode, err := chainA.GetNodeAtIndex(2)
	s.NoError(err)

	// finality provider balance before withdraw
	fpBabylonAddr := sdk.AccAddress(nonValidatorNode.SecretKey.PubKey().Address().Bytes())
	delBabylonAddr := sdk.AccAddress(nonValidatorNode.SecretKey.PubKey().Address().Bytes())
	fpBalance, err := nonValidatorNode.QueryBalances(fpBabylonAddr.String())
	s.NoError(err)
	// finality provider reward gauge should not be fully withdrawn
	fpRgs, err := nonValidatorNode.QueryRewardGauge(fpBabylonAddr)
	s.NoError(err)
	fpRg := fpRgs[itypes.FinalityProviderType.String()]
	s.T().Logf("finality provider's withdrawable reward before withdrawing: %s", fpRg.GetWithdrawableCoins().String())
	s.False(fpRg.IsFullyWithdrawn())

	// withdraw finality provider reward
	nonValidatorNode.WithdrawReward(itypes.FinalityProviderType.String(), initialization.ValidatorWalletName)
	nonValidatorNode.WaitForNextBlock()

	// balance after withdrawing finality provider reward
	fpBalance2, err := nonValidatorNode.QueryBalances(fpBabylonAddr.String())
	s.NoError(err)
	s.T().Logf("fpBalance2: %s; fpBalance: %s", fpBalance2.String(), fpBalance.String())
	s.True(fpBalance2.IsAllGT(fpBalance))
	// finality provider reward gauge should be fully withdrawn now
	fpRgs2, err := nonValidatorNode.QueryRewardGauge(fpBabylonAddr)
	s.NoError(err)
	fpRg2 := fpRgs2[itypes.FinalityProviderType.String()]
	s.T().Logf("finality provider's withdrawable reward after withdrawing: %s", fpRg2.GetWithdrawableCoins().String())
	s.True(fpRg2.IsFullyWithdrawn())

	// BTC delegation balance before withdraw
	btcDelBalance, err := nonValidatorNode.QueryBalances(delBabylonAddr.String())
	s.NoError(err)
	// BTC delegation reward gauge should not be fully withdrawn
	btcDelRgs, err := nonValidatorNode.QueryRewardGauge(delBabylonAddr)
	s.NoError(err)
	btcDelRg := btcDelRgs[itypes.BTCDelegationType.String()]
	s.T().Logf("BTC delegation's withdrawable reward before withdrawing: %s", btcDelRg.GetWithdrawableCoins().String())
	s.False(btcDelRg.IsFullyWithdrawn())

	// withdraw BTC delegation reward
	nonValidatorNode.WithdrawReward(itypes.BTCDelegationType.String(), initialization.ValidatorWalletName)
	nonValidatorNode.WaitForNextBlock()

	// balance after withdrawing BTC delegation reward
	btcDelBalance2, err := nonValidatorNode.QueryBalances(delBabylonAddr.String())
	s.NoError(err)
	s.T().Logf("btcDelBalance2: %s; btcDelBalance: %s", btcDelBalance2.String(), btcDelBalance.String())
	s.True(btcDelBalance2.IsAllGT(btcDelBalance))
	// BTC delegation reward gauge should be fully withdrawn now
	btcDelRgs2, err := nonValidatorNode.QueryRewardGauge(delBabylonAddr)
	s.NoError(err)
	btcDelRg2 := btcDelRgs2[itypes.BTCDelegationType.String()]
	s.T().Logf("BTC delegation's withdrawable reward after withdrawing: %s", btcDelRg2.GetWithdrawableCoins().String())
	s.True(btcDelRg2.IsFullyWithdrawn())
}

// Test5SubmitStakerUnbonding is an end-to-end test for user unbonding
func (s *BTCStakingTestSuite) Test5SubmitStakerUnbonding() {
	chainA := s.configurer.GetChainConfig(0)
	chainA.WaitUntilHeight(1)
	nonValidatorNode, err := chainA.GetNodeAtIndex(2)
	s.NoError(err)
	// wait for a block so that above txs take effect
	nonValidatorNode.WaitForNextBlock()

	activeDelsSet := nonValidatorNode.QueryFinalityProviderDelegations(fp.BtcPk.MarshalHex())
	s.Len(activeDelsSet, 1)
	activeDels := activeDelsSet[0]
	s.Len(activeDels.Dels, 1)
	activeDelResp := activeDels.Dels[0]
	activeDel, err := ParseRespBTCDelToBTCDel(activeDelResp)
	s.NoError(err)
	s.NotNil(activeDel.CovenantSigs)

	// staking tx hash
	stakingMsgTx, err := bbn.NewBTCTxFromBytes(activeDel.StakingTx)
	s.NoError(err)
	stakingTxHash := stakingMsgTx.TxHash()

	// delegator signs unbonding tx
	params := nonValidatorNode.QueryBTCStakingParams()
	delUnbondingSig, err := activeDel.SignUnbondingTx(params, net, delBTCSK)
	s.NoError(err)

	// submit the message for creating BTC undelegation
	nonValidatorNode.BTCUndelegate(&stakingTxHash, delUnbondingSig)
	// wait for a block so that above txs take effect
	nonValidatorNode.WaitForNextBlock()

	// Wait for unbonded delegations to be created
	var unbondedDelsResp []*bstypes.BTCDelegationResponse
	s.Eventually(func() bool {
		unbondedDelsResp = nonValidatorNode.QueryUnbondedDelegations()
		return len(unbondedDelsResp) > 0
	}, time.Minute, time.Second*2)

	unbondDel, err := ParseRespBTCDelToBTCDel(unbondedDelsResp[0])
	s.NoError(err)
	s.Equal(stakingTxHash, unbondDel.MustGetStakingTxHash())
}

// ParseRespsBTCDelToBTCDel parses an BTC delegation response to BTC Delegation
func ParseRespsBTCDelToBTCDel(resp *bstypes.BTCDelegatorDelegationsResponse) (btcDels *bstypes.BTCDelegatorDelegations, err error) {
	if resp == nil {
		return nil, nil
	}
	btcDels = &bstypes.BTCDelegatorDelegations{
		Dels: make([]*bstypes.BTCDelegation, len(resp.Dels)),
	}

	for i, delResp := range resp.Dels {
		del, err := ParseRespBTCDelToBTCDel(delResp)
		if err != nil {
			return nil, err
		}
		btcDels.Dels[i] = del
	}
	return btcDels, nil
}

// ParseRespBTCDelToBTCDel parses an BTC delegation response to BTC Delegation
func ParseRespBTCDelToBTCDel(resp *bstypes.BTCDelegationResponse) (btcDel *bstypes.BTCDelegation, err error) {
	stakingTx, err := hex.DecodeString(resp.StakingTxHex)
	if err != nil {
		return nil, err
	}

	delSig, err := bbn.NewBIP340SignatureFromHex(resp.DelegatorSlashSigHex)
	if err != nil {
		return nil, err
	}

	slashingTx, err := bstypes.NewBTCSlashingTxFromHex(resp.SlashingTxHex)
	if err != nil {
		return nil, err
	}

	btcDel = &bstypes.BTCDelegation{
		// missing BabylonPk, Pop
		// these fields are not sent out to the client on BTCDelegationResponse
		BtcPk:            resp.BtcPk,
		FpBtcPkList:      resp.FpBtcPkList,
		StartHeight:      resp.StartHeight,
		EndHeight:        resp.EndHeight,
		TotalSat:         resp.TotalSat,
		StakingTx:        stakingTx,
		DelegatorSig:     delSig,
		StakingOutputIdx: resp.StakingOutputIdx,
		CovenantSigs:     resp.CovenantSigs,
		UnbondingTime:    resp.UnbondingTime,
		SlashingTx:       slashingTx,
	}

	if resp.UndelegationResponse != nil {
		ud := resp.UndelegationResponse
		unbondTx, err := hex.DecodeString(ud.UnbondingTxHex)
		if err != nil {
			return nil, err
		}

		slashTx, err := bstypes.NewBTCSlashingTxFromHex(ud.SlashingTxHex)
		if err != nil {
			return nil, err
		}

		delSlashingSig, err := bbn.NewBIP340SignatureFromHex(ud.DelegatorSlashingSigHex)
		if err != nil {
			return nil, err
		}

		btcDel.BtcUndelegation = &bstypes.BTCUndelegation{
			UnbondingTx:              unbondTx,
			CovenantUnbondingSigList: ud.CovenantUnbondingSigList,
			CovenantSlashingSigs:     ud.CovenantSlashingSigs,
			SlashingTx:               slashTx,
			DelegatorSlashingSig:     delSlashingSig,
		}

		if len(ud.DelegatorUnbondingSigHex) > 0 {
			delUnbondingSig, err := bbn.NewBIP340SignatureFromHex(ud.DelegatorUnbondingSigHex)
			if err != nil {
				return nil, err
			}
			btcDel.BtcUndelegation.DelegatorUnbondingSig = delUnbondingSig
		}
	}

	return btcDel, nil
}

func (s *BTCStakingTestSuite) equalFinalityProviderResp(fp *bstypes.FinalityProvider, fpResp *bstypes.FinalityProviderResponse) {
	s.Equal(fp.Description, fpResp.Description)
	s.Equal(fp.Commission, fpResp.Commission)
	s.Equal(fp.BabylonPk, fpResp.BabylonPk)
	s.Equal(fp.BtcPk, fpResp.BtcPk)
	s.Equal(fp.Pop, fpResp.Pop)
	s.Equal(fp.MasterPubRand, fpResp.MasterPubRand)
	s.Equal(fp.RegisteredEpoch, fpResp.RegisteredEpoch)
	s.Equal(fp.SlashedBabylonHeight, fpResp.SlashedBabylonHeight)
	s.Equal(fp.SlashedBtcHeight, fpResp.SlashedBtcHeight)
}
