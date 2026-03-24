package commands

import (
	"fmt"
	"plugin/internal/bank"
	"plugin/internal/commands/gamble"
	"plugin/internal/config"
	"plugin/internal/discord/webhook"
	"plugin/internal/links"
	"plugin/internal/players"
	"plugin/internal/rcon"
	"plugin/internal/register"
	"plugin/internal/stats"
	"plugin/internal/utils"
	"plugin/internal/wallet"
)

func registerClientCommands(
	cfg *config.Config,
	rc *rcon.RCON,
	reg *register.Register,

	players *players.Service,
	wallet *wallet.Service,
	bank *bank.Service,
	links *links.Service,

	playerStats *stats.PlayerStatsService,
	gambleStats *stats.GamblingStatsService,
	walletStats *stats.WalletStatsService,

	webhook *webhook.Webhook,
) {
	// !link (!lnk)
	// link your ingame to a discord account
	reg.RegisterCommand(register.Command{
		Name:     "link",
		Aliases:  []string{"lnk", "linkdc"},
		MinLevel: LevelUser,
		MinArgs:  0,
		Help:     "Usage: ^6!link",
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			discordID, err := players.GetDiscordIDByID(playerID)
			if err != nil {
				rc.Tell(clientNum, "^1Error ^7checking your account, try again later")
				return
			}

			if discordID != "" {
				rc.Tell(clientNum, "You have ^6already linked ^7your account")
				return
			}

			code := utils.GenerateCode()
			if code == "" {
				rc.Tell(clientNum, "^1Failed ^7to generate link code, ^6try again ^7later")
				return
			}

			if err = links.Create(playerID, code); err != nil {
				rc.Tell(clientNum, "Failed to save link code to the database")
				return
			}

			rc.Tell(clientNum, fmt.Sprintf("Your code is: ^6%s", code))
			rc.Tell(clientNum, "use ^6/link <code> ^7in discord to link your account")
		},
	})

	// !gamble (!g) <amount>
	// gamble money like a boss
	reg.RegisterCommand(register.Command{
		Name:     "gamble",
		Aliases:  []string{"g", "cf", "coinflip"},
		MinLevel: LevelUser,
		Help:     "Usage: ^6!gamble ^7<amount>",
		MinArgs:  1,
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			balance, err := wallet.GetBalance(playerID)
			if err != nil {
				rc.Tell(clientNum, "Couldnt get your balance")
				return
			}

			amount, err := utils.ParseAmountArg(args[0], balance)
			if err != nil {
				rc.Tell(clientNum, fmt.Sprintf("%s (%q)", err, args[0]))
				return
			}

			res, err := gamble.Gamble(playerID, playerName, amount, cfg, players, wallet, bank, playerStats, gambleStats, walletStats, webhook)
			if err != nil {
				rc.Tell(clientNum, err.Error())
				return
			}

			rc.Tell(clientNum, res.Message)
			if res.Won {
				rc.Say(fmt.Sprintf("%s just ^6won ^7%s%d!", playerName, cfg.Gambling.Currency, res.Amount))
			} else {
				rc.Say(fmt.Sprintf("%s just ^6lost ^7%s%d!", playerName, cfg.Gambling.Currency, res.Amount))
			}
		},
	})

	// !pay (!pp) <player> <amount>
	// pay a player money
	reg.RegisterCommand(register.Command{
		Name:     "pay",
		Aliases:  []string{"pp", "payplayer", "transfer"},
		MinLevel: LevelUser,
		Help:     "Usage: ^6!pay <player> <amount>",
		MinArgs:  2,
		Handler: func(clientNum uint8, playerID int, playerName, xuid string, level int, args []string) {
			balance, err := wallet.GetBalance(playerID)
			if err != nil {
				rc.Tell(clientNum, "Couldnt get your balance")
				return
			}

			amount, err := utils.ParseAmountArg(args[0], balance)
			if err != nil {
				rc.Tell(clientNum, fmt.Sprintf("%s ^6(%q)", err, args[0]))
				return
			}

			t := reg.FindPlayer(args[0])
			if t == nil {
				rc.Tell(clientNum, fmt.Sprintf("player ^6%s ^7couldnt be found", args[0]))
				return
			}

			target, err := players.GetByGUID(t.GUID)
			if err != nil {
				rc.Tell(clientNum, t.Name+" doesnt exists")
				return
			}

		},
	})

	// !balance (!bal) <player (optional)>
	// check your or another players balance
	reg.RegisterCommand(register.Command{})

	// !bankbalance (!bank)
	// check banks balance
	reg.RegisterCommand(register.Command{})

	// !richest (!rich)
	// lists top 5 richest players
	reg.RegisterCommand(register.Command{})

	// !poorest (!poor)
	// lists top 5 poorest players
	reg.RegisterCommand(register.Command{})
}
