package config

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

var (
	RPCs = map[string]string{
		"eth":       "https://eth.drpc.org",
		"base":      "https://mainnet.base.org",
		"arbitrum":  "https://arbitrum.drpc.org",
		"optimism":  "https://rpc.ankr.com/optimism",
		"polygon":   "https://polygon.drpc.org",
		"avalanche": "https://avalanche.drpc.org",
	}
)

var (
	Slippage   = big.NewFloat(0.98) // 2% проскальзывания
	MinBalance = big.NewInt(1e15)
)

const MaxUint256Str = "115792089237316195423570985008687907853269984665640564039457584007913129639935"

var (
	MaxUint256          = new(big.Int)
	MinBalanceInDollars = big.NewFloat(1.0)
	Erc20ABI            *abi.ABI
)

var (
	DEFAULT_actionNumMin  = 10
	DEFAULT_actionNumMax  = 20
	DEFAULT_actionTimeMin = 10
	DEFAULT_actionTimeMax = 30
)

var (
	BSN_price_3length  = big.NewFloat(0.1)
	BSN_price_4length  = big.NewFloat(0.01)
	BSN_price_5length  = big.NewFloat(0.001)
	BSN_price_10length = big.NewFloat(0.0001)
)

var (
	USDC         = common.HexToAddress("0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913")
	USDbC        = common.HexToAddress("0xd9aAEc86B65D86f6A7B5B1b0c42FFA531710b6CA")
	WETH         = common.HexToAddress("0x4200000000000000000000000000000000000006")
	AaveWETH     = common.HexToAddress("0xD4a0e0b9149BCee3C920d2E00b5dE09138fd8bb7")
	AaveUSDC     = common.HexToAddress("0x4e65fe4dba92790696d040ac24aa414708f5c0ab")
	MoonwellWETH = common.HexToAddress("0x628ff693426583D9a7FB391E54366292F509D457")
	WooFiETH     = common.HexToAddress("0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE")
)

var (
	AviableTokens = []common.Address{WETH, USDC, USDbC}
)

var TokenDecimals = map[common.Address]uint8{
	WETH:         18,
	WooFiETH:     18,
	AaveWETH:     18,
	MoonwellWETH: 18,

	AaveUSDC: 6,
	USDC:     6,
	USDbC:    6,
}

var TokenPrice = map[common.Address]*big.Float{
	WETH:     big.NewFloat(3500.0),
	AaveUSDC: big.NewFloat(1.0),
	USDC:     big.NewFloat(1.0),
	USDbC:    big.NewFloat(1.0),
}

var (
	PROTOCOLS_CAs = map[common.Address][]common.Address{
		USDC: {
			common.HexToAddress("0x678Aa4bF4E210cf2166753e054d5b7c31cc7fa86"),
			common.HexToAddress("0x2626664c2603336E57B271c5C0b26F421741e481"),
			common.HexToAddress("0x4c4AF8DBc524681930a27b2F1Af5bcC8062E6fB7"),
			common.HexToAddress("0xA238Dd80C259a72e81d7e4664a9801593F98d1c5"),
		},
		USDbC: {
			common.HexToAddress("0x678Aa4bF4E210cf2166753e054d5b7c31cc7fa86"),
			common.HexToAddress("0x2626664c2603336E57B271c5C0b26F421741e481"),
		},
		AaveWETH: {
			common.HexToAddress("0x729b3EA8C005AbC58c9150fb57Ec161296F06766"),
		},
		AaveUSDC: {
			common.HexToAddress("0xA238Dd80C259a72e81d7e4664a9801593F98d1c5"),
		},
	}
)

