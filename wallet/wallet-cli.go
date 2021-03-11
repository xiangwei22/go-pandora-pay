package wallet

import (
	"encoding/json"
	"fmt"
	bolt "go.etcd.io/bbolt"
	"os"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/config"
	"pandora-pay/gui"
	"pandora-pay/store"
	"strconv"
)

func initWalletCLI(wallet *Wallet) {

	cliListAddresses := func(cmd string) {

		gui.OutputWrite("Wallet")
		gui.OutputWrite("Version: " + wallet.Version.String())
		gui.OutputWrite("Encrypted: " + wallet.Encrypted.String())
		gui.OutputWrite("Count: " + strconv.Itoa(wallet.Count))

		gui.OutputWrite("")

		if err := store.StoreBlockchain.DB.View(func(tx *bolt.Tx) (err error) {

			accs := accounts.NewAccounts(tx)
			toks := tokens.NewTokens(tx)

			for _, walletAddress := range wallet.Addresses {
				addressStr := walletAddress.Address.EncodeAddr()
				gui.OutputWrite(walletAddress.Name + " : " + walletAddress.Address.Version.String() + " : " + addressStr)

				if walletAddress.Address.Version == addresses.SimplePublicKeyHash ||
					walletAddress.Address.Version == addresses.SimplePublicKey {

					acc := accs.GetAccount(&walletAddress.PublicKeyHash)

					if acc == nil {
						gui.OutputWrite(fmt.Sprintf("%18s: %s", "", "EMPTY"))
					} else {
						if len(acc.Balances) > 0 {
							gui.OutputWrite(fmt.Sprintf("%18s: %s", "BALANCES", ""))
							for _, balance := range acc.Balances {

								token := toks.GetAnyToken(&balance.Token)
								gui.OutputWrite(fmt.Sprintf("%18s: %s", strconv.FormatUint(config.ConvertToBase(balance.Amount), 10), token.Name))
							}
						} else {
							gui.OutputWrite(fmt.Sprintf("%18s: %s", "BALANCES", "EMPTY"))
						}
						if acc.HasDelegatedStake() {
							gui.OutputWrite(fmt.Sprintf("%18s: %s", "Stake Available", strconv.FormatUint(config.ConvertToBase(acc.DelegatedStake.StakeAvailable), 10)))

							if len(acc.DelegatedStake.StakesPending) > 0 {
								gui.OutputWrite(fmt.Sprintf("%18s: %s", "PENDING STAKES", ""))
								for _, stakePending := range acc.DelegatedStake.StakesPending {
									gui.OutputWrite(fmt.Sprintf("%18s: %10s %t", strconv.FormatUint(stakePending.ActivationHeight, 10), strconv.FormatUint(config.ConvertToBase(stakePending.PendingAmount), 10), stakePending.PendingType))
								}
							} else {
								gui.OutputWrite(fmt.Sprintf("%18s: %s", "PENDING STAKES:", "EMPTY"))
							}
						}
					}

				}

			}

			return
		}); err != nil {
			panic(err)
		}

		return
	}

	cliExportJSONWallet := func(cmd string) {

		str := <-gui.OutputReadString("Path to export")
		f, err := os.Create(str)
		if err != nil {
			panic("File can not be written")
		}
		defer f.Close()

		cliListAddresses("")
		index := <-gui.OutputReadInt("Select Address to be Exported")
		wallet.RLock()
		defer wallet.RUnlock()

		if index < 0 {
			panic("Invalid index")
		}

		var marshal []byte
		var obj interface{}

		if index > len(wallet.Addresses) {
			obj = wallet
		} else {
			obj = wallet.Addresses[index]
		}

		if marshal, err = json.Marshal(obj); err != nil {
			panic("Error marshaling wallet")
		}

		if _, err = fmt.Fprint(f, string(marshal)); err != nil {
			panic("Error writing into file")
		}

		gui.Info("Exported successfully")
	}

	cliCreateNewAddress := func(cmd string) {
		wallet.AddNewAddress()
		cliListAddresses(cmd)
	}

	cliRemoveAddress := func(cmd string) {

		cliListAddresses("")
		index := <-gui.OutputReadInt("Select Address to be Removed")

		wallet.RemoveAddress(index)
		cliListAddresses("")
		gui.OutputWrite("Address removed")
	}

	cliShowMnemonic := func(string) {
		gui.OutputWrite("Mnemonic \n")
		gui.OutputWrite(wallet.Mnemonic)

		gui.OutputWrite("Seed \n")
		gui.OutputWrite(wallet.Seed)
	}

	cliShowPrivateKey := func(cmd string) {

		cliListAddresses("")

		index := <-gui.OutputReadInt("Select Address")

		gui.OutputWrite(wallet.ShowPrivateKey(index))
	}

	gui.CommandDefineCallback("List Addresses", cliListAddresses)
	gui.CommandDefineCallback("Create New Address", cliCreateNewAddress)
	gui.CommandDefineCallback("Show Mnemnonic", cliShowMnemonic)
	gui.CommandDefineCallback("Show Private Key", cliShowPrivateKey)
	gui.CommandDefineCallback("Remove Address", cliRemoveAddress)
	gui.CommandDefineCallback("Export (JSON) Wallet", cliExportJSONWallet)

}
