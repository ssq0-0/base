package domains

import (
	"base/account"
	"base/ethClient"
	"base/models"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type BSN struct {
	RegisterABI *abi.ABI
	ResolverABI *abi.ABI
	RegisterCA  common.Address
	ResolverCA  common.Address
	Client      *ethClient.Client
}

func NewBSN(client *ethClient.Client, registerCA, resolverCA common.Address, registerABI, resolverABI *abi.ABI) (*BSN, error) {
	return &BSN{
		RegisterABI: registerABI,
		ResolverABI: resolverABI,
		RegisterCA:  registerCA,
		ResolverCA:  resolverCA,
		Client:      client,
	}, nil
}

func (bsn *BSN) RegisterName(name string, price *big.Int, acc *account.Account) error {
	node := namehash(fmt.Sprintf("%s%s", name, ".base.eth"))

	data, err := bsn.packResolverData(node, name, acc.Address, "An example description")
	if err != nil {
		return err
	}

	packedData, err := bsn.RegisterABI.Pack("register", models.DSNDomain{
		Name:          name,
		Owner:         acc.Address,
		Duration:      big.NewInt(31557600),
		Resolver:      bsn.ResolverCA,
		Data:          data,
		ReverseRecord: true,
	})
	if err != nil {
		return err
	}

	return bsn.Client.SendTransaction(acc.PrivateKey, acc.Address, bsn.RegisterCA, bsn.Client.GetNonce(acc.Address), price, packedData)
}

func (bsn *BSN) packResolverData(node common.Hash, name string, addr common.Address, description string) ([][]byte, error) {
	dataSetAddr, err := bsn.ResolverABI.Pack("setAddr", node, addr)
	if err != nil {
		return nil, fmt.Errorf("failed to pack setAddr data: %w", err)
	}

	dataSetName, err := bsn.ResolverABI.Pack("setName", node, name)
	if err != nil {
		return nil, fmt.Errorf("failed to pack setName data: %w", err)
	}

	dataSetText, err := bsn.ResolverABI.Pack("setText", node, "description", description)
	if err != nil {
		return nil, fmt.Errorf("failed to pack setText data: %w", err)
	}

	return [][]byte{dataSetAddr, dataSetName, dataSetText}, nil
}

func namehash(name string) common.Hash {
	node := make([]byte, 32)
	if name != "" {
		labels := strings.Split(name, ".")
		for i := len(labels) - 1; i >= 0; i-- {
			label := labels[i]
			labelHash := crypto.Keccak256Hash([]byte(label))
			node = crypto.Keccak256Hash(append(node, labelHash.Bytes()...)).Bytes()
		}
	}
	return common.BytesToHash(node)
}
