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
	"strconv"
	"strings"
)

func registerDeveloperCommands(
	cfg *config.Config,
	rc *rcon.RCON,
	reg *register.Register,

	players *players.Service,
	wallet *wallet.Service,
	bank *bank.Service,
) {
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

				rc.Tell(clientNum, fmt.Sprintf("Printed ^6%s%s^7 to the bank", cfg.Gambling.Currency, utils.FormatMoney(int(amount))))
				return
			}

			if err := wallet.Deposit(playerID, int(amount)); err != nil {
				rc.Tell(clientNum, "Failed to print money")
				return
			}
			rc.Tell(clientNum, fmt.Sprintf("Printed ^6%s%s^7 to wallet", cfg.Gambling.Currency, utils.FormatMoney(int(amount))))
		},
	})

	// !xuid (!info) <player>
	// show player name, xuid and client num
	reg.RegisterCommand(&register.Command{
		Name:     "xuid",
		Aliases:  aliases{"info"},
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

	// !crash (!panic)
	// make the plugin crash on purpouse
	reg.RegisterCommand(&register.Command{
		Name:     "crash",
		Aliases:  aliases{"panic"},
		MinLevel: levelDeveloper,
		MinArgs:  0,
		Help:     "Usage: ^6!crash",
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			panic("Crashing the plugin on purpose")
		},
	})

	// !rcon (!rc)
	// send rcon commands
	reg.RegisterCommand(&register.Command{
		Name:     "rcon",
		Aliases:  aliases{"rc"},
		MinLevel: levelDeveloper,
		MinArgs:  2,
		Help:     "Usage: ^6!rcon ^7<command> <args...>",
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			switch strings.ToLower(args[0]) {
			case "say", "s":
				rc.SayRaw(strings.Join(args[1:], " "))
			case "tell":
				if len(args) < 3 {
					rc.Tell(clientNum, "Usage: ^6!rcon ^7tell <client_num> <message>")
					return
				}

				cn, err := strconv.Atoi(args[1])
				if err != nil {
					rc.Tell(clientNum, "Invalid client number")
					return
				}

				rc.Tell(uint8(cn), strings.Join(args[2:], " "))

			case "dvar":
				if len(args) < 3 {
					rc.Tell(clientNum, "Usage: ^6!rcon ^7dvar <dvar> <value>")
					return
				}

				rc.SetDvar(args[1], strings.Join(args[2:], " "))

			case "guid":
				if len(args) < 2 {
					rc.Tell(clientNum, "Usage: ^6!rcon ^7guid <client_num>")
					return
				}

				cn, err := strconv.Atoi(args[1])
				if err != nil {
					rc.Tell(clientNum, "Invalid client number")
					return
				}

				guid := rc.GUIDByClientNum(uint8(cn))
				if guid == "" {
					rc.Tell(clientNum, "player ^6isnt ^7online")
					return
				}

				rc.Tell(clientNum, fmt.Sprintf("GUID: ^6%s", guid))
			}
		},
	})
}
