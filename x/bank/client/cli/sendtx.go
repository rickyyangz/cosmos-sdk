package cli

import (
	"fmt"
	errs "errors"

	"github.com/pkg/errors"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/bank/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	flagTo     = "to"
	flagAmount = "amount"
)

// SendTxCmd will create a send tx and sign it with the given key
func SendTxCmd(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Create and sign a send tx",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))

			// get the from/to address
			from, err := ctx.GetFromAddresses()
			if err != nil {
				return err
			}
			if len(from) != 1 {
				return errs.New("Must provide single from address for this transaction")
			}

			fmt.Println(from)
			fmt.Println(ctx.AccountNumbers)

			fromAcc, err := ctx.QueryStore(auth.AddressStoreKey(from[0]), ctx.AccountStore)
			if err != nil {
				return err
			}

			// Check if account was found
			if fromAcc == nil {
				return errors.Errorf("No account with address %s was found in the state.\nAre you sure there has been a transaction involving it?", from)
			}

			toStr := viper.GetString(flagTo)

			to, err := sdk.AccAddressFromBech32(toStr)
			if err != nil {
				return err
			}
			// parse coins trying to be sent
			amount := viper.GetString(flagAmount)
			coins, err := sdk.ParseCoins(amount)
			if err != nil {
				return err
			}

			// ensure account has enough coins
			account, err := ctx.Decoder(fromAcc)
			if err != nil {
				return err
			}
			if !account.GetCoins().IsGTE(coins) {
				return errors.Errorf("Address %s doesn't have enough coins to pay for this transaction.", from)
			}

			// build and sign the transaction, then broadcast to Tendermint
			msg := client.BuildMsg(from[0], to, coins)

			err = ctx.EnsureSignBuildBroadcast(ctx.FromAddressNames, []sdk.Msg{msg}, cdc)
			if err != nil {
				return err
			}
			return nil

		},
	}

	cmd.Flags().String(flagTo, "", "Address to send coins")
	cmd.Flags().String(flagAmount, "", "Amount of coins to send")

	return cmd
}
