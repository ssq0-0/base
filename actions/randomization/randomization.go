package randomization

import (
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"sync"

	"base/account"
	"base/actions"
	"base/actions/types"
	"base/config"
	"base/ethClient"
	"base/logger"
	"base/models"

	"github.com/ethereum/go-ethereum/common"
)

type Randomizer struct {
	availableTokens []common.Address
	availableNFTs   map[string]map[common.Address]*big.Int
	Clients         map[string]*ethClient.Client
	tokenMutex      sync.Mutex
	nftMutex        sync.Mutex
}

func NewRandomizer(availableTokens []common.Address, availableNFTs map[string]map[common.Address]*big.Int, clients map[string]*ethClient.Client) *Randomizer {
	return &Randomizer{
		availableTokens: availableTokens,
		availableNFTs:   availableNFTs,
		Clients:         clients,
	}
}

func (r *Randomizer) GenerateActionSequence(modules *account.ModulesConfig, walletConfig *account.WalletConfig, acc *account.Account) ([]actions.Action, error) {
	numActions, err := getNumActions(walletConfig)
	if err != nil {
		numActions = 10
	}

	availableActionTypes := getAvailableActions(modules, walletConfig)
	if len(availableActionTypes) == 0 {
		return nil, errors.New("no action types available for generation")
	}

	if len(availableActionTypes) == 1 && availableActionTypes[0] == types.CollectorModAction {
		action, err := r.GenerateSingleAction(availableActionTypes[0], acc)
		if err != nil {
			return nil, err
		}
		return []actions.Action{action}, nil
	}

	actionsList, actionTypeList := make([]actions.Action, 0, numActions), make([]string, 0, numActions)
	baseNameActionAdded := walletConfig.NameUsed
	for i := 0; i < numActions; i++ {
		actionType := availableActionTypes[rand.Intn(len(availableActionTypes))]

		if actionType == types.BaseNameAction {
			if baseNameActionAdded {
				continue
			}
			baseNameActionAdded = true
		}

		if (isDepositAction(actionType) || isWithdrawAction(actionType)) &&
			!isValidPoolAction(actionType, actionTypeList) {
			continue
		}

		action, err := r.GenerateSingleAction(actionType, acc)
		if err != nil {
			continue
		}

		actionsList = append(actionsList, action)
		actionTypeList = append(actionTypeList, string(actionType))
	}

	return actionsList, nil
}

func (r *Randomizer) GenerateSingleAction(actionType types.ActionType, acc *account.Account) (actions.Action, error) {
	switch actionType {
	case types.UniswapAction, types.PancakeAction, types.WoofiAction:
		return r.generateSwapAction(actionType, acc)
	case types.ZoraAction, types.NFT2MeAction:
		return r.generateNFTAction(actionType)
	case types.AaveETHDepositAction, types.AaveETHWithdrawAction, types.AaveUSDCSupplyAction,
		types.AaveUSDCWithdrawAction, types.MoonwellDepositAction, types.MoonwellWithdrawAction:
		return r.generatePoolAction(actionType, acc)
	case types.BridgeAction, types.DmailAction, types.CollectorModAction:
		return actions.Action{Type: actionType}, nil
	case types.BaseNameAction:
		return actions.Action{
			Type: actionType,
			BSNParams: types.BSNParams{
				Name: acc.BaseName,
			},
		}, nil
	default:
		return actions.Action{}, errors.New("неизвестный тип действия")
	}
}

func (r *Randomizer) generateNFTAction(actionType types.ActionType) (actions.Action, error) {
	r.nftMutex.Lock()
	defer r.nftMutex.Unlock()

	moduleNFTs, ok := r.availableNFTs[string(actionType)]
	if !ok || len(moduleNFTs) == 0 {
		return actions.Action{}, errors.New("нет доступных NFT для генерации для модуля " + string(actionType))
	}

	selectedContract, price := getRandomNFT(moduleNFTs)
	delete(moduleNFTs, selectedContract)

	return actions.Action{
		Type: actionType,
		NftMintParams: types.NftMintParams{
			MintCA: selectedContract,
			Price:  price,
		},
	}, nil
}

func (r *Randomizer) generateSwapAction(actionType types.ActionType, acc *account.Account) (actions.Action, error) {
	r.tokenMutex.Lock()
	defer r.tokenMutex.Unlock()

	filteredTokens := r.filtredTokenForDex(actionType)

	fromToken, err := r.selectFromToken(actionType, acc, filteredTokens)
	if err != nil {
		return actions.Action{}, err
	}

	toToken, err := r.selectToToken(acc, fromToken, filteredTokens)
	if err != nil {
		return actions.Action{}, err
	}

	return actions.Action{
		Type: actionType,
		DexParams: types.DexParams{
			FromToken: fromToken,
			ToToken:   toToken,
		},
	}, nil
}

