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
	"strconv"
	"strings"
	"time"
)

func registerDeveloperCommands(
	cfg *config.Config,
	rc *rcon.RCON,
	reg *register.Register,

	players *players.Service,
	wallet *wallet.Service,
	bank *bank.Service,
	log *logger.Logger,
) {
	// !printmoney (!print) <amount>
	// print more money fuck the economy!
	reg.RegisterCommand(&register.Command{
		Name:     "printmoney",
		Aliases:  aliases{"print", "yidish", "shalom"},
		MinLevel: levelDeveloper,
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

				log.Infof("%s (%d) printed %s%s to the bank\n", playerName, playerID, cfg.Gambling.Currency, utils.FormatMoney(int(amount)))
				rc.Tell(clientNum, fmt.Sprintf("Printed ^6%s%s^7 to the bank", cfg.Gambling.Currency, utils.FormatMoney(int(amount))))
				return
			}

			if err := wallet.Deposit(playerID, int(amount)); err != nil {
				rc.Tell(clientNum, "Failed to print money")
				return
			}

			log.Infof("%s (%d) printed %s%s to the wallet\n", playerName, playerID, cfg.Gambling.Currency, utils.FormatMoney(int(amount)))
			rc.Tell(clientNum, fmt.Sprintf("Printed ^6%s%s^7 to wallet", cfg.Gambling.Currency, utils.FormatMoney(int(amount))))
		},
	})

	// !xuid (!info) <player>
	// show player name, xuid and client num
	reg.RegisterCommand(&register.Command{
		Name:     "xuid",
		Aliases:  aliases{"info"},
		MinLevel: levelDeveloper,
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
			rc.Say("Crashing the plugin in 5 seconds")
			log.Warnln("Crashing the plugin on purpose in 5 seconds")

			for i := 4; i >= 1; i-- {
				time.Sleep(time.Second)
				rc.Say(fmt.Sprintf("%d", i))
			}
			time.Sleep(time.Second)

			rc.Say("Crashing plugin, good bye!")
			log.Warnln("Crashing plugin, good bye!")

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

	// !lookup (!find) <name>
	// look up players xuid from db
	reg.RegisterCommand(&register.Command{
		Name:     "lookup",
		Aliases:  aliases{"find"},
		MinLevel: levelDeveloper,
		MinArgs:  1,
		Help:     "Usage: ^6!lookup^7 <name>",
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			target, err := players.GetByGUID(*reg.FindPlayerPartial(strings.Join(args, " ")).GUID)
			if err != nil {
				rc.Tell(clientNum, "Player couldnt be found ("+args[0]+")")
				return
			}

			rc.Tell(clientNum, target.Name+" has XUID: "+target.XUID)
		},
	})

	// !discordinvite (!invite) <link>
	// change the discord invite in config
	reg.RegisterCommand(&register.Command{
		Name:     "discordinvite",
		Aliases:  aliases{"invite"},
		MinLevel: levelDeveloper,
		MinArgs:  1,
		Help:     "Usage: ^6!discordinvite ^7<link>",
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			cfg.Discord.InviteLink = args[0]
			if err := cfg.Save(); err != nil {
				rc.Tell(clientNum, "Failed to save config")
				return
			}

			log.Infof("%s (%d) updated the discord invite link to %s\n", playerName, playerID, args[0])
			rc.Tell(clientNum, "Discord invite updated")
		},
	})
}
