package ethClient

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
)

func (c *Client) GetNonce(address common.Address) uint64 {
	nonce, err := c.Client.PendingNonceAt(context.Background(), address)
	if err != nil {
		return 0
	}
	return nonce
}
