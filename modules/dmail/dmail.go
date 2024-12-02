package dmail

import (
	"base/account"
	"base/ethClient"
	"base/utils"
	"crypto/sha256"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

type Dmail struct {
	ABI    *abi.ABI
	Client *ethClient.Client
	CA     common.Address
}

func NewDmail(client *ethClient.Client, ca, abiPath string) (*Dmail, error) {
	abi, err := utils.ReadAbi(abiPath)
	if err != nil {
		return nil, err
	}
	return &Dmail{
		ABI:    abi,
		Client: client,
		CA:     common.HexToAddress(ca),
	}, nil
}

func (d *Dmail) SendMail(acc *account.Account) error {
	emailHash, err := d.GenerateRandomSHA256()
	if err != nil {
		return fmt.Errorf("failed to generate email hash: %w", err)
	}

	themeHash, err := d.GenerateRandomSHA256()
	if err != nil {
		return fmt.Errorf("failed to generate theme hash: %w", err)
	}

	data, err := d.ABI.Pack("send_mail", emailHash, themeHash)
	if err != nil {
		return err
	}

	return d.Client.SendTransaction(acc.PrivateKey, acc.Address, d.CA, d.Client.GetNonce(acc.Address), big.NewInt(0), data)
}

func (d *Dmail) GenerateRandomSHA256() (string, error) {
	length, err := generateRandomInt(1, 16)
	if err != nil {
		return "", fmt.Errorf("failed to generate random length: %w", err)
	}

	randomStr, err := generateRandomString(length)
	if err != nil {
		return "", fmt.Errorf("failed to generate random string: %w", err)
	}

	hashBytes := sha256.Sum256([]byte(randomStr))
	return fmt.Sprintf("%x", hashBytes[:]), nil
}
