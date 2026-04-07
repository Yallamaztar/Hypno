package commands

import (
	"fmt"
	"plugin/internal/config"
	"plugin/internal/discord/webhook"
	"plugin/internal/logger"
	"plugin/internal/players"
	"plugin/internal/rcon"
	"plugin/internal/register"
	"plugin/internal/utils"
	"strings"
)

func registerOwnerCommands(
	cfg *config.Config,
	rc *rcon.RCON,
	reg *register.Register,
	players *players.Service,
	log *logger.Logger,
	webhook *webhook.Webhook,
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

			log.Infof("%s (%d) changed gambling status to %t\n", playerName, playerID, cfg.Gambling.Enabled)
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

			log.Infof("%s (%d) set max bet to %s%d\n", playerName, playerID, cfg.Gambling.Currency, amount)
			webhook.MaxBetWebhook(playerName, int(amount))
		},
	})

	// !addowner <xuid>
	// add a new owner
	reg.RegisterCommand(&register.Command{
		Name:     "addowner",
		Aliases:  aliases{"ao"},
		MinLevel: levelOwner,
		MinArgs:  1,
		Help:     "Usage: ^6!addowner^7 <xuid>",
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			target, err := players.GetByXUID(args[0])
			if err != nil {
				rc.Tell(clientNum, "Player couldnt be found ("+args[0]+")")
				return
			}

			if target.Level >= levelOwner {
				rc.Tell(clientNum, target.Name+" is already an owner")
				return
			}

			if err := players.UpdateLevel(target.ID, levelOwner); err != nil {
				log.Errorln(err)
				rc.Tell(clientNum, "Failed to update player level")
				return
			}

			log.Infof("%s (%d) promoted %s (%d) to owner\n", playerName, playerID, target.Name, target.ID)
			rc.Tell(clientNum, target.Name+" level set to owner")
		},
	})

	// !removeowner <xuid>
	// remove an owner
	reg.RegisterCommand(&register.Command{
		Name:     "removeowner",
		Aliases:  aliases{"ro"},
		MinLevel: levelOwner,
		MinArgs:  1,
		Help:     "Usage: ^6!removeowner^7 <xuid>",
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			target, err := players.GetByXUID(args[0])
			if err != nil {
				rc.Tell(clientNum, "Player couldnt be found ("+args[0]+")")
				return
			}

			if target.Level <= levelAdmin {
				rc.Tell(clientNum, target.Name+" is not an owner")
				return
			}

			if err := players.UpdateLevel(target.ID, levelUser); err != nil {
				log.Errorln(err)
				rc.Tell(clientNum, "Failed to update player level")
				return
			}

			log.Infof("%s (%d) demoted %s (%d) to user\n", playerName, playerID, target.Name, target.ID)
			rc.Tell(clientNum, target.Name+" level set to user")
		},
	})

	// !addadmin <xuid>
	// add a new admin
	reg.RegisterCommand(&register.Command{
		Name:     "addadmin",
		Aliases:  aliases{"aa"},
		MinLevel: levelOwner,
		MinArgs:  1,
		Help:     "Usage: ^6!addadmin^7 <xuid>",
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			target, err := players.GetByXUID(args[0])
			if err != nil {
				rc.Tell(clientNum, "Player couldnt be found ("+args[0]+")")
				return
			}

			if target.Level >= levelAdmin {
				rc.Tell(clientNum, target.Name+" is already an admin")
				return
			}

			if err := players.UpdateLevel(target.ID, levelAdmin); err != nil {
				log.Errorln(err)
				rc.Tell(clientNum, "Failed to update player level")
				return
			}

			log.Infof("%s (%d) promoted %s (%d) to admin\n", playerName, playerID, target.Name, target.ID)
			rc.Tell(clientNum, target.Name+" level set to admin")
		},
	})

	// !removeadmin <xuid>
	// remove an admin
	reg.RegisterCommand(&register.Command{
		Name:     "removeadmin",
		Aliases:  aliases{"ra"},
		MinLevel: levelOwner,
		MinArgs:  1,
		Help:     "Usage: ^6!removeadmin^7 <xuid>",
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			target, err := players.GetByXUID(args[0])
			if err != nil {
				rc.Tell(clientNum, "Player couldnt be found ("+args[0]+")")
				return
			}

			if target.Level <= levelAdmin {
				rc.Tell(clientNum, target.Name+" is not an admin")
				return
			}

			if err := players.UpdateLevel(target.ID, levelUser); err != nil {
				log.Errorln(err)
				rc.Tell(clientNum, "Failed to update player level")
				return
			}

			log.Infof("%s (%d) demoted %s (%d) to user\n", playerName, playerID, target.Name, target.ID)
			rc.Tell(clientNum, target.Name+" level set to user")
		},
	})
}
