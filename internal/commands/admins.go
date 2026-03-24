package commands

import (
	"plugin/internal/bank"
	"plugin/internal/config"
	"plugin/internal/players"
	"plugin/internal/rcon"
	"plugin/internal/register"
	"plugin/internal/wallet"
)

func registerAdminCommands(
	cfg *config.Config,
	rc *rcon.RCON,
	reg *register.Register,

	players *players.Service,
	wallet *wallet.Service,
	bank *bank.Service,
) {
	// !freeze (!fz) <player>
	// lock controls on a player
	reg.RegisterCommand(register.Command{})

	// !dropgun (!dg) <player>
	// drop a players weapon
	reg.RegisterCommand(register.Command{})

	// !setspeed (!ss) <player (optional)> <amount>
	// set your or another players speed (float: 0.01-unlimited)
	reg.RegisterCommand(register.Command{})

	// !killplayer (!kpl) <player>
	// kill a player
	reg.RegisterCommand(register.Command{})

	// !hide (!hd) <player (optional)>
	// hide yourself or a player
	reg.RegisterCommand(register.Command{})
}
