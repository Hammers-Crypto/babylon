//go:build e2e
// +build e2e

package e2e

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"time"

	"github.com/babylonchain/babylon/test/e2e/configurer/config"
	"github.com/babylonchain/babylon/test/e2e/initialization"
	bbn "github.com/babylonchain/babylon/types"
	ct "github.com/babylonchain/babylon/x/checkpointing/types"
	zctypes "github.com/babylonchain/babylon/x/zoneconcierge/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/stretchr/testify/require"
)

// Most simple test, just checking that two chains are up and connected through
// ibc
func (s *IntegrationTestSuite) TestConnectIbc() {
	chainA := s.configurer.GetChainConfig(0)
	chainB := s.configurer.GetChainConfig(1)
	_, err := chainA.GetDefaultNode()
	s.NoError(err)
	_, err = chainB.GetDefaultNode()
	s.NoError(err)
}

func (s *IntegrationTestSuite) TestBTCBaseHeader() {
	hardcodedHeader, _ := bbn.NewBTCHeaderBytesFromHex("0100000000000000000000000000000000000000000000000000000000000000000000003ba3edfd7a7b12b27ac72c3e67768f617fc81bc3888a51323a9fb8aa4b1e5e4a45068653ffff7f2002000000")
	hardcodedHeaderHeight := uint64(0)

	chainA := s.configurer.GetChainConfig(0)
	nonValidatorNode, err := chainA.GetNodeAtIndex(2)
	s.NoError(err)
	baseHeader, err := nonValidatorNode.QueryBtcBaseHeader()
	s.True(baseHeader.Hash.Eq(hardcodedHeader.Hash()))
	s.Equal(hardcodedHeaderHeight, baseHeader.Height)
}

func (s *IntegrationTestSuite) TestEpochInterval() {
	chainA := s.configurer.GetChainConfig(0)
	nonValidatorNode, err := chainA.GetNodeAtIndex(2)
	s.NoError(err)

	epochInterval, err := nonValidatorNode.QueryEpochInterval()
	s.NoError(err)
	s.Equal(epochInterval, uint64(initialization.BabylonEpochInterval))
}

func (s *IntegrationTestSuite) TestSendTx() {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	chainA := s.configurer.GetChainConfig(0)
	nonValidatorNode, err := chainA.GetNodeAtIndex(2)
	s.NoError(err)

	tip1, err := nonValidatorNode.QueryTip()
	s.NoError(err)

	nonValidatorNode.InsertNewEmptyBtcHeader(r)

	tip2, err := nonValidatorNode.QueryTip()
	s.NoError(err)

	s.Equal(tip1.Height+1, tip2.Height)
}

func (s *IntegrationTestSuite) TestPhase1_IbcCheckpointing() {
	endEpochNum := uint64(3)

	chainA := s.configurer.GetChainConfig(0)
	nonValidatorNode, err := chainA.GetNodeAtIndex(2)
	s.NoError(err)
	nonValidatorNode.WaitUntilHeight(int64(initialization.BabylonEpochInterval*endEpochNum + 3))

	// Query checkpoint chain info for opposing chain
	chainsInfo, err := nonValidatorNode.QueryChainsInfo([]string{initialization.ChainBID})
	s.NoError(err)
	s.Equal(chainsInfo[0].ChainId, initialization.ChainBID)

	nonValidatorNode.FinalizeSealedEpochs(1, endEpochNum)

	endEpoch, err := nonValidatorNode.QueryRawCheckpoint(endEpochNum)
	s.NoError(err)
	s.Equal(endEpoch.Status, ct.Finalized)

	// Check we have epoch info for opposing chain and some basic assertions
	epochChainsInfo, err := nonValidatorNode.QueryEpochChainsInfo(endEpochNum, []string{initialization.ChainBID})
	s.NoError(err)
	s.Equal(epochChainsInfo[0].ChainId, initialization.ChainBID)
	s.Equal(epochChainsInfo[0].LatestHeader.BabylonEpoch, endEpochNum)

	// Check we have finalized epoch info for opposing chain and some basic assertions
	finalizedChainsInfo, err := nonValidatorNode.QueryFinalizedChainsInfo([]string{initialization.ChainBID})
	s.NoError(err)

	// TODO Add more assertion here. Maybe check proofs ?
	s.Equal(finalizedChainsInfo[0].FinalizedChainInfo.ChainId, initialization.ChainBID)
	s.Equal(finalizedChainsInfo[0].EpochInfo.EpochNumber, endEpochNum)

	currEpoch, err := nonValidatorNode.QueryCurrentEpoch()
	s.NoError(err)

	heightAtEndedEpoch, err := nonValidatorNode.QueryLightClientHeightEpochEnd(currEpoch - 1)
	s.NoError(err)
	s.Greater(heightAtEndedEpoch, uint64(0), fmt.Sprintf("Light client height should be  > 0 on epoch %d", currEpoch-1))

	chainB := s.configurer.GetChainConfig(1)
	_, err = chainB.GetDefaultNode()
	s.NoError(err)
}

