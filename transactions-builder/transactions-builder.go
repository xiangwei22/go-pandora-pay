package transactions_builder

import (
	"encoding/binary"
	"errors"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/blockchain/tokens/token"
	"pandora-pay/blockchain/transactions/transaction"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	transaction_simple_extra "pandora-pay/blockchain/transactions/transaction/transaction-simple/transaction-simple-extra"
	"pandora-pay/config"
	"pandora-pay/gui"
	"pandora-pay/mempool"
	"pandora-pay/store"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
	"pandora-pay/transactions-builder/wizard"
	"pandora-pay/wallet"
	wallet_address "pandora-pay/wallet/address"
	"sync"
)

type TransactionsBuilder struct {
	wallet  *wallet.Wallet
	mempool *mempool.Mempool
	chain   *blockchain.Blockchain
	lock    *sync.Mutex //TODO replace sync.Mutex with a snyc.Map in order to optimize the transactions creation
}

func (builder *TransactionsBuilder) checkTx(accs *accounts.Accounts, chainHeight uint64, tx *transaction.Transaction) (err error) {

	var available uint64
	for _, vin := range tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple).Vin {

		var acc *account.Account

		acc, err = accs.GetAccount(vin.Bloom.PublicKeyHash, chainHeight)
		if err != nil {
			return
		}

		if acc == nil {
			return errors.New("Account doesn't even exist")
		}

		available = acc.GetAvailableBalance(vin.Token)

		if available, err = builder.mempool.GetBalance(vin.Bloom.PublicKeyHash, available, vin.Token); err != nil {
			return
		}
		if available < vin.Amount {
			return errors.New("You don't have enough coins")
		}
	}
	return
}

func (builder *TransactionsBuilder) CreateSimpleTx_Float(propagateTx bool, from []string, nonce uint64, amounts []float64, amountsTokens [][]byte, dsts []string, dstsAmounts []float64, dstsTokens [][]byte, feePerByte int, feeToken []byte) (*transaction.Transaction, error) {

	amountsFinal := make([]uint64, len(from))
	dstsAmountsFinal := make([]uint64, len(dsts))

	if err := store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
		toks := tokens.NewTokens(reader)
		for i := range from {

			var tok *token.Token
			if tok, err = toks.GetToken(amountsTokens[i]); err != nil {
				return
			}

			if tok == nil {
				return errors.New("Token was not found")
			}
			if amountsFinal[i], err = tok.ConvertToUnits(amounts[i]); err != nil {
				return
			}
		}
		for i := range dstsTokens {
			var tok *token.Token
			if tok, err = toks.GetToken(dstsTokens[i]); err != nil {
				return
			}

			if tok == nil {
				return errors.New("Token was not found")
			}
			if dstsAmountsFinal[i], err = tok.ConvertToUnits(dstsAmounts[i]); err != nil {
				return
			}
		}

		return
	}); err != nil {
		return nil, err
	}

	return builder.CreateSimpleTx(propagateTx, from, nonce, amountsFinal, amountsTokens, dsts, dstsAmountsFinal, dstsTokens, feePerByte, feeToken)
}

func (builder *TransactionsBuilder) CreateSimpleTx(propagateTx bool, from []string, nonce uint64, amounts []uint64, amountsTokens [][]byte, dsts []string, dstsAmounts []uint64, dstsTokens [][]byte, feePerByte int, feeToken []byte) (*transaction.Transaction, error) {

	builder.lock.Lock()
	defer builder.lock.Unlock()

	var tx *transaction.Transaction
	if err := store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		accs := accounts.NewAccounts(reader)

		chainHeight, _ := binary.Uvarint(reader.Get("chainHeight"))

		keys := make([][]byte, len(from))

		for i, fromAddress := range from {

			var fromWalletAddress *wallet_address.WalletAddress
			if fromWalletAddress, err = builder.wallet.GetWalletAddressByEncodedAddress(fromAddress); err != nil {
				return
			}

			var acc *account.Account
			if acc, err = accs.GetAccount(fromWalletAddress.PublicKeyHash, chainHeight); err != nil {
				return
			}

			if acc == nil {
				return errors.New("Account doesn't exist")
			}

			available := acc.GetAvailableBalance(amountsTokens[i])
			if available < amounts[i] {
				return errors.New("Not enough funds")
			}

			if i == 0 && nonce == 0 {
				nonce = builder.mempool.GetNonce(fromWalletAddress.PublicKeyHash, acc.Nonce)
			}
			keys[i] = fromWalletAddress.PrivateKey.Key
		}

		if tx, err = wizard.CreateSimpleTx(nonce, keys, amounts, amountsTokens, dsts, dstsAmounts, dstsTokens, feePerByte, feeToken); err != nil {
			gui.GUI.Error("Error creating Tx: ", tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple).Nonce, err)
			return
		}

		if err = builder.checkTx(accs, chainHeight, tx); err != nil {
			return
		}

		return
	}); err != nil {
		return nil, err
	}

	if propagateTx {
		if err := builder.mempool.AddTxToMemPool(tx, builder.chain.GetChainData().Height, true, true); err != nil {
			return nil, err
		}
	}

	return tx, nil

}

