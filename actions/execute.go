package actions

import (
	"base/account"
	"base/actions/handlers"
	"base/actions/types"
	"base/config"
	"base/ethClient"
	"base/modules"
	"errors"
)

type Action struct {
	Type          types.ActionType
	DexParams     types.DexParams
	NftMintParams types.NftMintParams
	BridgeParams  types.BridgeParams
	RefuelParams  types.RefuelParams
	LiquidParams  types.LiquidParams
	BSNParams     types.BSNParams
}

func (a Action) TakeActions(mods modules.Modules, acc *account.Account, action Action, client *ethClient.Client, config *config.Config) error {
	handler, err := GetActionHandler(action)
	if err != nil {
		return err
	}

	return handler.Execute(acc, mods, client, config)
}

func GetActionHandler(action Action) (handlers.ActionHandler, error) {
	switch action.Type {
	case types.UniswapAction, types.PancakeAction, types.WoofiAction, types.OdosAction, types.OpenOceanAction:
		return handlers.DexHandler{
			DexParams:  action.DexParams,
			ActionType: action.Type,
		}, nil
	case types.BridgeAction:
		return handlers.BridgeHandler{
			BridgeParams: action.BridgeParams,
		}, nil
	case types.ZoraAction:
		return handlers.ZoraHandler{
			NftMintParams: action.NftMintParams,
		}, nil
	case types.NFT2MeAction:
		return handlers.Nft2MeHandler{
			NftMintParams: action.NftMintParams,
		}, nil
	case types.AaveETHDepositAction, types.AaveETHWithdrawAction, types.AaveUSDCSupplyAction, types.AaveUSDCWithdrawAction:
		return handlers.AaveHandler{
			LiquidParams: action.LiquidParams,
		}, nil
	case types.MoonwellDepositAction, types.MoonwellWithdrawAction:
		return handlers.MoonwellHandler{
			LiquidParams: action.LiquidParams,
		}, nil
	case types.BaseNameAction:
		return handlers.BaseNameHandler{}, nil
	case types.DmailAction:
		return handlers.DmailHandler{}, nil
	case types.CollectorModAction:
		return &handlers.CollectorHandler{}, nil
	case types.RefuelAction:
		return &handlers.RefuelHandler{
			RefuelParams: action.RefuelParams,
		}, nil
	default:
		return nil, errors.New("неизвестный тип действия")
	}
}