func (r *Randomizer) generatePoolAction(actionType types.ActionType, acc *account.Account) (actions.Action, error) {
	lastActionIsDeposit := isLastActionDeposit(acc.LastPoolAction)

	if (isDepositAction(actionType) && lastActionIsDeposit) ||
		(isWithdrawAction(actionType) && !lastActionIsDeposit) {
		return actions.Action{}, fmt.Errorf("action %v out of sequence deposit-withdraw", string(actionType))
	}

	selectedToken, err := r.findEligibleToken(actionType, acc)
	if err != nil {
		return actions.Action{}, err
	}

	if !isValidTokenForLiquidAction(actionType, selectedToken) {
		return actions.Action{}, fmt.Errorf("the token %s doesn't fit for %s", selectedToken.Hex(), actionType)
	}

	acc.LastPoolAction = updateActionHistory(acc.LastPoolAction, actionType)

	return actions.Action{
		Type: actionType,
		LiquidParams: types.LiquidParams{
			Type:  string(actionType),
			Token: selectedToken,
		},
	}, nil
}

func (r *Randomizer) filtredTokenForDex(actionType types.ActionType) []common.Address {
	filtred := []common.Address{}
	for _, token := range r.availableTokens {
		if actionType == types.WoofiAction && token == config.USDbC {
			continue
		}
		filtred = append(filtred, token)
	}
	return filtred
}

func (r *Randomizer) selectFromToken(actionType types.ActionType, acc *account.Account, filtredTokens []common.Address) (common.Address, error) {
	if len(acc.LastSwaps) == 0 {
		return r.selectTokenWithHighestBalance(acc, filtredTokens)
	}

	fromToken := acc.LastSwaps[len(acc.LastSwaps)-1].To
	if actionType == types.WoofiAction && fromToken == config.USDbC {
		return common.Address{}, errors.New("woofi don't support usdbc")
	}

	return fromToken, nil
}

func (r *Randomizer) selectToToken(acc *account.Account, fromToken common.Address, filtredTokens []common.Address) (common.Address, error) {
	for attemps := 0; attemps <= len(filtredTokens); attemps++ {
		toToken, err := getRandomToken(filtredTokens, fromToken)
		if err != nil {
			continue
		}

		if fromToken == toToken {
			continue
		}

		if len(acc.LastSwaps) > 0 && acc.LastSwaps[len(acc.LastSwaps)-1].To == toToken {
			continue
		}

		acc.LastSwaps = append(acc.LastSwaps, models.SwapPair{From: fromToken, To: toToken})
		if len(acc.LastSwaps) > 10 {
			acc.LastSwaps = acc.LastSwaps[1:]
		}

		return toToken, nil
	}

	return common.Address{}, errors.New("failed generate 'to token' for swap")
}

func (r *Randomizer) selectTokenWithHighestBalance(acc *account.Account, filtredTokens []common.Address) (common.Address, error) {
	var selectedToken common.Address
	highestBalance := big.NewFloat(0)

	baseClient, exists := r.Clients["base"]
	if !exists {
		return common.Address{}, errors.New("haven't client for base chain")
	}

	for _, token := range filtredTokens {
		balance, err := baseClient.BalanceCheck(acc.Address, token)
		if err != nil {
			logger.GlobalLogger.Warn("failed check balance for: %s %v", token, err)
			continue
		}

		normilizeBalance, err := baseClient.NormalizeBalance(balance, token)
		if err != nil {
			logger.GlobalLogger.Warn("failed convert to $ balance for: %s %v", token, err)
			continue
		}

		if normilizeBalance.Cmp(highestBalance) > 0 {
			highestBalance = normilizeBalance
			selectedToken = token
		}
	}

	if selectedToken == (common.Address{}) {
		return common.Address{}, errors.New("token with balance > 0 not found")
	}

	return selectedToken, nil
}

func (r *Randomizer) findEligibleToken(actionType types.ActionType, acc *account.Account) (common.Address, error) {
	baseClient, exists := r.Clients["base"]
	if !exists {
		return common.Address{}, errors.New("haven't client for base chain")
	}
	highestBalance := big.NewInt(0)
	var selectedToken common.Address

	for _, token := range r.availableTokens {
		balance, err := baseClient.BalanceCheck(acc.Address, token)
		if err != nil {
			continue
		}

		if balance.Sign() > 0 && isValidTokenForLiquidAction(actionType, token) {
			if balance.Cmp(highestBalance) > 0 {
				highestBalance = balance
				selectedToken = token
			}
		}
	}

	if selectedToken == (common.Address{}) {
		return common.Address{}, fmt.Errorf("haven't eligble token for pool")
	}

	return selectedToken, nil
}
