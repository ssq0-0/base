package dex

import (
	"base/account"
	"base/config"
	"base/ethClient"
	"base/httpClient"
	"base/models"
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"net/url"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

type OpenOcean struct {
	SwapQuoteEndpoint string
	CA                common.Address
	Client            *ethClient.Client
	HttpClient        *httpClient.HttpClient
}

func NewOpenOcean(client *ethClient.Client, ca common.Address, proxy *string) (*OpenOcean, error) {
	httpcl := httpClient.NewHttpClient(proxy)
	return &OpenOcean{
		CA:                ca,
		SwapQuoteEndpoint: "https://open-api.openocean.finance/v3/8453/swap_quote",
		Client:            client,
		HttpClient:        httpcl,
	}, nil
}

func (o *OpenOcean) Swap(fromToken, toToken common.Address, amount *big.Int, acc *account.Account) error {
	swapData, err := o.swapQuote(fromToken, toToken, amount, acc)
	if err != nil {
		return err
	}

	value := new(big.Int)
	if _, ok := value.SetString(swapData.Data.Value, 10); !ok {
		return fmt.Errorf("invalid value for big.Int: %s", swapData.Data.Value)
	}

	dataStr := strings.TrimPrefix(swapData.Data.Data, "0x")
	txData, err := hex.DecodeString(dataStr)
	if err != nil {
		return fmt.Errorf("failed to decode txData as hex: %v", err)
	}

	if len(txData) == 0 {
		txData, err = base64.StdEncoding.DecodeString(dataStr)
		if err != nil {
			return fmt.Errorf("failed to decode txData as Base64: %v", err)
		}
	}

	return o.Client.SendTransaction(acc.PrivateKey, acc.Address, common.HexToAddress(swapData.Data.To), o.Client.GetNonce(acc.Address), value, txData)
}

func (o *OpenOcean) swapQuote(fromToken, toToken common.Address, amount *big.Int, acc *account.Account) (*models.SwapQuoteResponse, error) {
	var quote models.SwapQuoteResponse
	if err := o.HttpClient.SendGetRequest(o.setParams(fromToken, toToken, amount, acc, o.getGasForOP()), &quote); err != nil {
		return nil, err
	}

	return &quote, nil
}

func (o *OpenOcean) setParams(fromToken, toToken common.Address, amount *big.Int, acc *account.Account, gasPrice string) string {
	params := url.Values{}
	params.Set("inTokenAddress", fromToken.Hex())
	params.Set("outTokenAddress", toToken.Hex())
	params.Set("amount", o.amountConverter(fromToken, amount))
	params.Set("gasPrice", gasPrice)
	params.Set("slippage", "1")
	params.Set("account", acc.Address.Hex())

	return fmt.Sprintf("%s?%s", o.SwapQuoteEndpoint, params.Encode())
}

func (o *OpenOcean) amountConverter(token common.Address, amount *big.Int) string {
	var decimals float64 = 1e6
	if token == config.WETH || token == config.WooFiETH {
		decimals = 1e18
	}

	ethFloat := new(big.Float).Quo(new(big.Float).SetInt(amount), big.NewFloat(decimals))
	floatVal, _ := ethFloat.Float64()
	return strconv.FormatFloat(floatVal, 'f', 18, 64)
}

func (o *OpenOcean) getGasForOP() string {
	gasPrice, err := o.Client.Client.SuggestGasPrice(context.Background())
	if err != nil {
		return ""
	}

	gwei := new(big.Float).Quo(new(big.Float).SetInt(gasPrice), big.NewFloat(1e9))
	floatVal, _ := gwei.Float64()
	return strconv.FormatFloat(floatVal, 'f', 18, 64)
}
