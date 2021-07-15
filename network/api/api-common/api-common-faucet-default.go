// +build !wasm

package api_common

import (
	"encoding/json"
	"errors"
	"go.jolheiser.com/hcaptcha"
	"pandora-pay/config"
	"pandora-pay/network/api/api-common/api_types"
	transactions_builder "pandora-pay/transactions-builder"
	"pandora-pay/wallet"
)

type APICommonFaucet struct {
	hcpatchaClient      *hcaptcha.Client
	wallet              *wallet.Wallet
	transactionsBuilder *transactions_builder.TransactionsBuilder
}

func (api *APICommonFaucet) GetFaucetInfo() ([]byte, error) {
	return json.Marshal(&api_types.APIFaucetInfo{
		config.HCAPTCHA_SITE_KEY,
		config.FAUCET_TESTNET_ENABLED,
		config.FAUCET_TESTNET_COINS,
	})
}

func (api *APICommonFaucet) GetFaucetCoins(request *api_types.APIFaucetCoinsRequest) ([]byte, error) {

	if !config.FAUCET_TESTNET_ENABLED {
		return nil, errors.New("Faucet Testnet is not enabled")
	}

	resp, err := api.hcpatchaClient.Verify(request.FaucetToken, hcaptcha.PostOptions{})
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, errors.New("Faucet token is invalid")
	}

	addr, err := api.wallet.GetWalletAddress(0)
	if err != nil {
		return nil, err
	}

	tx, err := api.transactionsBuilder.CreateSimpleTx([]string{addr.AddressEncoded}, 0, []uint64{config.FAUCET_TESTNET_COINS_UNITS}, [][]byte{config.NATIVE_TOKEN}, []string{request.Address}, []uint64{config.FAUCET_TESTNET_COINS_UNITS}, [][]byte{config.NATIVE_TOKEN}, 0, []byte{})
	if err != nil {
		return nil, err
	}

	return tx.Bloom.Hash, nil
}

func createAPICommonFaucet(wallet *wallet.Wallet, transactionsBuilder *transactions_builder.TransactionsBuilder) (*APICommonFaucet, error) {

	api := &APICommonFaucet{
		nil, wallet, transactionsBuilder,
	}

	// Dummy secret https://docs.hcaptcha.com/#integrationtest
	hcpatchaClient, err := hcaptcha.New(config.HCAPTCHA_SECRET_KEY)
	if err != nil {
		return nil, err
	}

	api.hcpatchaClient = hcpatchaClient

	return api, nil
}
