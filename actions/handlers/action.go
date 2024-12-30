package handlers

import (
	"base/account"
	"base/config"
	"base/ethClient"
	"base/modules"
)

type ActionHandler interface {
	Execute(acc *account.Account, mods modules.Modules, client *ethClient.Client, config *config.Config) error
}