func (s *IntegrationTestSuite) TestPhase2_BabylonContract() {
	// chain A
	chainA := s.configurer.GetChainConfig(0)
	nonValidatorNode, err := chainA.GetNodeAtIndex(2)
	s.NoError(err)

	// chain B
	chainB := s.configurer.GetChainConfig(1)

	// deploy Babylon contract at chain B
	contractPath := "/bytecode/babylon_contract.wasm"
	initMsg := fmt.Sprintf(`{"btc_confirmation_depth":%d,"checkpoint_finalization_timeout":%d,"network":"Regtest","babylon_tag":[1,2,3,4],"notify_cosmos_zone":false}`, initialization.BabylonBtcConfirmationPeriod, initialization.BabylonBtcFinalizationPeriod)
	contractAddr, err := s.configurer.DeployWasmContract(contractPath, chainB, initMsg)
	s.NoError(err)

	// establish IBC channel between chain A ZoneConcierge and chain B Babylon contract
	channelCfg := config.NewIBCChannelConfigWithBabylonContract(chainA.Id, chainB.Id, contractAddr)
	err = s.configurer.ConnectIBCChains(channelCfg)
	s.NoError(err)

	// there should be 2 open channels, one connecting to chain B ZoneConcierge, the other connecting to chain B Babylon contract
	channels := nonValidatorNode.QueryIBCChannels()
	openedChannels := []*channeltypes.IdentifiedChannel{}
	for _, channel := range channels {
		if channel.State == channeltypes.OPEN {
			openedChannels = append(openedChannels, channel)
		}
	}
	s.Len(openedChannels, 2)

	// Finalize 3 new epochs
	startEpochNum := uint64(4)
	endEpochNum := uint64(6)
	nonValidatorNode.WaitUntilHeight(int64(initialization.BabylonEpochInterval*endEpochNum + 3))
	nonValidatorNode.FinalizeSealedEpochs(startEpochNum, endEpochNum)
	nonValidatorNode.WaitForNextBlock()

	// chain A must be sending 1 IBC packet to the IBC channel with Babylon contract
	nextSeq, err := nonValidatorNode.QueryIBCNextSequence(channels[1].ChannelId, zctypes.PortID)
	s.NoError(err)
	s.Equal(uint64(1), nextSeq)
}

func (s *IntegrationTestSuite) TestWasm() {
	// deploy the storage contract
	contractPath := "/bytecode/storage_contract.wasm"
	chainA := s.configurer.GetChainConfig(0)
	initMsg := `{}`
	contractAddr, err := s.configurer.DeployWasmContract(contractPath, chainA, initMsg)
	require.NoError(s.T(), err)

	data := []byte{1, 2, 3, 4, 5}
	dataHex := hex.EncodeToString(data)
	dataHash := sha256.Sum256(data)
	dataHashHex := hex.EncodeToString(dataHash[:])

	storeMsg := fmt.Sprintf(`{"save_data":{"data":"%s"}}`, dataHex)
	nonValidatorNode, err := chainA.GetNodeAtIndex(2)
	nonValidatorNode.WasmExecute(contractAddr, storeMsg, initialization.ValidatorWalletName)
	nonValidatorNode.WaitForNextBlock()
	queryMsg := fmt.Sprintf(`{"check_data": {"data_hash":"%s"}}`, dataHashHex)
	queryResult, err := nonValidatorNode.QueryWasmSmartObject(contractAddr, queryMsg)
	require.NoError(s.T(), err)
	finalized := queryResult["finalized"].(bool)
	latestFinalizedEpoch := int(queryResult["latest_finalized_epoch"].(float64))
	saveEpoch := int(queryResult["save_epoch"].(float64))

	require.False(s.T(), finalized)
	// in previous test we already finalized epoch 6
	require.Equal(s.T(), 6, latestFinalizedEpoch)
	// data is not finalized yet, so save epoch should be strictly greater than latest finalized epoch
	require.Greater(s.T(), saveEpoch, latestFinalizedEpoch)
}
