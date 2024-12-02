package ethClient

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
)

func (c *Client) GetGasValues(msg ethereum.CallMsg) (uint64, *big.Int, *big.Int, error) {
	header, err := c.Client.HeaderByNumber(context.Background(), nil)
	if err != nil {
		return 0, nil, nil, err
	}
	baseFee := header.BaseFee

	maxPriorityFeePerGas := big.NewInt(1e7)
	maxFeePerGas := new(big.Int).Add(baseFee, maxPriorityFeePerGas)

	gasLimit, err := c.Client.EstimateGas(context.Background(), msg)
	if err != nil {
		return 0, nil, nil, err
	}

	return gasLimit, maxPriorityFeePerGas, maxFeePerGas, nil
}
