package nftmints

import (
	"base/account"
	"base/ethClient"
	"base/utils"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

type Nft2Me struct {
	ABI    *abi.ABI
	Client *ethClient.Client
}

func NewNft2Me(client *ethClient.Client, abiPath string) (*Nft2Me, error) {
	abi, err := utils.ReadAbi(abiPath)
	if err != nil {
		return nil, err
	}

	return &Nft2Me{
		ABI:    abi,
		Client: client,
	}, nil
}

func (nft *Nft2Me) Mint(mintCA common.Address, amount, price *big.Int, acc *account.Account) error {
	data, err := nft.ABI.Pack("mint", amount)
	if err != nil {
		return fmt.Errorf("failed pack data for mint nft2me: %v", err)
	}

	return nft.Client.SendTransaction(acc.PrivateKey, acc.Address, mintCA, nft.Client.GetNonce(acc.Address), price, data)
}