func (builder *TransactionsBuilder) CreateUnstakeTx_Float(propagateTx bool, from string, nonce uint64, unstakeAmount float64, feePerByte int, feeToken []byte, payFeeInExtra bool) (*transaction.Transaction, error) {

	unstakeAmountFinal, err := config.ConvertToUnits(unstakeAmount)
	if err != nil {
		return nil, err
	}

	return builder.CreateUnstakeTx(propagateTx, from, nonce, unstakeAmountFinal, feePerByte, feeToken, payFeeInExtra)
}

func (builder *TransactionsBuilder) CreateUnstakeTx(propagateTx bool, from string, nonce uint64, unstakeAmount uint64, feePerByte int, feeToken []byte, payFeeInExtra bool) (*transaction.Transaction, error) {

	builder.lock.Lock()
	defer builder.lock.Unlock()

	fromWalletAddress, err := builder.wallet.GetWalletAddressByEncodedAddress(from)
	if err != nil {
		return nil, err
	}

	var tx *transaction.Transaction
	if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		chainHeight, _ := binary.Uvarint(reader.Get("chainHeight"))

		accs := accounts.NewAccounts(reader)
		var account *account.Account
		if account, err = accs.GetAccount(fromWalletAddress.PublicKeyHash, chainHeight); err != nil {
			return
		}

		if account == nil {
			return errors.New("Account doesn't exist")
		}

		availableUnstake := account.GetDelegatedStakeAvailable()

		if availableUnstake < unstakeAmount {
			return errors.New("You don't have enough staked coins")
		}

		if nonce == 0 {
			nonce = builder.mempool.GetNonce(fromWalletAddress.PublicKeyHash, account.Nonce)
		}

		if tx, err = wizard.CreateUnstakeTx(nonce, fromWalletAddress.PrivateKey.Key, unstakeAmount, feePerByte, feeToken, payFeeInExtra); err != nil {
			return
		}

		availableDelegatedStake := account.GetDelegatedStakeAvailable()

		if availableDelegatedStake < tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple).Vin[0].Amount+tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple).TransactionSimpleExtraInterface.(*transaction_simple_extra.TransactionSimpleUnstake).FeeExtra {
			return errors.New("You don't have enough staked coins to pay for the fee")
		}

		return
	}); err != nil {
		return nil, err
	}

	if propagateTx {
		if err = builder.mempool.AddTxToMemPool(tx, builder.chain.GetChainData().Height, true, true); err != nil {
			return nil, err
		}
	}

	return tx, nil
}

func (builder *TransactionsBuilder) CreateDelegateTx_Float(propagateTx bool, from string, nonce uint64, delegateAmount float64, delegateNewPubKeyHashGenerate bool, delegateNewPubKeyHash []byte, feePerByte int, feeToken []byte) (*transaction.Transaction, error) {

	delegateAmountFinal, err := config.ConvertToUnits(delegateAmount)
	if err != nil {
		return nil, err
	}

	return builder.CreateDelegateTx(propagateTx, from, nonce, delegateAmountFinal, delegateNewPubKeyHashGenerate, delegateNewPubKeyHash, feePerByte, feeToken)
}

func (builder *TransactionsBuilder) CreateDelegateTx(propagateTx bool, from string, nonce uint64, delegateAmount uint64, delegateNewPubKeyHashGenerate bool, delegateNewPubKeyHash []byte, feePerByte int, feeToken []byte) (*transaction.Transaction, error) {

	builder.lock.Lock()
	defer builder.lock.Unlock()

	fromWalletAddress, err := builder.wallet.GetWalletAddressByEncodedAddress(from)
	if err != nil {
		return nil, err
	}

	var tx *transaction.Transaction
	if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		chainHeight, _ := binary.Uvarint(reader.Get("chainHeight"))

		accs := accounts.NewAccounts(reader)
		var acc *account.Account
		if acc, err = accs.GetAccount(fromWalletAddress.PublicKeyHash, chainHeight); err != nil {
			return
		}

		if acc == nil {
			return errors.New("Account doesn't exist")
		}

		available := acc.GetAvailableBalance(config.NATIVE_TOKEN)

		if available < delegateAmount {
			return errors.New("You don't have enough coins to delegate")
		}

		if nonce == 0 {
			nonce = builder.mempool.GetNonce(fromWalletAddress.PublicKeyHash, acc.Nonce)
		}

		if delegateNewPubKeyHashGenerate {

			var delegatedStake *wallet_address.WalletAddressDelegatedStake
			if delegatedStake, err = fromWalletAddress.DeriveDelegatedStake(uint32(nonce)); err != nil {
				return
			}
			delegateNewPubKeyHash = delegatedStake.PublicKeyHash

		}

		if tx, err = wizard.CreateDelegateTx(nonce, fromWalletAddress.PrivateKey.Key, delegateAmount, delegateNewPubKeyHash, feePerByte, feeToken); err != nil {
			return
		}

		if err = builder.checkTx(accs, chainHeight, tx); err != nil {
			return
		}

		return
	}); err != nil {
		return nil, err
	}

	if propagateTx {
		if err = builder.mempool.AddTxToMemPool(tx, builder.chain.GetChainData().Height, true, true); err != nil {
			return nil, err
		}
	}

	return tx, nil
}

func TransactionsBuilderInit(wallet *wallet.Wallet, mempool *mempool.Mempool, chain *blockchain.Blockchain) (builder *TransactionsBuilder) {

	builder = &TransactionsBuilder{
		wallet:  wallet,
		chain:   chain,
		mempool: mempool,
		lock:    &sync.Mutex{},
	}

	builder.initCLI()

	return
}
