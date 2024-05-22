package cli

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	bbn "github.com/babylonchain/babylon/types"
	"github.com/babylonchain/babylon/x/finality/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		NewCommitPubRandListCmd(),
		NewAddFinalitySigCmd(),
	)

	return cmd
}

func NewCommitPubRandListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "commit-pubrand-list [fp_btc_pk] [start_height] [num_pub_rand] [commitment] [sig]",
		Args:  cobra.MinimumNArgs(5),
		Short: "Commit a list of public randomness",
		Long: strings.TrimSpace(
			`Commit a list of public randomness.`, // TODO: example
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// get finality provider BTC PK
			fpBTCPK, err := bbn.NewBIP340PubKeyFromHex(args[0])
			if err != nil {
				return err
			}

			// get start height
			startHeight, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			numPubRand, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return err
			}

			commitment, err := hex.DecodeString(args[3])
			if err != nil {
				return err
			}

			// get signature
			sig, err := bbn.NewBIP340SignatureFromHex(args[4])
			if err != nil {
				return err
			}

			msg := types.MsgCommitPubRandList{
				Signer:      clientCtx.FromAddress.String(),
				FpBtcPk:     fpBTCPK,
				StartHeight: startHeight,
				NumPubRand:  numPubRand,
				Commitment:  commitment,
				Sig:         sig,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewAddFinalitySigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-finality-sig [fp_btc_pk] [block_height] [block_app_hash] [finality_sig]",
		Args:  cobra.ExactArgs(4),
		Short: "Add a finality signature",
		Long: strings.TrimSpace(
			`Add a finality signature.`, // TODO: example
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// get finality provider BTC PK
			fpBTCPK, err := bbn.NewBIP340PubKeyFromHex(args[0])
			if err != nil {
				return err
			}

			// get block height
			blockHeight, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			// get block last commit hash
			blockLch, err := hex.DecodeString(args[2])
			if err != nil {
				return err
			}

			// get finality signature
			finalitySig, err := bbn.NewSchnorrEOTSSigFromHex(args[3])
			if err != nil {
				return err
			}

			msg := types.MsgAddFinalitySig{
				Signer:       clientCtx.FromAddress.String(),
				FpBtcPk:      fpBTCPK,
				BlockHeight:  blockHeight,
				BlockAppHash: blockLch,
				FinalitySig:  finalitySig,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