var (
	OtherTokens = map[string]common.Address{
		"ethereum_usdc": common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"),
		"ethereum_dai":  common.HexToAddress("0x6B175474E89094C44Da98b954EedeAC495271d0F"),

		"bsc_usdt": common.HexToAddress("0x55d398326f99059fF775485246999027B3197955"),

		"avalanche_usdc": common.HexToAddress("0xB97EF9Ef8734C71904D8002F8b6Bc66Dd9c48a6E"),
		"avalanche_usdt": common.HexToAddress("0x9702230A8Ea53601f5cD2dc00fDBc13d4dF4A8c7"),

		"polygon_usdce": common.HexToAddress("0x2791bca1f2de4661ed88a30c99a7a9449aa84174"),
		"polygon_usdt":  common.HexToAddress("0xc2132d05d31c914a87c6611c10748aeb04b58e8f"),

		"arbitrum_usdc": common.HexToAddress("0xff970a61a04b1ca14834a43f5de4533ebddb5cc8"),
		"arbitrum_usdt": common.HexToAddress("0xfd086bc7cd5c481dcc9c85ebe478a1c0b69fcbb9"),
		"arbitrum_eth":  common.HexToAddress("0x82aF49447D8a07e3bd95BD0d56f35241523fBab1"),

		"optimism_eth":  common.HexToAddress("0x4200000000000000000000000000000000000006"),
		"optimism_usdc": common.HexToAddress("0x7F5c764cBc14f9669B88837ca1490cCa17c31607"),
	}
)

var (
	LZ_Main_CA = map[string]map[string]common.Address{
		"ethereum": {
			"swap_ca": common.HexToAddress("0x8731d54E9D02c286767d56ac03e8037C07e01e98"),
			"fee_ca":  common.HexToAddress("0x296F55F8Fb28E498B858d0BcDA06D955B2Cb3f97"),
		},
		"bsc": {
			"swap_ca": common.HexToAddress("0x4a364f8c717cAAD9A442737Eb7b8A55cc6cf18D8"),
			"fee_ca":  common.HexToAddress("0x6694340fc020c5E6B96567843da2df01b2CE1eb6"),
		},
		"avalanche": {
			"swap_ca": common.HexToAddress("0x45A01E4e04F14f7A4a6702c74187c5F6222033cd"),
			"fee_ca":  common.HexToAddress("0x9d1B1669c73b033DFe47ae5a0164Ab96df25B944"),
		},
		"polygon": {
			"swap_ca": common.HexToAddress("0x45A01E4e04F14f7A4a6702c74187c5F6222033cd"),
			"fee_ca":  common.HexToAddress("0x9d1B1669c73b033DFe47ae5a0164Ab96df25B944"),
		},
		"arbitrum": {
			"swap_ca": common.HexToAddress("0x53Bf833A5d6c4ddA888F69c22C88C9f356a41614"),
			"fee_ca":  common.HexToAddress("0x352d8275AAE3e0c2404d9f68f6cEE084B5bEB3DD"),
		},
		"optimism": {
			"swap_ca": common.HexToAddress("0xB0D502E938ed5f4df2E681fE6E419ff29631d62b"),
			"fee_ca":  common.HexToAddress("0x701a95707A0290AC8B90b3719e8EE5b210360883"),
		},
	}
)

var (
	Erc20JSON = []byte(`[
	{
		"constant":true,
		"inputs":[{"name":"account","type":"address"}],
		"name":"balanceOf",
		"outputs":[{"name":"","type":"uint256"}],
		"payable":false,
		"stateMutability":"view",
		"type":"function"
	},
	{
		"constant":true,
		"inputs":[{"name":"spender","type":"address"},{"name":"owner","type":"address"}],
		"name":"allowance",
		"outputs":[{"name":"","type":"uint256"}],
		"payable":false,
		"stateMutability":"view",
		"type":"function"
	},
	{
		"constant":false,
		"inputs":[{"name":"spender","type":"address"},{"name":"amount","type":"uint256"}],
		"name":"approve",
		"outputs":[{"name":"","type":"bool"}],
		"payable":false,
		"stateMutability":"nonpayable",
		"type":"function"
	},
	{
		"constant":false,
		"inputs":[{"name":"recipient","type":"address"},{"name":"amount","type":"uint256"}],
		"name":"transfer",
		"outputs":[{"name":"","type":"bool"}],
		"payable":false,
		"stateMutability":"nonpayable",
		"type":"function"
	},
	{
		"constant":false,
		"inputs":[{"name":"sender","type":"address"},{"name":"recipient","type":"address"},{"name":"amount","type":"uint256"}],
		"name":"transferFrom",
		"outputs":[{"name":"","type":"bool"}],
		"payable":false,
		"stateMutability":"nonpayable",
		"type":"function"
	},
	{
		"constant":true,
		"inputs":[],
		"name":"decimals",
		"outputs":[{"name":"","type":"uint8"}],
		"payable":false,
		"stateMutability":"view",
		"type":"function"
	},
	{
		"constant":true,
		"inputs":[],
		"name":"name",
		"outputs":[{"name":"","type":"string"}],
		"payable":false,
		"stateMutability":"view",
		"type":"function"
	},
	{
		"constant":true,
		"inputs":[],
		"name":"symbol",
		"outputs":[{"name":"","type":"string"}],
		"payable":false,
		"stateMutability":"view",
		"type":"function"
	},
	{
		"constant":true,
		"inputs":[],
		"name":"totalSupply",
		"outputs":[{"name":"","type":"uint256"}],
		"payable":false,
		"stateMutability":"view",
		"type":"function"
	}
]`)
)

