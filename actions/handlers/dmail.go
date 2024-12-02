package handlers

import (
	"base/account"
	"base/config"
	"base/ethClient"
	"base/modules"
)

type DmailHandler struct {
}

func (dh DmailHandler) Execute(acc *account.Account, mods modules.Modules, client *ethClient.Client, config *config.Config) error {
	return mods.Dmail.SendMail(acc)
}
