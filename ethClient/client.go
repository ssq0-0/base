package ethClient

import (
	"github.com/ethereum/go-ethereum/ethclient"
)

type Client struct {
	Client *ethclient.Client
}

func NewClient(rpc string) *Client {
	client, err := ethclient.Dial(rpc)
	if err != nil {
		return nil
	}

	return &Client{
		Client: client,
	}
}

func CloseAllClients(clients map[string]*Client) {
	for _, client := range clients {
		if client.Client != nil {
			client.Client.Close()
		}
	}
}
