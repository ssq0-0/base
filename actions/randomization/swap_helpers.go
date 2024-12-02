package randomization

import (
	"base/logger"
	"errors"
	"math/rand"

	"github.com/ethereum/go-ethereum/common"
)

func getRandomToken(availableTokens []common.Address, tokensToExclude ...common.Address) (common.Address, error) {
	excludeMap := make(map[string]bool)
	for _, token := range tokensToExclude {
		excludeMap[token.Hex()] = true
	}

	filteredTokens := []common.Address{}
	for _, token := range availableTokens {
		if !excludeMap[token.Hex()] {
			filteredTokens = append(filteredTokens, token)
		} else {
			logger.GlobalLogger.Debugf("Токен %s исключён из выбора", token.Hex())
		}
	}

	if len(filteredTokens) == 0 {
		return common.Address{}, errors.New("нет доступных токенов после исключения")
	}

	selected := filteredTokens[rand.Intn(len(filteredTokens))]
	return selected, nil
}
