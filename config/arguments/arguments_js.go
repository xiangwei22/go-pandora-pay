//go:build wasm
// +build wasm

package arguments

var commands = `PANDORA PAY WASM.

Usage:
  pandorapay [--pprof] [--version] [--network=network] [--debug] [--gui-type=type] [--forging] [--new-devnet] [--node-name=name] [--set-genesis=genesis] [--store-wallet-type=type] [--store-chain-type=type] [--node-consensus=type] [--tcp-max-clients=limit] [--node-provide-extended-info-app=bool] [--wallet-encrypt=args] [--wallet-decrypt=password] [--wallet-remove-encryption] [--wallet-export-shared-staked-address=args] [--wallet-import-secret-mnemonic=mnemonic] [--wallet-import-secret-entropy=entropy] [--instance=prefix] [--instance-id=id] [--balance-decryptor-disable-init] [--tcp-connections-ready=threshold] [--exit]
  pandorapay -h | --help
  pandorapay -v | --version

Options:
  -h --help                                          Show this screen.
  --version                                          Show version.
  --debug                                            Debug mode enabled (print log message).
  --instance=prefix                                  Prefix of the instance [default: 0].
  --instance-id=id                                   Number of forked instance (when you open multiple instances). It should be a string number like "1","2","3","4" etc
  --network=network                                  Select network. Accepted values: "mainnet|testnet|devnet". [default: mainnet]
  --new-devnet                                       Create a new devnet genesis.
  --set-genesis=genesis                              Manually set the Genesis via a JSON. Used for devnet genesis in Browser.
  --store-wallet-type=type                           Set Wallet Store Type. Accepted values: "bunt-memory|memory|js". [default: js]
  --store-chain-type=type                            Set Chain Store Type. Accepted values: "bunt-memory|memory|js". [default: memory].
  --forging                                          Start forging blocks.
  --tcp-max-clients=limit                            Change limit of clients [default: 1].
  --tcp-connections-ready=threshold                  Number of connections to become "ready" state [default: 1].
  --node-name=name                                   Change node name.
  --node-consensus=type                              Consensus type. Accepted values: "full|wallet|none". [default: full]
  --node-provide-extended-info-app=bool              Storing and serving additional info to wallet nodes. [default: true]. To enable, it requires full node
  --wallet-import-secret-mnemonic=mnemonic           Import Wallet from a given Mnemonic. It will delete your existing wallet. 
  --wallet-import-secret-entropy=entropy             Import Wallet from a given Entropy. It will delete your existing wallet.
  --wallet-encrypt=args                              Encrypt wallet. Argument must be "password,difficulty".
  --wallet-decrypt=password                          Decrypt wallet.
  --wallet-remove-encryption                         Remove wallet encryption.
  --wallet-export-shared-staked-address=args         Derive and export Staked address. Argument must be "account,nonce,path".
  --balance-decryptor-disable-init                   Disable first balance decryptor initialization. 
  --exit                                             Exit node.
`