const (
	Logo = `                                                                                              
::::::................:::::....:.................................:.....................................::::............:::....:.........::::.............:::::::::
:::::::............+@@@@@@@@%=@%:*@@@@@@-...*@@@@@@-:#@@@@@@@@@@@@=.:%@@@@@*.=@@@@@@@@@@@@=.........+@@@@@@@@#@*....%@@@@@@@#@%......#@@@@@@@@+..........:::::::::
::::::::.........*@@@=.....-@@@:..:@@@#......=@@@#....+@@@+.....*@+...%@@@-...-@@@%.....%@=.......:%@@=....-@@@:..+@@%.....*@@=....%@@#.....*@@@+........:::::::::
:::::::::.......#@@@-.......-@*...:@@@#......=@@@#....+@@@+......*+...%@@@:...-@@@#......#=.......+@@@*.....:@=...%@@@:.....##...:%@@#.......*@@@*.........::::.::
:::::::::......=@@@#..............:@@@#......=@@@%....+@@@+...-@=.....%@@@:...-@@@#...:%=.........=@@@@@%*:.......#@@@@@#-.......#@@@=.......:@@@@-..............:
:::::::::......#@@@#..............-@@@@@@@@@@@@@@%....+@@@@@@@@@-.....%@@@:...-@@@@@%%@@-..........=@@@@@@@@#=....:#@@@@@@@%+:..:@@@@=........%@@@#..............:
::::::::.......#@@@#..............-@@@#::::::=@@@#....+@@@+...#@-.....%@@@:...-@@@@-:-%@:............:*%@@@@@@@-.....=#@@@@@@@*..@@@@+........%@@@+......:::....::
:::::::::......-@@@@:.............-@@@#......=@@@#....+@@@*...:+:..:..%@@@:...-@@@@...:%:.........:*.....:*@@@@%..++.....=%@@@@-.+@@@*...:-:..%@@%....:::::::...::
::::::::...:-...+@@@@:........#%:.:@@@#......=@@@#....+@@@*.......#@:.%@@@:...=@@@%...............%@*......-@@@#..@@-......*@@@:..#@@@:.#@@@#=@@@=....::::::::::::
:::::::....-@-...:@@@@+....:%@%:..-@@@#......=@@@%....+@@@#.....+@@=..%@@@=...=@@@@..............=@@@@:....=@@@:.+@@@+.....#@@=....=@@@*:-.#@@@%:....:-:::::::::.:
:::::::....=@:%:...:#@@@@@@@#:..-@@@@@@@#..:%@@@@@@#-@@@@@@@@@@@@@#.+@@@@@@@=%@@@@@@@=...........#@::%@@@@@@*:..:%*:+@@@@@@@=........=%@@@@@@@@@%+:=%@=.::::::::::
:::::::....=@:%-:............................................................................................................................-%@@@@@+:..::::::..::
`
	Subscribe = `Subscribe <3 		   https://t.me/cheifssq 			<3`
	DonateSOL = `SOL      <3 6xJrAzhGFJ58snkgeVsPpALMkppCHaoYc841REpT5Py <3`
	DonateEVM = `EVM      <3 0x899e6Bf266754Df7C1E589367aFAb118fd15735C <3`
	DonateBTC = `BTC      <3 0x899e6Bf266754Df7C1E589367aFAb118fd15735C <3`
)
