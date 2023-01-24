package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/babylonchain/babylon/x/btccheckpoint/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	// Group btccheckpoint queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdQueryParams())

	cmd.AddCommand(CmdBtcCheckpointHeightAndHash())
	cmd.AddCommand(CmdEpochSubmissions())
	return cmd
}

func CmdBtcCheckpointHeightAndHash() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "btc-height-hash <epoch_number>",
		Short: "retrieve earliest btc height and hash for given epoch",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			epoch_num, err := strconv.ParseUint(args[0], 10, 64)

			if err != nil {
				return err
			}

			req := types.QueryBtcCheckpointHeightAndHashRequest{EpochNum: epoch_num}

			resp, err := queryClient.BtcCheckpointHeightAndHash(context.Background(), &req)

			if err != nil {
				return err
			}

			return clientCtx.PrintProto(resp)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdEpochSubmissions() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "epoch-submissions <epoch_number>",
		Short: "all checkpoint submissions for given epoch",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			epoch_num, err := strconv.ParseUint(args[0], 10, 64)

			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			params := types.QueryEpochSubmissionsRequest{EpochNum: epoch_num, Pagination: pageReq}

			res, err := queryClient.EpochSubmissions(context.Background(), &params)

			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
