package dex

import (
	"base/account"
	"base/ethClient"
	"base/httpClient"
	"base/models"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

type Odos struct {
	CA               common.Address
	QuoteEndpoint    string
	AssembleEndpoint string
	Client           *ethClient.Client
	HttpClient       *httpClient.HttpClient
}

func NewOdos(client *ethClient.Client, routerCA common.Address, proxy *string) (*Odos, error) {
	return &Odos{
		CA:               routerCA,
		Client:           client,
		HttpClient:       httpClient.NewHttpClient(proxy),
		QuoteEndpoint:    "https://api.odos.xyz/sor/quote/v2",
		AssembleEndpoint: "https://api.odos.xyz/sor/assemble",
	}, nil
}

func (o *Odos) Swap(fromToken, toToken common.Address, amountIn *big.Int, acc *account.Account) error {
	pathId, err := o.quote(fromToken, toToken, amountIn, acc)
	if err != nil {
		return err
	}

	assemblresp, err := o.assemble(pathId, acc.Address)
	if err != nil {
		return err
	}

	txData, err := hex.DecodeString(strings.TrimPrefix(assemblresp.Transaction.Data, "0x"))
	if err != nil {
		return fmt.Errorf("failed to decode txData as hex: %v", err)
	}

	value := new(big.Int)
	if _, ok := value.SetString(assemblresp.Transaction.Value, 10); !ok {
		return err
	}

	return o.Client.SendTransaction(acc.PrivateKey, acc.Address, common.HexToAddress(assemblresp.Transaction.To), o.Client.GetNonce(acc.Address), value, txData)
}

func (o *Odos) quote(fromToken, toToken common.Address, amountIn *big.Int, acc *account.Account) (string, error) {
	params := map[string]interface{}{
		"chainId":              8453,
		"compact":              true,
		"gasPrice":             20,
		"inputTokens":          []map[string]interface{}{{"amount": amountIn.String(), "tokenAddress": fromToken.Hex()}},
		"outputTokens":         []map[string]interface{}{{"proportion": 1, "tokenAddress": toToken.Hex()}},
		"referralCode":         0,
		"slippageLimitPercent": 0.5,
		"sourceBlacklist":      []string{},
		"sourceWhitelist":      []string{},
		"userAddr":             acc.Address.Hex(),
	}

	var quoteResp struct {
		PathId      string  `json:"pathId"`
		PercentDiff float64 `json:"percentDiff"`
	}
	if err := o.HttpClient.SendJSONRequest(o.QuoteEndpoint, "POST", params, &quoteResp); err != nil {
		return "", err
	}

	if quoteResp.PercentDiff > 1 {
		return "", errors.New("low liquidity! Exchange changes the price of the asset by more than 1%")
	}

	return quoteResp.PathId, nil
}

func (o *Odos) assemble(pathID string, userAddr common.Address) (*models.AssembleResponse, error) {
	assembleReq := map[string]interface{}{
		"pathId":   pathID,
		"simulate": false,
		"userAddr": userAddr.Hex(),
	}

	var assembleResp models.AssembleResponse
	if err := o.HttpClient.SendJSONRequest(o.AssembleEndpoint, "POST", assembleReq, &assembleResp); err != nil {
		return nil, fmt.Errorf("ошибка при распарсивании ответа: %v", err)
	}

	return &assembleResp, nil
}
