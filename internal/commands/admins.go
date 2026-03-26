package commands

import (
	"fmt"
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
	reg.RegisterCommand(register.Command{
		Name:     "freeze",
		Aliases:  []string{"fz", "freez"},
		MinLevel: levelAdmin,
		MinArgs:  0,
		Help:     "Usage: ^6!freeze ^7<player>",
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			args, err := parseArgs(args)
			if err != nil {
				rc.Tell(clientNum, err.Error())
				return
			}

			cn, err := resolveClientNum(rc, reg, clientNum, args)
			if err != nil {
				rc.Tell(clientNum, err.Error())
				return
			}

			if cn == -1 {
				rc.Tell(clientNum, "player ^6couldnt ^7be found")
				return
			}

			rc.SetInDvar(fmt.Sprintf("killplayer %d", cn))
		},
	})

	// !dropgun (!dg) <player>
	// drop a players weapon
	reg.RegisterCommand(register.Command{
		Name:     "dropgun",
		Aliases:  []string{"dg", "drop"},
		MinLevel: levelAdmin,
		MinArgs:  0,
		Help:     "Usage: ^6!dropgun ^7<player>",
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			args, err := parseArgs(args)
			if err != nil {
				rc.Tell(clientNum, err.Error())
				return
			}

			cn, err := resolveClientNum(rc, reg, clientNum, args)
			if err != nil {
				rc.Tell(clientNum, err.Error())
				return
			}

			if cn == -1 {
				rc.Tell(clientNum, "player ^6couldnt ^7be found")
				return
			}

			rc.SetInDvar(fmt.Sprintf("killplayer %d", cn))
		},
	})

	// !setspeed (!ss) <player (optional)> <amount>
	// set your or another players speed (float: 0.01-unlimited)
	reg.RegisterCommand(register.Command{
		Name:     "setspeed",
		Aliases:  []string{"ss", "sets", "sspeed"},
		MinLevel: levelAdmin,
		MinArgs:  0,
		Help:     "Usage ^6!setspeed ^7<player> <amount>",
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			args, err := parseArgs(args)
			if err != nil {
				rc.Tell(clientNum, err.Error())
				return
			}

			cn, err := resolveClientNum(rc, reg, clientNum, args)
			if err != nil {
				rc.Tell(clientNum, err.Error())
				return
			}

			if cn == -1 {
				rc.Tell(clientNum, "player ^6couldnt ^7be found")
				return
			}

			rc.SetInDvar(fmt.Sprintf("setspeed %d", cn))
		},
	})

	// !xuid (!info) <player>
	// show player name, xuid and client num
	reg.RegisterCommand(register.Command{
		Name:     "xuid",
		Aliases:  []string{"info"},
		MinLevel: levelAdmin,
		MinArgs:  1,
		Help:     "Usage: ^6!xuid ^7<player>",
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			args, err := parseArgs(args)
			if err != nil {
				rc.Tell(clientNum, err.Error())
				return
			}

			cn, err := resolveClientNum(rc, reg, clientNum, args)
			if err != nil {
				rc.Tell(clientNum, err.Error())
				return
			}

			if cn == -1 {
				rc.Tell(clientNum, "player ^6couldnt ^7be found")
				return
			}

			guid := rc.GUIDByClientNum(uint8(cn))
			if guid == "" {
				rc.Tell(clientNum, "player ^6isnt ^7online")
				return
			}

			target, err := players.GetByGUID(guid)
			if err != nil {
				rc.Tell(clientNum, "player ^6couldnt ^7be found")
				return
			}

			rc.Tell(clientNum, fmt.Sprintf("[^6%s^7] XUID: ^6%s^7 | ClientNum: ^6%d", target.Name, target.XUID, cn))
		},
	})

	// !swap (!swp) <player> <player (optional)>
	// swap places with player or player with another player
	reg.RegisterCommand(register.Command{
		Name:     "swap",
		Aliases:  []string{"swp"},
		MinLevel: levelAdmin,
		MinArgs:  1,
		Help:     "Usage: ^6!swap ^7<player> <player (optional)>",
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			args, err := parseArgs(args)
			if err != nil {
				rc.Tell(clientNum, err.Error())
				return
			}

			cn1, cn2, err := resolveClientNums(rc, reg, clientNum, args)
			if err != nil {
				rc.Tell(clientNum, err.Error())
				return
			}

			rc.SetInDvar(fmt.Sprintf("swap %d %d", cn1, cn2))
		},
	})

	// !dropgun (!dg) <player>
	// drop a players gun
	reg.RegisterCommand(register.Command{})

	// !teleport (!tp) <player> <player (optional)>
	// teleport to a player or a player to a player
	reg.RegisterCommand(register.Command{})

	// !sayas (!says) <player> <message>
	// send a message as a player
	reg.RegisterCommand(register.Command{})

	// !loadout (!ld) <player>
	// give a player random loadout
	reg.RegisterCommand(register.Command{})

	// !stealmoney (!steal) <player> <amount>
	// steal money from a player
	reg.RegisterCommand(register.Command{})

	// !givemoney (!give) <player> <amount>
	// give money to a player
	reg.RegisterCommand(register.Command{})

	// !giveall (!ga) <amount>
	// gives all players ingame money
	reg.RegisterCommand(register.Command{})

	// !setposition (!sp) <player (optional) <x> <y> <z>
	// set players origin to given coords
	reg.RegisterCommand(register.Command{})

	/*
	 * IW4M-Admin gameinterface overrides
	 * these just work way faster than iw4m-admins
	 * slow gameinterface, if you dont have IW4M-Admin,
	 * these commands will still work
	 */

	// !giveweapon (!gw) <player> <weapon>
	reg.RegisterCommand(register.Command{})

	// !takeweapons (!tw) <player>
	reg.RegisterCommand(register.Command{})

	// !switchteams (!st) <player>
	reg.RegisterCommand(register.Command{})

	// !hide (!hd) <player (optional)>
	// hide yourself or a player
	reg.RegisterCommand(register.Command{
		Name:     "hide",
		Aliases:  []string{"hd", "hid", "invisible", "invis"},
		MinLevel: levelAdmin,
		MinArgs:  0,
		Help:     "Usage: ^6!hide ^7<player>",
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			args, err := parseArgs(args)
			if err != nil {
				rc.Tell(clientNum, err.Error())
				return
			}

			cn, err := resolveClientNum(rc, reg, clientNum, args)
			if err != nil {
				rc.Tell(clientNum, err.Error())
				return
			}

			if cn == -1 {
				rc.Tell(clientNum, "player ^6couldnt ^7be found")
				return
			}

			rc.SetInDvar(fmt.Sprintf("hideplayer %d", cn))
		},
	})

	// !alert (!alr) <player> <message>
	reg.RegisterCommand(register.Command{})

	// !kill (!kpl) <player>
	// kill a player
	reg.RegisterCommand(register.Command{
		Name:     "killplayer",
		Aliases:  []string{"kpl", "kplayer", "killp"},
		MinLevel: levelAdmin,
		MinArgs:  0,
		Help:     "Usage: ^6!killplayer ^7<player>",
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			args, err := parseArgs(args)
			if err != nil {
				rc.Tell(clientNum, err.Error())
				return
			}

			cn, err := resolveClientNum(rc, reg, clientNum, args)
			if err != nil {
				rc.Tell(clientNum, err.Error())
				return
			}

			if cn == -1 {
				rc.Tell(clientNum, "player ^6couldnt ^7be found")
				return
			}

			rc.SetInDvar(fmt.Sprintf("killplayer %d", cn))
		},
	})

	// !setspectator (!spec) <player>
	// set player to codcaster
	reg.RegisterCommand(register.Command{})
}
