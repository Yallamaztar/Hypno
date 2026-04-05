package commands

import (
	"fmt"
	"plugin/internal/bank"
	"plugin/internal/config"
	"plugin/internal/players"
	"plugin/internal/rcon"
	"plugin/internal/register"
	"plugin/internal/utils"
	"plugin/internal/wallet"
	"strings"
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
	reg.RegisterCommand(&register.Command{
		Name:     "gambling",
		Aliases:  aliases{"gmbl"},
		MinLevel: levelOwner,
		MinArgs:  1,
		Help:     "Usage: ^6!gambling ^7(!gmbl) <enable|disable|status>",
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			switch strings.ToLower(args[0]) {
			case "enable", "on":
				cfg.Gambling.Enabled = true
				if err := cfg.Save(); err != nil {
					rc.Tell(clientNum, "Couldnt enabled gambling")
					return
				}

			case "disable", "off":
				cfg.Gambling.Enabled = false
				if err := cfg.Save(); err != nil {
					rc.Tell(clientNum, "Couldnt disable gambling")
					return
				}

			case "status":
				if cfg.Gambling.Enabled {
					rc.Tell(clientNum, "Gambling is currently ^6enabled")
				} else {
					rc.Tell(clientNum, "Gambling is currently ^1disabled")
				}

			default:
				rc.Tell(clientNum, "Usage: ^6!gambling ^7(!gmbl) <enable|disable|status>")
			}
		},
	})

	// !maxbet (!mb) <amount|status>
	// set the max bet amount or view the status of max bet
	reg.RegisterCommand(&register.Command{
		Name:     "maxbet",
		Aliases:  aliases{"mb", "max"},
		MinLevel: levelOwner,
		MinArgs:  1,
		Help:     "Usage: ^6!maxbet ^7<amount|status>",
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			if strings.ToLower(args[0]) == "status" {
				max := cfg.Economy.MaxBet
				if max == 0 {
					rc.Tell(clientNum, "Max bet is currently disabled")
				} else {
					rc.Tell(clientNum, fmt.Sprintf("Max bet is currently %s%d", cfg.Gambling.Currency, max))
				}
			}

			amount := utils.ParseAmount(args[0])
			if amount == 0 {
				if strings.TrimSpace(args[0]) != "0" {
					rc.Tell(clientNum, "Invalid amount")
					return
				}
			}

			if amount < 0 {
				rc.Tell(clientNum, "Invalid amount")
				return
			}

			cfg.Economy.MaxBet = int(amount)
			if err := cfg.Save(); err != nil {
				rc.Tell(clientNum, "Failed to set max bet")
				return
			}

			if amount == 0 {
				rc.Say("Max bet has been ^6disabled")
			} else {
				rc.Say(fmt.Sprintf("Max bet has been set to ^6 %s%d", cfg.Gambling.Currency, amount))
			}
		},
	})

	// !printmoney (!print) <amount>
	// print more money fuck the economy!
	reg.RegisterCommand(&register.Command{
		Name:     "printmoney",
		Aliases:  aliases{"print", "yidish", "shalom"},
		MinLevel: levelOwner,
		MinArgs:  1,
		Help:     "Usage: ^6!printmoney ^7<amount>",
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			printToBank := false
			if len(args) >= 2 && args[1] == "--bank" || args[1] == "-b" {
				printToBank = true
			}

			amount := utils.ParseAmount(args[0])
			if amount <= 0 {
				rc.Tell(clientNum, "Invalid amount")
				return
			}

			if printToBank {
				if err := bank.Deposit(int(amount)); err != nil {
					rc.Tell(clientNum, "Failed to print money")
					return
				}
			}

			if err := wallet.Deposit(playerID, int(amount)); err != nil {
				rc.Tell(clientNum, "Failed to print money")
				return
			}
		},
	})

	// !addowner <xuid>
	// add a new owner
	reg.RegisterCommand(&register.Command{
		Name:     "addowner",
		Aliases:  aliases{"add"},
		MinLevel: levelOwner,
		MinArgs:  1,
		Help:     "!addowner <xuid>",
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			target, err := players.GetByXUID(args[0])
			if err != nil {
				rc.Tell(clientNum, "Player couldnt be found ("+args[0]+")")
				return
			}

			players.UpdateLevel(target.ID, levelOwner)
			rc.Tell(clientNum, target.Name+" level set to owner")
		},
	})

	// !removeowner <xuid>
	// remove an owner
	reg.RegisterCommand(&register.Command{
		Name:     "removeowner",
		Aliases:  aliases{"remove"},
		MinLevel: levelOwner,
		MinArgs:  1,
		Help:     "!removeowner <xuid>",
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			target, err := players.GetByXUID(args[0])
			if err != nil {
				rc.Tell(clientNum, "Player couldnt be found ("+args[0]+")")
				return
			}

			players.UpdateLevel(target.ID, levelUser)
			rc.Tell(clientNum, target.Name+" level set to user")
		},
	})

	// !addadmin <player> <xuid>
	// add a new admin
	reg.RegisterCommand(&register.Command{
		Name:     "addadmin",
		Aliases:  aliases{"add"},
		MinLevel: levelOwner,
		MinArgs:  1,
		Help:     "!addadmin <xuid>",
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			target, err := players.GetByXUID(args[0])
			if err != nil {
				rc.Tell(clientNum, "Player couldnt be found ("+args[0]+")")
				return
			}

			players.UpdateLevel(target.ID, levelAdmin)
			rc.Tell(clientNum, target.Name+" level set to admin")
		},
	})

	// !removeadmin <xuid>
	// remove an admin
	reg.RegisterCommand(&register.Command{
		Name:     "removeadmin",
		Aliases:  aliases{"remove"},
		MinLevel: levelOwner,
		MinArgs:  1,
		Help:     "!removeadmin <xuid>",
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			target, err := players.GetByXUID(args[0])
			if err != nil {
				rc.Tell(clientNum, "Player couldnt be found ("+args[0]+")")
				return
			}

			players.UpdateLevel(target.ID, levelUser)
			rc.Tell(clientNum, target.Name+" level set to user")
		},
	})
}
