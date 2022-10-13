package tx

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdktx "github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/pflag"
)

func SendMsgToTendermint(clientCtx client.Context, msg sdk.Msg) (*sdk.TxResponse, error) {
	return SendMsgsToTendermint(clientCtx, []sdk.Msg{msg})
}

func SendMsgsToTendermint(clientCtx client.Context, msgs []sdk.Msg) (*sdk.TxResponse, error) {
	for _, msg := range msgs {
		if err := msg.ValidateBasic(); err != nil {
			return nil, err
		}
	}

	// TODO make the fee it dynamic
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	fs.String(flags.FlagFees, "", "Fees to pay along with transaction; eg: 10ubbn")
	err := fs.Set(flags.FlagFees, "100stake")
	if err != nil {
		return nil, err
	}

	txf := sdktx.NewFactoryCLI(clientCtx, fs)

	return BroadcastTx(clientCtx, txf, msgs...)
}

// BroadcastTx attempts to generate, sign and broadcast a transaction with the
// given set of messages. It will also simulate gas requirements if necessary.
// It will return an error upon failure.
// The code is based on cosmos-sdk: https://github.com/cosmos/cosmos-sdk/blob/7781cdb3d20bc7ebac017452897ce1e6ab3903ef/client/tx/tx.go#L65
// it treats non-zero response code as errors
func BroadcastTx(clientCtx client.Context, txf sdktx.Factory, msgs ...sdk.Msg) (*sdk.TxResponse, error) {
	txf, err := prepareFactory(clientCtx, txf)
	if err != nil {
		return nil, err
	}

	tx, err := sdktx.BuildUnsignedTx(txf, msgs...)
	if err != nil {
		return nil, err
	}

	tx.SetFeeGranter(clientCtx.GetFeeGranterAddress())
	err = sdktx.Sign(txf, clientCtx.GetFromName(), tx, true)
	if err != nil {
		return nil, err
	}

	txBytes, err := clientCtx.TxConfig.TxEncoder()(tx.GetTx())
	if err != nil {
		return nil, err
	}

	// broadcast to a Tendermint node
	res, err := clientCtx.BroadcastTx(txBytes)
	if err != nil {
		return nil, err
	}

	// transaction was executed, log the success or failure using the tx response code
	// NOTE: error is nil, logic should use the returned error to determine if the
	// transaction was successfully executed.
	if res.Code != 0 {
		_ = clientCtx.PrintProto(res)
		return res, fmt.Errorf("transaction failed with code: %d", res.Code)
	}

	return res, nil
}

// prepareFactory ensures the account defined by ctx.GetFromAddress() exists and
// if the account number and/or the account sequence number are zero (not set),
// they will be queried for and set on the provided Factory. A new Factory with
// the updated fields will be returned.
func prepareFactory(clientCtx client.Context, txf sdktx.Factory) (sdktx.Factory, error) {
	from := clientCtx.GetFromAddress()

	if err := txf.AccountRetriever().EnsureExists(clientCtx, from); err != nil {
		return txf, err
	}

	initNum, initSeq := txf.AccountNumber(), txf.Sequence()
	if initNum == 0 || initSeq == 0 {
		num, seq, err := txf.AccountRetriever().GetAccountNumberSequence(clientCtx, from)
		if err != nil {
			return txf, err
		}

		if initNum == 0 {
			txf = txf.WithAccountNumber(num)
		}

		if initSeq == 0 {
			txf = txf.WithSequence(seq)
		}
	}

	return txf, nil
}
