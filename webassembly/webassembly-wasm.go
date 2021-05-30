// +build wasm

package webassembly

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"pandora-pay/blockchain/transactions/transaction/transaction-simple"
	"pandora-pay/blockchain/transactions/transaction/transaction-type"
	"pandora-pay/config"
	"pandora-pay/config/globals"
	"pandora-pay/gui"
	"pandora-pay/helpers/events"
	"pandora-pay/helpers/identicon"
	"pandora-pay/wallet"
	"pandora-pay/wallet/address"
	"sync/atomic"
	"syscall/js"
)

var subscriptionsIndex uint64
var startMainCallback func()

func convertJSON(obj interface{}) interface{} {
	str, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}

	return string(str)
}

func subscribeEvents(none js.Value, args []js.Value) interface{} {

	if len(args) == 0 || args[0].Type() != js.TypeFunction {
		return errors.New("Argument must be a callback")
	}

	index := atomic.AddUint64(&subscriptionsIndex, 1)
	channel := globals.MainEvents.AddListener()
	callback := args[0]

	go func() {
		for {
			dataValue := <-channel
			data := dataValue.(*events.EventData)

			var final interface{}

			switch v := data.Data.(type) {
			case string:
				final = data.Data
			case interface{}:
				final = convertJSON(v)
			default:
				final = data.Data
			}

			callback.Invoke(data.Name, final)
		}
	}()

	return index
}

func helloPandora(js.Value, []js.Value) interface{} {
	gui.GUI.Info("HelloPandora works!")
	return true
}

func start(js.Value, []js.Value) interface{} {
	startMainCallback()
	return true
}

func getWallet(js.Value, []js.Value) interface{} {
	return convertJSON(globals.Data["wallet"].(*wallet.Wallet))
}

func getIdenticon(a js.Value, b []js.Value) interface{} {
	out, err := identicon.GenerateToBytes([]byte(b[0].String()), b[1].Int(), b[2].Int())
	if err != nil {
		panic(err)
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(out)
}

func Initialize(startMainCb func()) {

	startMainCallback = startMainCb

	PandoraPayExport := map[string]interface{}{
		"helpers": js.ValueOf(map[string]interface{}{
			"helloPandora": js.FuncOf(helloPandora),
			"start":        js.FuncOf(start),
			"getIdenticon": js.FuncOf(getIdenticon),
		}),
		"events": js.ValueOf(map[string]interface{}{
			"subscribe": js.FuncOf(subscribeEvents),
		}),
		"wallet": js.ValueOf(map[string]interface{}{
			"getWallet": js.FuncOf(getWallet),
		}),
		"enums": js.ValueOf(map[string]interface{}{
			"transactions": js.ValueOf(map[string]interface{}{
				"transactionType": js.ValueOf(map[string]interface{}{
					"TX_SIMPLE": js.ValueOf(uint64(transaction_type.TX_SIMPLE)),
				}),
				"transactionSimple": js.ValueOf(map[string]interface{}{
					"scriptType": js.ValueOf(map[string]interface{}{
						"SCRIPT_NORMAL":   js.ValueOf(uint64(transaction_simple.SCRIPT_NORMAL)),
						"SCRIPT_UNSTAKE":  js.ValueOf(uint64(transaction_simple.SCRIPT_UNSTAKE)),
						"SCRIPT_WITHDRAW": js.ValueOf(uint64(transaction_simple.SCRIPT_WITHDRAW)),
						"SCRIPT_DELEGATE": js.ValueOf(uint64(transaction_simple.SCRIPT_DELEGATE)),
					}),
				}),
			}),
			"wallet": js.ValueOf(map[string]interface{}{
				"version": js.ValueOf(map[string]interface{}{
					"VERSION_SIMPLE": js.ValueOf(int(wallet.VERSION_SIMPLE)),
				}),
				"encryptedVersion": js.ValueOf(map[string]interface{}{
					"ENCRYPTED_VERSION_PLAIN_TEXT": js.ValueOf(int(wallet.ENCRYPTED_VERSION_PLAIN_TEXT)),
					"ENCRYPTED_VERSION_ENCRYPTION": js.ValueOf(int(wallet.ENCRYPTED_VERSION_ENCRYPTION)),
				}),
				"address": js.ValueOf(map[string]interface{}{
					"version": js.ValueOf(map[string]interface{}{
						"versionTransparent": js.ValueOf(int(wallet_address.VERSION_TRANSPARENT)),
					}),
				}),
			}),
		}),
		"config": js.ValueOf(map[string]interface{}{
			"NAME":                    js.ValueOf(config.NAME),
			"NETWORK_SELECTED":        js.ValueOf(config.NETWORK_SELECTED),
			"NETWORK_SELECTED_NAME":   js.ValueOf(config.GetNetworkName()),
			"NETWORK_SELECTED_PREFIX": js.ValueOf(config.GetNetworkPrefix()),
		}),
	}

	js.Global().Set("PandoraPay", js.ValueOf(PandoraPayExport))

}
