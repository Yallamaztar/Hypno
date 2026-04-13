package commands

import (
	"fmt"
	"plugin/internal/bank"
	"plugin/internal/commands/gamble"
	"plugin/internal/commands/pay"
	"plugin/internal/config"
	"plugin/internal/discord/webhook"
	"plugin/internal/iw4m"
	"plugin/internal/links"
	"plugin/internal/logger"
	"plugin/internal/players"
	"plugin/internal/rcon"
	"plugin/internal/register"
	"plugin/internal/stats"
	"plugin/internal/utils"
	"plugin/internal/wallet"
	"strconv"
	"strings"
	"time"
)

func registerClientCommands(
	cfg *config.Config,
	rc *rcon.RCON,
	reg *register.Register,
	iw4m *iw4m.IW4MWrapper,

	players *players.Service,
	wallet *wallet.Service,
	bank *bank.Service,
	links *links.Service,

	playerStats *stats.PlayerStatsService,
	gambleStats *stats.GamblingStatsService,
	walletStats *stats.WalletStatsService,

	log *logger.Logger,
	webhook *webhook.Webhook,
) {
	// !level (!lvl)
	// check your level
	reg.RegisterCommand(&register.Command{
		Name:     "level",
		Aliases:  aliases{"lvl"},
		MinLevel: levelUser,
		MinArgs:  0,
		Help:     "Usage: ^6!level",
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			player, err := players.GetByXUID(xuid)
			if err != nil {
				rc.Tell(clientNum, "^1Error ^7getting player info")
				return
			}

			if player == nil {
				rc.Tell(clientNum, "^1Error ^7player not found")
				return
			}

			rc.Tell(clientNum, fmt.Sprintf("You are level ^6%s", LevelToString(player.Level)))
		},
	})

	// !claimrole (!claim) <role>
	// claim the owner | developer role ingame (one time use)
	reg.RegisterCommand(&register.Command{
		Name:     "claimrole",
		Aliases:  aliases{"claim"},
		MinLevel: levelUser,
		MinArgs:  1,
		Help:     "Usage: ^6!claimrole ^7<role>",
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			role := strings.ToLower(strings.TrimSpace(args[0]))

			var targetLevel int

			switch role {
			case "owner":
				targetLevel = levelOwner

				existsOwner, err := players.ExistsByLevel(levelOwner)
				if err != nil {
					log.Errorln("[ClaimRole] %s (%d): Error checking owner status\n", playerName, playerID)
					rc.Tell(clientNum, "^1Error ^7checking owner status")
					return
				}

				if existsOwner {
					log.Infoln("[ClaimRole] %s (%d): Tried to claim owner role\n", playerName, playerID)
					rc.Tell(clientNum, "^6Owner ^7role already claimed")
					return
				}

				existsDeveloper, err := players.ExistsByLevel(levelDeveloper)
				if err != nil {
					log.Errorln("[ClaimRole] %s (%d): Error checking developer status\n", playerName, playerID)
					rc.Tell(clientNum, "^1Error ^7checking developer status")
					return
				}

				if existsDeveloper {
					log.Infoln("[ClaimRole] %s (%d): Tried to claim owner role", playerName, playerID)
					rc.Tell(clientNum, "^6Developer ^7role already claimed")
					return
				}

				if err := players.UpdateLevel(playerID, levelOwner); err != nil {
					log.Errorln("[ClaimRole] %s (%d): Error updating player level\n", playerName, playerID)
					rc.Tell(clientNum, "^1Error ^7updating player level")
					return
				}

			case "developer":
				targetLevel = levelDeveloper

				exists, err := players.ExistsByLevel(levelDeveloper)
				if err != nil {
					log.Errorln("[ClaimRole] %s (%d): Error checking developer status\n", playerName, playerID)
					rc.Tell(clientNum, "^1Error ^7checking developer status")
					return
				}
				if exists {
					log.Infoln("[ClaimRole] %s (%d): Tried to claim developer role\n", playerName, playerID)
					rc.Tell(clientNum, "^6Developer ^7role already claimed")
					return
				}

				if err := players.UpdateLevel(playerID, levelDeveloper); err != nil {
					log.Errorln("[ClaimRole] %s (%d): Error updating player level\n", playerName, playerID)
					rc.Tell(clientNum, "^1Error ^7updating player level")
					return
				}

			default:
				log.Infoln("[ClaimRole] %s (%d): Invalid role requested %s\n", playerName, playerID, role)
				rc.Tell(clientNum, "^1Invalid ^7role ("+role+")")
				return
			}

			log.Infof("[ClaimRole] %s (%d): claimed the %s role\n", playerName, playerID, LevelToString(targetLevel))
			rc.Tell(clientNum, fmt.Sprintf("You claimed the ^6%s^7 role", LevelToString(targetLevel)))
		},
	})

	// !link (!lnk)
	// link your ingame to a discord account
	reg.RegisterCommand(&register.Command{
		Name:     "link",
		Aliases:  aliases{"lnk", "linkdc"},
		MinLevel: levelUser,
		MinArgs:  0,
		Help:     "Usage: ^6!link <reset (optional)>",
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			discordID, err := players.GetDiscordIDByID(playerID)
			if err != nil {
				log.Errorf("[Link] %s (%d): Error checking %+v\n", playerName, playerID, err)
				rc.Tell(clientNum, "^1Error ^7checking your account, try again later")
				return
			}

			if len(args) >= 1 && args[0] == "reset" {
				if discordID == "0" {
					rc.Tell(clientNum, "You have ^6not ^7linked your account")
					return
				}
			}

			if discordID != "0" {
				rc.Tell(clientNum, "You have ^6already linked ^7your account")
				return
			}

			code := utils.GenerateCode()
			if code == "" {
				log.Errorf("[Link] %s (%d): Failed to generate link code\n", playerName, playerID)
				rc.Tell(clientNum, "^1Failed ^7to generate link code, ^6try again ^7later")
				return
			}

			if err = links.Create(playerID, code); err != nil {
				log.Errorf("[Link] %s (%d): Error saving link code to database: %+v\n", playerName, playerID, err)
				rc.Tell(clientNum, "Failed to save link code to the database")
				return
			}

			log.Infof("[Link] %s (%d): generated a new link code (%s)\n", playerName, playerID, code)
			rc.Tell(clientNum, fmt.Sprintf("Your code is: ^6%s", code))
			rc.Tell(clientNum, "use ^6/link <code> ^7in discord to link your account")
		},
	})

	// !banflip (!bf) <duration> <amount>
	// like gamble but more stakes and more fun naturally
	reg.RegisterCommand(&register.Command{
		Name:     "banflip",
		Aliases:  aliases{"bf"},
		MinLevel: levelUser,
		Help:     "Usage: ^6!banflip ^7<duration> <amount>",
		MinArgs:  2,
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			if !cfg.IW4MAdmin.Enabled {
				rc.Tell(clientNum, "^6BANFLIP ^7is not enabled")
				return
			}

			balance, err := wallet.Balance(playerID)
			if err != nil {
				rc.Tell(clientNum, "Couldnt get your balance")
				return
			}

			amount, err := utils.ParseAmountArg(args[1], balance)
			if err != nil {
				rc.Tell(clientNum, fmt.Sprintf("%s (%q)", err, args[1]))
				return
			}

			multiplier, err := utils.ParseDurationMultiplier(args[0])
			if err != nil {
				rc.Tell(clientNum, fmt.Sprintf("Invalid duration: %s", args[0]))
				return
			}

			rc.Say(fmt.Sprintf("Starting ^6BANFLIP^7 for %s with multiplier ^6x%d", playerName, multiplier))
			res, err := gamble.Gamble(playerID, playerName, (amount * multiplier), cfg, players, wallet, bank, playerStats, gambleStats, walletStats, webhook)
			if err != nil {
				rc.Say("it failed mb men")
				return
			}

			if res.Won {
				rc.Tell(clientNum, fmt.Sprintf("You just ^6won %s%d! WTF FUCK", cfg.Gambling.Currency, res.Amount))
				rc.Say(fmt.Sprintf("WTF FUCK %s just ^2WON ^7%s%d!", playerName, cfg.Gambling.Currency, res.Amount))
			} else {
				rc.Tell(clientNum, fmt.Sprintf("You just ^1lost^7 %s%d!", cfg.Gambling.Currency, res.Amount))
				rc.Say(fmt.Sprintf("LOL %s just ^1LOST ^7the BANFLIP!", playerName))
				rc.Say("^1^Fget fucked kid")

				iw4m.TempBan(
					*iw4m.ClientIDFromGUID(rc.GUIDByClientNum(clientNum)),
					args[0],
					"I Lost Banflip with a potential multiplier of "+
						strconv.Itoa(multiplier),
				)
			}

			log.Infoln("[BANFLIP] %s (%d): %t | multiplier: %d | duration: %s", playerName, playerID, map[bool]string{true: "won", false: "lost"}[res.Won], multiplier, args[0])
		},
	})

	// !gamble (!g) <amount>
	// gamble money like a boss
	reg.RegisterCommand(&register.Command{
		Name:     "gamble",
		Aliases:  aliases{"g", "cf", "coinflip"},
		MinLevel: levelUser,
		Help:     "Usage: ^6!gamble ^7<amount>",
		MinArgs:  1,
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			balance, err := wallet.Balance(playerID)
			if err != nil {
				rc.Tell(clientNum, "Couldnt get your balance")
				return
			}

			amount, err := utils.ParseAmountArg(args[0], balance)
			if err != nil {
				rc.Tell(clientNum, fmt.Sprintf("%s (%q)", err, args[0]))
				return
			}

			if amount == 0 {
				rc.Say(fmt.Sprintf("%s is ^1^Fgay n poor", playerName))
				return
			}

			res, err := gamble.Gamble(playerID, playerName, amount, cfg, players, wallet, bank, playerStats, gambleStats, walletStats, webhook)
			if err != nil {
				rc.Tell(clientNum, err.Error())
				return
			}

			if res.Won {
				rc.Tell(clientNum, fmt.Sprintf("You just ^6won^7 %s%d!", cfg.Gambling.Currency, amount))
				time.Sleep(100 * time.Millisecond)
				rc.Say(fmt.Sprintf("%s just ^6won ^7%s%d!", playerName, cfg.Gambling.Currency, res.Amount))
			} else {
				rc.Tell(clientNum, fmt.Sprintf("You just ^1lost^7 %s%d!", cfg.Gambling.Currency, amount))
				time.Sleep(100 * time.Millisecond)
				rc.Say(fmt.Sprintf("%s just ^6lost ^7%s%d!", playerName, cfg.Gambling.Currency, res.Amount))
			}

			log.Infof("[GAMBLE] %s (%d): %s | amount: %s%d\n", playerName, playerID, map[bool]string{true: "won", false: "lost"}[res.Won], cfg.Gambling.Currency, res.Amount)
		},
	})

	// !pay (!pp) <player> <amount>
	// pay a player money
	reg.RegisterCommand(&register.Command{
		Name:     "pay",
		Aliases:  aliases{"pp", "payplayer", "transfer"},
		MinLevel: levelUser,
		Help:     "Usage: ^6!pay <player> <amount>",
		MinArgs:  2,
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			t := reg.FindPlayerPartial(args[0])
			if t == nil {
				rc.Tell(clientNum, fmt.Sprintf("player ^6%s ^7couldnt be found", args[0]))
				return
			}

			balance, err := wallet.Balance(playerID)
			if err != nil {
				rc.Tell(clientNum, "Couldnt get your balance")
				return
			}

			amount, err := utils.ParseAmountArg(args[1], balance)
			if err != nil {
				rc.Tell(clientNum, fmt.Sprintf("%s ^6(%q)", err, args[1]))
				return
			}

			target, err := players.GetByGUID(t.GUID)
			if err != nil {
				rc.Tell(clientNum, t.Name+" doesnt exists")
				return
			}

			res, err := pay.Pay(playerID, target.ID, amount, cfg, players, wallet, walletStats, webhook)
			if err != nil {
				rc.Tell(clientNum, err.Error())
				return
			}

			rc.Tell(clientNum, res.Message)
			log.Infof("[PAY] %s (%d): paid %s (%d) %s%d\n", playerName, playerID, target.Name, target.ID, cfg.Gambling.Currency, res.Amount)
		},
	})

	// !balance (!bal) <player (optional)>
	// check your or another players balance
	reg.RegisterCommand(&register.Command{
		Name:     "balance",
		Aliases:  aliases{"bal", "money", "wallet"},
		MinLevel: levelUser,
		Help:     "Usage: ^6!balance <player>",
		MinArgs:  0,
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			if len(args) == 0 {
				balance, err := wallet.Balance(playerID)
				if err != nil {
					rc.Tell(clientNum, "^1Error ^7getting your balance")
					return
				}

				rc.Tell(clientNum, fmt.Sprintf("Your balance: ^6%s%d", cfg.Gambling.Currency, balance))
				return
			}

			t := reg.FindPlayerPartial(args[0])
			if t == nil {
				rc.Tell(clientNum, fmt.Sprintf("player ^6%s ^7couldnt be found", args[0]))
				return
			}

			target, err := players.GetByGUID(t.GUID)
			if err != nil {
				rc.Tell(clientNum, t.Name+" doesnt exists")
				return
			}

			balance, err := wallet.Balance(target.ID)
			if err != nil {
				rc.Tell(clientNum, "^1Error ^7getting player's balance")
				return
			}

			rc.Tell(clientNum, fmt.Sprintf("^6%s's ^7balance: ^6%s%d", t.Name, cfg.Gambling.Currency, balance))
			log.Infof("[BALANCE] %s (%d): checked %s's balance - %s%d\n", playerName, playerID, t.Name, cfg.Gambling.Currency, balance)
		},
	})

	// !bankbalance (!bank)
	// check banks balance
	reg.RegisterCommand(&register.Command{
		Name:     "bankbalance",
		Aliases:  aliases{"bb", "bank", "bankbal"},
		MinLevel: levelUser,
		Help:     "Usage: ^6!bankbalance",
		MinArgs:  0,
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			bal, err := bank.Balance()
			if err != nil {
				rc.Tell(clientNum, "Couldnt get the banks balance")
				return
			}

			rc.Tell(clientNum, fmt.Sprintf("Bank balance is ^6%s%d", cfg.Gambling.Currency, bal))
			log.Infof("[BANKBALANCE] %s (%d): Checked bank balance %s%d\n", playerName, playerID, cfg.Gambling.Currency, bal)
		},
	})

	// !richest (!rich)
	// lists top 5 richest players
	reg.RegisterCommand(&register.Command{
		Name:     "richest",
		Aliases:  aliases{"rich"},
		MinLevel: levelUser,
		Help:     "Usage: ^6!richest",
		MinArgs:  0,
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			wallets, err := wallet.GetTop5RichestWallets()
			if err != nil {
				rc.Tell(clientNum, "Couldnt get wallets")
				return
			}

			log.Infoln("[RICHEST] top 5 Richest Players:")
			for i, w := range wallets {
				log.Infof("  [%d] %s %s%s\n", i+1, w.Name, cfg.Gambling.Currency, utils.FormatMoney(w.Balance))
				rc.Tell(clientNum, fmt.Sprintf("[%d] %s %s%s", i+1, w.Name, cfg.Gambling.Currency, utils.FormatMoney(w.Balance)))
			}
		},
	})

	// !poorest (!poor)
	// lists top 5 poorest players
	reg.RegisterCommand(&register.Command{
		Name:     "poorest",
		Aliases:  aliases{"poor"},
		MinLevel: levelUser,
		Help:     "Usage: ^6!poorest",
		MinArgs:  0,
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			wallets, err := wallet.GetTop5PoorestWallets()
			if err != nil {
				rc.Tell(clientNum, "Couldnt get wallets")
				return
			}

			log.Infoln("[POOREST] top 5 Poorest Players:")
			for i, w := range wallets {
				log.Infof("  [%d] %s %s%s\n", i+1, w.Name, cfg.Gambling.Currency, utils.FormatMoney(w.Balance))
				rc.Tell(clientNum, fmt.Sprintf("[%d] %s %s%s", i+1, w.Name, cfg.Gambling.Currency, utils.FormatMoney(w.Balance)))
			}
		},
	})

	// !discord (!dc)
	// show the discord invite link (if discord enabled)
	reg.RegisterCommand(&register.Command{
		Name:     "discord",
		Aliases:  aliases{"dc", "disc"},
		MinLevel: levelUser,
		Help:     "Usage: ^6!discord",
		MinArgs:  1,
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			if cfg.Discord.Enabled && cfg.Discord.InviteLink != "" {
				rc.Tell(clientNum, "^6"+cfg.Discord.InviteLink)
			}
		},
	})
}
