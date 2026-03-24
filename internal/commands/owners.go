package commands

import (
	"plugin/internal/bank"
	"plugin/internal/config"
	"plugin/internal/players"
	"plugin/internal/rcon"
	"plugin/internal/register"
	"plugin/internal/wallet"
)

func registerOwnerCommands(
	cfg *config.Config,
	rc *rcon.RCON,
	reg *register.Register,

	players *players.Service,
	wallet *wallet.Service,
	bank *bank.Service,
) {
	// !gambling (!gmbl) <enable|disable|status>
	// enable / disable gambling or view status of gambling
	reg.RegisterCommand(register.Command{})

	// !maxbet (!mb) <amount|statu>
	// set the max bet amount or view the status of max bet
	reg.RegisterCommand(register.Command{})

	// !printmoney (!print) <amount>
	// print more money fuck the economy!
	reg.RegisterCommand(register.Command{})

	// !addowner <player> <xuid>
	// add a new owner
	reg.RegisterCommand(register.Command{})

	// !removeowner <xuid>
	// remove an owner
	reg.RegisterCommand(register.Command{})

	// !addadmin <player> <xuid>
	// add a new admin
	reg.RegisterCommand(register.Command{})

	// !removeadmin <xuid>
	// remove an admin
	reg.RegisterCommand(register.Command{})
}
