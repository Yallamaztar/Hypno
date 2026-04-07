package commands

import (
	"fmt"
	"plugin/internal/bank"
	"plugin/internal/config"
	"plugin/internal/logger"
	"plugin/internal/players"
	"plugin/internal/rcon"
	"plugin/internal/register"
	"plugin/internal/utils"
	"plugin/internal/wallet"
	"strings"
)

func registerAdminCommands(
	cfg *config.Config,
	rc *rcon.RCON,
	reg *register.Register,

	players *players.Service,
	wallet *wallet.Service,
	bank *bank.Service,

	log *logger.Logger,
) {
	// !freeze (!fz) <player>
	// lock controls on a player
	reg.RegisterCommand(&register.Command{
		Name:     "freeze",
		Aliases:  aliases{"fz", "freez"},
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
				rc.Tell(clientNum, "^6"+rc.NameByClientNum(cn)+" ^7couldnt be found")
				return
			}

			rc.SetInDvar(fmt.Sprintf("freeze %d %d", clientNum, cn))
		},
	})

	// !setspeed (!ss) <player (optional)> <amount>
	// set your or another players speed (float: 0.01-unlimited)
	reg.RegisterCommand(&register.Command{
		Name:     "setspeed",
		Aliases:  aliases{"ss", "sets", "sspeed"},
		MinLevel: levelAdmin,
		MinArgs:  1,
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
				rc.Tell(clientNum, "^6"+rc.NameByClientNum(cn)+" ^7couldnt be found")
				return
			}

			var speed string
			if len(args) == 1 {
				rc.Tell(clientNum, fmt.Sprintf("Setting your speed to: %s", speed))
				speed = args[0]
			} else {
				rc.Tell(clientNum, fmt.Sprintf("Setting %s's speed to %s", rc.NameByClientNum(cn), speed))
				speed = args[1]
			}

			rc.SetInDvar(fmt.Sprintf("setspeed %d %s", cn, speed))
		},
	})

	// !swap (!swp) <player> <player (optional)>
	// swap places with player or player with another player
	reg.RegisterCommand(&register.Command{
		Name:     "swap",
		Aliases:  aliases{"swp"},
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

			name2 := rc.NameByClientNum(cn2)
			if name2 == "" {
				rc.Tell(clientNum, "couldnt find target")
				return
			}

			if int(clientNum) == cn1 {
				rc.Tell(clientNum, "swapping with ^6"+name2)
			} else {
				name := rc.NameByClientNum(cn1)
				rc.Tell(clientNum, "swapping ^6"+name+" ^7with ^6"+name2)
			}

			rc.SetInDvar(fmt.Sprintf("swap %d %d", cn1, cn2))
		},
	})

	// !dropgun (!dg) <player>
	// drop a players gun
	reg.RegisterCommand(&register.Command{
		Name:     "dropgun",
		Aliases:  aliases{"dg", "dw", "dropweapon"},
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
				rc.Tell(clientNum, "^6"+rc.NameByClientNum(cn)+" ^7couldnt be found")
				return
			}

			if int(clientNum) == cn {
				rc.Tell(clientNum, "dropping your weapon")
			} else {
				rc.Tell(clientNum, "Dropping "+rc.NameByClientNum(cn)+"'s weapon")
			}

			rc.SetInDvar(fmt.Sprintf("dropgun %d", cn))
		},
	})

	// !teleport (!tp) <player> <player (optional)>
	// teleport to a player or a player to a player
	reg.RegisterCommand(&register.Command{
		Name:     "teleport",
		Aliases:  aliases{"tp"},
		MinLevel: levelAdmin,
		MinArgs:  1,
		Help:     "Usage: ^6!teleport ^7<player> <player (optional)>",
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

			if cn1 == cn2 {
				rc.Tell(clientNum, "cannot tp same player")
				return
			}

			rc.SetInDvar(fmt.Sprintf("teleport %d %d", cn1, cn2))
		},
	})

	// !sayas (!says) <player> <message>
	// send a message as a player
	reg.RegisterCommand(&register.Command{
		Name:     "sayas",
		Aliases:  aliases{"says", "sa"},
		MinLevel: levelAdmin,
		MinArgs:  2,
		Help:     "Usage: ^6!sayas ^7<player> <message> [--dead (-d) | --enemy (-e)]",
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			args, err := parseArgs(args)
			if err != nil {
				rc.Tell(clientNum, err.Error())
				return
			}

			var dead, enemy bool
			filtered := make([]string, 0, len(args))

			for _, arg := range args {
				switch arg {
				case "-d", "--dead":
					dead = true
					continue
				case "-e", "--enemy":
					enemy = true
					continue
				}
				filtered = append(filtered, arg)
			}

			if len(filtered) < 2 {
				rc.Tell(clientNum, "Usage: ^6!sayas ^7<player> <message> [--dead (-d) | --enemy (-e)]")
				return
			}

			target := reg.FindPlayerPartial(filtered[0])
			if target == nil || *target.ClientNum == -1 {
				target = &register.PlayerInfo{Name: filtered[0]}
			}

			message := strings.Join(filtered[1:], " ")

			prefix := "^2"
			if enemy {
				prefix = "^1"
			}

			channel := "[Playin-All]"
			if dead {
				channel = "[Dead-All]"
			}

			rc.SayRaw(prefix + target.Name + " " + channel + ": ^7" + message)
			rc.Tell(clientNum, "Message sent as "+target.Name)
		},
	})

	// !stealmoney (!steal) <player> <amount>
	// steal money from a player (cfg.Economy.MaxSteal)
	reg.RegisterCommand(&register.Command{
		Name:     "stealmoney",
		Aliases:  aliases{"steal", "take"},
		MinLevel: levelAdmin,
		MinArgs:  2,
		Help:     "Usage: ^6!stealmoney ^7<player> <amount>",
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			args, err := parseArgs(args)
			if err != nil {
				rc.Tell(clientNum, err.Error())
				return
			}

			t := reg.FindPlayerPartial(args[0])
			if t == nil {
				rc.Tell(clientNum, args[0]+" not found")
				return
			}

			target, err := players.GetByGUID(*t.GUID)
			if err != nil {
				rc.Tell(clientNum, t.Name+" not found")
				return
			}

			bal, err := wallet.Balance(target.ID)
			if err != nil {
				rc.Tell(clientNum, "couldnt get "+t.Name+"'s wallet")
				return
			}

			amount, err := utils.ParseAmountArg(args[1], bal)
			if err != nil {
				rc.Tell(clientNum, err.Error())
				return
			}

			if level == levelAdmin {
				max := int(float64(bal) * cfg.Economy.MaxSteal)
				if amount > max {
					rc.Tell(clientNum, fmt.Sprintf("You can only steal ^6%s%s^7 from %s", cfg.Gambling.Currency, utils.FormatMoney(max), target.Name))
					return
				}
			}

			if amount > bal {
				rc.Tell(clientNum, fmt.Sprintf("%s has ^1insufficient funds^7 (%s%s)", target.Name, cfg.Gambling.Currency, utils.FormatMoney(bal)))
				return
			}

			if err := wallet.WalletToWallet(target.ID, playerID, amount); err != nil {
				rc.Tell(clientNum, err.Error())
				return
			}

			log.Infof("%s (%d) stole %s%s from %s (%d)\n", playerName, playerID, cfg.Gambling.Currency, utils.FormatMoney(amount), target.Name, target.ID)
			rc.Tell(clientNum, fmt.Sprintf("Took ^6%s%s ^7from %s", cfg.Gambling.Currency, utils.FormatMoney(amount), target.Name))
			rc.Tell(uint8(*t.ClientNum), fmt.Sprintf("%s took ^6%s%s ^7from you LOL", playerName, cfg.Gambling.Currency, utils.FormatMoney(amount)))
		},
	})

	// !givemoney (!give) <player> <amount>
	// give money to a player (cfg.Economy.MaxGive)
	reg.RegisterCommand(&register.Command{
		Name:     "givemoney",
		Aliases:  aliases{"give", "gi"},
		MinLevel: levelAdmin,
		MinArgs:  2,
		Help:     "Usage: ^6!givemoney ^7<player> <amount>",
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			args, err := parseArgs(args)
			if err != nil {
				log.Errorln("Failed to parse args for !givemoney command")
				rc.Tell(clientNum, err.Error())
				return
			}

			t := reg.FindPlayerPartial(args[0])
			if t == nil {
				log.Errorln("Failed to find player")
				rc.Tell(clientNum, args[0]+" not found")
				return
			}

			target, err := players.GetByGUID(*t.GUID)
			if err != nil {
				log.Errorln("Failed to get player by GUID")
				rc.Tell(clientNum, t.Name+" not found")
				return
			}

			amount := utils.ParseAmount(args[1])
			if amount <= 0 {
				log.Errorln("Failed to parse amount for !givemoney command")
				rc.Tell(clientNum, "Invalid amount")
				return
			}

			if target.Level == levelAdmin {
				bankbal, err := bank.Balance()
				if err != nil {
					log.Errorln("Failed to get bank balance")
					rc.Tell(clientNum, "Couldnt get bank balance")
					return
				}

				max := int(float64(bankbal) * cfg.Economy.MaxGive)
				if int(amount) > max {
					rc.Tell(clientNum, fmt.Sprintf("You can ^1only^7 give up to ^6%s%s^7", cfg.Gambling.Currency, utils.FormatMoney(max)))
					return
				}
			}

			if err := bank.Withdraw(int(amount)); err != nil {
				log.Errorln("Failed to withdraw money from the bank")
				rc.Tell(clientNum, "Couldnt get money from the bank")
				return
			}

			if err := wallet.Deposit(target.ID, int(amount)); err != nil {
				log.Errorln("Failed to deposit money into player's wallet")
				rc.Tell(clientNum, "Transfer failed")
				return
			}

			log.Infof("%s (%d) gave %s%s to %s (%d)\n", playerName, playerID, cfg.Gambling.Currency, utils.FormatMoney(int(amount)), target.Name, target.ID)
			rc.Tell(clientNum, fmt.Sprintf("Gave %s ^6%s%s", target.Name, cfg.Gambling.Currency, utils.FormatMoney(int(amount))))
			rc.Tell(uint8(*t.ClientNum), fmt.Sprintf("%s gave you ^6%s%s", playerName, cfg.Gambling.Currency, utils.FormatMoney(int(amount))))
		},
	})

	// !giveall (!ga) <amount>
	// gives all players online money
	reg.RegisterCommand(&register.Command{
		Name:     "giveall",
		Aliases:  aliases{"givea", "ga"},
		MinLevel: levelAdmin,
		MinArgs:  1,
		Help:     "Usage: ^6!giveall <amount>",
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			amount := utils.ParseAmount(args[0])
			if amount <= 0 {
				rc.Tell(clientNum, "Invalid amount")
				return
			}

			status, err := rc.Status()
			if err != nil {
				rc.Tell(clientNum, "^1error ^7occurred, please try again later")
				return
			}

			for _, p := range status.Players {
				t, err := players.GetByGUID(p.GUID)
				if err != nil {
					rc.Tell(clientNum, "couldnt get player ("+p.GUID+")")
					continue
				}

				bal, err := bank.Balance()
				if err != nil {
					rc.Tell(clientNum, "couldnt get bank balance")
					continue
				}

				if int(amount) > bal {
					rc.Tell(clientNum, "bank ^1ran ^7out of money boohoo")
					break
				}

				if err := bank.Withdraw(int(amount)); err != nil {
					rc.Tell(clientNum, "Couldnt get money from the bank")
					break
				}

				if err := wallet.Deposit(t.ID, int(amount)); err != nil {
					rc.Tell(clientNum, "Transfer failed")
					break
				}
			}

			log.Infof("%s (%d) gave %s%s to %d players\n", playerName, playerID, cfg.Gambling.Currency, utils.FormatMoney(int(amount)), len(status.Players))
			rc.Say(fmt.Sprintf("%s gave ^6%s%s to everyone", playerName, cfg.Gambling.Currency, utils.FormatMoney(int(amount))))
		},
	})

	/*
	 * IW4M-Admin gameinterface overrides
	 * these just work way faster than iw4m-admins
	 * slow gameinterface, if you dont have IW4M-Admin,
	 * these commands will still work
	 */

	// !giveweapon (!gw) <player (optional)> <weapon>
	reg.RegisterCommand(&register.Command{
		Name:     "giveweapon",
		Aliases:  aliases{"gw"},
		MinLevel: levelAdmin,
		MinArgs:  0,
		Help:     "Usage: ^6!giveweapon ^7<player (optional)> <weapon>",
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
				rc.Tell(clientNum, "^6"+rc.NameByClientNum(cn)+" ^7couldnt be found")
				return
			}

			weapon := args[len(args)-1]
			if int(clientNum) == cn {
				rc.Tell(clientNum, "giving you weapon:")
				rc.Tell(clientNum, "^6"+weapon)
			} else {
				rc.Tell(clientNum, "giving "+rc.NameByClientNum(cn)+" weapon:")
				rc.Tell(clientNum, "^6"+weapon)
			}

			rc.SetInDvar(fmt.Sprintf("giveweapon %d %s", cn, weapon))
		},
	})

	// !takeweapons (!tw) <player (optional)>
	reg.RegisterCommand(&register.Command{
		Name:     "takeweapons",
		Aliases:  aliases{"tw"},
		MinLevel: levelAdmin,
		MinArgs:  0,
		Help:     "usage: ^6!takeweapons ^7<player>",
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
				rc.Tell(clientNum, "^6"+rc.NameByClientNum(cn)+" ^7couldnt be found")
				return
			}

			if int(clientNum) == cn {
				rc.Tell(clientNum, "Taking your weapons")
			} else {
				rc.Tell(clientNum, "Taking "+rc.NameByClientNum(cn)+"'s weapons")
			}

			rc.SetInDvar(fmt.Sprintf("takeweapons %d", cn))
		},
	})

	// !switchteams (!st) <player (optional)>
	reg.RegisterCommand(&register.Command{
		Name:     "switchteams",
		Aliases:  aliases{"st"},
		MinLevel: levelAdmin,
		MinArgs:  0,
		Help:     "usage: ^6!switchteams ^7<player (optional)>",
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
				rc.Tell(clientNum, "^6"+rc.NameByClientNum(cn)+" ^7couldnt be found")
				return
			}

			rc.Tell(clientNum, "Switching teams")
			rc.SetInDvar(fmt.Sprintf("switchteams %d", cn))
		},
	})

	// !hide (!hd) <player (optional)>
	// hide yourself or a player
	reg.RegisterCommand(&register.Command{
		Name:     "hide",
		Aliases:  aliases{"hd", "hid", "invisible", "invis"},
		MinLevel: levelAdmin,
		MinArgs:  0,
		Help:     "Usage: ^6!hide ^7<player (optional)>",
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
				rc.Tell(clientNum, "^6"+rc.NameByClientNum(cn)+" ^7couldnt be found")
				return
			}

			rc.SetInDvar(fmt.Sprintf("hideplayer %d %d", clientNum, cn))
		},
	})

	// !alert (!alr) <player> <message>
	reg.RegisterCommand(&register.Command{
		Name:     "alert",
		Aliases:  aliases{"alr"},
		MinLevel: levelAdmin,
		MinArgs:  2,
		Help:     "usage: ^6!alert ^7<player> <message>",
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
				rc.Tell(clientNum, "^6"+rc.NameByClientNum(cn)+" ^7couldnt be found")
				return
			}

			rc.SetInDvar(fmt.Sprintf("alert %d %s", cn, strings.Join(args[1:], " ")))
		},
	})

	// !kill (!kpl) <player>
	// kill a player
	reg.RegisterCommand(&register.Command{
		Name:     "killplayer",
		Aliases:  aliases{"kpl", "kplayer", "killp"},
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
				rc.Tell(clientNum, "^6"+rc.NameByClientNum(cn)+" ^7couldnt be found")
				return
			}

			rc.Tell(clientNum, "Killing player ^6"+rc.NameByClientNum(cn))
			rc.SetInDvar(fmt.Sprintf("killplayer %d", cn))
		},
	})

	// !setspectator (!spec) <player>
	// set player to codcaster
	reg.RegisterCommand(&register.Command{
		Name:     "setspectator",
		Aliases:  aliases{"spec"},
		MinLevel: levelAdmin,
		MinArgs:  1,
		Help:     "usage: ^6!setspectator ^7<player>",
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
				rc.Tell(clientNum, "^6"+rc.NameByClientNum(cn)+" ^7couldnt be found")
				return
			}

			rc.Tell(clientNum, "Setting ^6"+rc.NameByClientNum(cn)+" ^7to spectator")
			rc.SetInDvar(fmt.Sprintf("setspectator %d", cn))
		},
	})

	// !goto (!g2) <player (optional)> <x> <y> <z>
	// set players origin to given coords
	reg.RegisterCommand(&register.Command{
		Name:     "goto",
		Aliases:  aliases{"g2"},
		MinLevel: levelAdmin,
		MinArgs:  3,
		Help:     "Usage: ^6!goto ^7<player (optional)> <x> <y> <z>",
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
				rc.Tell(clientNum, "^6"+rc.NameByClientNum(cn)+" ^7couldnt be found")
				return
			}

			rc.Tell(clientNum, "Setting origin to "+args[1]+", "+args[2]+", "+args[3])
			rc.SetInDvar(fmt.Sprintf("setorigin %d %s %s %s", cn, args[1], args[2], args[3]))
		},
	})
}
