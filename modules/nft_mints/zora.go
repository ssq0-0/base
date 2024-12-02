package nftmints

import (
	"base/account"
	"base/ethClient"
	"base/utils"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

type Zora struct {
	CA     common.Address
	ABI    *abi.ABI
	Client *ethClient.Client
}

func NewZora(client *ethClient.Client, ca, abiPath string) (*Zora, error) {
	abi, err := utils.ReadAbi(abiPath)
	if err != nil {
		return nil, err
	}

	return &Zora{
		CA:     common.HexToAddress(ca),
		ABI:    abi,
		Client: client,
	}, nil
}

func (z *Zora) Mint(nftCA common.Address, amountIn *big.Int, acc *account.Account) error {
	value := z.calculateMintPrice(amountIn)
	data, err := z.ABI.Pack("buy1155", nftCA, big.NewInt(1), acc.Address, acc.Address, value, big.NewInt(0))
	if err != nil {
		return err
	}

	return z.Client.SendTransaction(acc.PrivateKey, acc.Address, z.CA, z.Client.GetNonce(acc.Address), value, data)
}

func (z *Zora) calculateMintPrice(amountIn *big.Int) *big.Int {
	scaleFactor := new(big.Int).SetInt64(1150000000000000000)
	precision := new(big.Int).SetInt64(1000000000000000000)

	maxEthToSpend := new(big.Int).Mul(amountIn, scaleFactor)
	maxEthToSpend.Div(maxEthToSpend, precision)

	return maxEthToSpend
}
