package bot

import (
	"fmt"
	"plugin/internal/bank"
	"plugin/internal/commands/gamble"
	"plugin/internal/commands/pay"
	"plugin/internal/config"
	"plugin/internal/discord/webhook"
	"plugin/internal/links"
	"plugin/internal/players"
	"plugin/internal/stats"
	"plugin/internal/utils"
	"plugin/internal/wallet"

	"github.com/bwmarrin/discordgo"
)

func Run(
	cfg *config.Config,
	ps *players.Service,
	ws *wallet.Service,
	bs *bank.Service,
	ls *links.Service,
	playerStats *stats.PlayerStatsService,
	gambleStats *stats.GamblingStatsService,
	walletStats *stats.WalletStatsService,
	webhook *webhook.Webhook,
) error {
	dg, err := discordgo.New("Bot " + cfg.Discord.BotToken)
	if err != nil {
		return err
	}

	dg.AddHandler(func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		handleInteraction(cfg, ps, ws, bs, ls, playerStats, gambleStats, walletStats, webhook, session, interaction)
	})

	if err := dg.Open(); err != nil {
		return err
	}

	commands := createCommands(cfg)
	for _, cmd := range commands {
		if _, err := dg.ApplicationCommandCreate(dg.State.User.ID, "", cmd); err != nil {
			return err
		}
	}

	return nil
}

func createCommands(cfg *config.Config) []*discordgo.ApplicationCommand {
	commands := []*discordgo.ApplicationCommand{
		// /gamble <amount>
		{
			Name:        "gamble",
			Description: cfg.Discord.GamblingDescription,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "amount",
					Description: "Amount to gamble",
					Required:    true,
				},
			},
		},

		// /pay <player> <amount>
		{
			Name:        "pay",
			Description: cfg.Discord.PayDescription,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "user",
					Description: "User to pay money to",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "amount",
					Description: "Amount to pay",
					Required:    true,
				},
			},
		},

		// /balance <player (optional)>
		{
			Name:        "balance",
			Description: cfg.Discord.BalanceDescription,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "player",
					Description: "(Optional) View another players balance",
					Required:    false,
				},
			},
		},

		// /link <code>
		{
			Name:        "link",
			Description: cfg.Discord.LinkDescription,
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "code",
					Description: "The code you got ingame",
					Required:    true,
				},
			},
		},
	}

	return commands
}

func handleInteraction(
	cfg *config.Config,
	ps *players.Service,
	ws *wallet.Service,
	bs *bank.Service,
	ls *links.Service,
	playerStats *stats.PlayerStatsService,
	gambleStats *stats.GamblingStatsService,
	walletStats *stats.WalletStatsService,
	webhook *webhook.Webhook,

	session *discordgo.Session,
	interaction *discordgo.InteractionCreate,
) {
	if interaction.Type != discordgo.InteractionApplicationCommand {
		return
	}

	data := interaction.ApplicationCommandData()
	switch data.Name {
	case "gamble":
		handleGamble(cfg, ps, ws, bs, playerStats, gambleStats, walletStats, webhook, session, interaction, data)

	case "pay":
		handlePay(cfg, ps, ws, walletStats, webhook, session, interaction, data)

	case "balance":
		handleBalance(cfg, ps, ws, session, interaction, data)

	case "link":
		handleLink(ps, ls, session, interaction, data)
	}
}

func handleGamble(
	cfg *config.Config,
	ps *players.Service,
	ws *wallet.Service,
	bs *bank.Service,
	playerStats *stats.PlayerStatsService,
	gambleStats *stats.GamblingStatsService,
	walletStats *stats.WalletStatsService,
	webhook *webhook.Webhook,

	session *discordgo.Session,
	interaction *discordgo.InteractionCreate,
	data discordgo.ApplicationCommandInteractionData,
) {
	session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})

	var _amount string
	for _, opt := range data.Options {
		if opt.Name == "amount" {
			_amount = opt.StringValue()
		}
	}

	player, err := ps.GetByDiscordID(getIDSafe(interaction))
	if err != nil {
		session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: strPtr("You are not linked to an ingame account"),
		})
		return
	}

	bal, err := ws.GetBalance(player.ID)
	if err != nil {
		session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: strPtr("Couldnt get your wallet balance"),
		})
		return
	}

	amount, err := utils.ParseAmountArg(_amount, bal)
	if err != nil {
		session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: strPtr(err.Error()),
		})
		return
	}

	res, err := gamble.Gamble(player.ID, player.Name, amount, cfg, ps, ws, bs, playerStats, gambleStats, walletStats, webhook)
	if err != nil {
		session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: strPtr("Something went wrong while gambling"),
		})
		return
	}

	session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
		Content: strPtr("🎰 " + res.Message),
	})
}

func handlePay(
	cfg *config.Config,
	ps *players.Service,
	ws *wallet.Service,
	walletStats *stats.WalletStatsService,
	webhook *webhook.Webhook,

	session *discordgo.Session,
	interaction *discordgo.InteractionCreate,
	data discordgo.ApplicationCommandInteractionData,
) {
	session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})

	sender, err := ps.GetByDiscordID(getIDSafe(interaction))
	if err != nil {
		session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: strPtr("You are not linked to an ingame account"),
		})
		return
	}

	bal, err := ws.GetBalance(sender.ID)
	if err != nil {
		session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: strPtr("Couldnt get your wallet balance"),
		})
		return
	}

	var _target *discordgo.User
	var _amount string

	for _, opt := range data.Options {
		switch opt.Name {
		case "user":
			_target = opt.UserValue(session)

		case "amount":
			_amount = opt.StringValue()
		}
	}

	if _target == nil {
		session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: strPtr("User not found"),
		})
		return
	}

	target, err := ps.GetByDiscordID(_target.ID)
	if err != nil {
		session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: strPtr("User not found"),
		})
		return
	}

	amount, err := utils.ParseAmountArg(_amount, bal)
	if err != nil {
		session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: strPtr(err.Error()),
		})
		return
	}

	if amount <= 0 {
		session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: strPtr("Invalid amount"),
		})
		return
	}

	res, err := pay.Pay(sender.ID, target.ID, amount, cfg, ps, ws, walletStats, webhook)
	if err != nil {
		session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: strPtr("Something went wrong while trying to pay " + _target.GlobalName),
		})
		return
	}

	session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
		Content: strPtr("💳 " + res.Message),
	})
}

func handleBalance(
	cfg *config.Config,
	ps *players.Service,
	ws *wallet.Service,

	session *discordgo.Session,
	interaction *discordgo.InteractionCreate,
	data discordgo.ApplicationCommandInteractionData,
) {
	session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})

	var _target string
	_target = getIDSafe(interaction)

	for _, opt := range data.Options {
		if opt.Name == "player" {
			u := opt.UserValue(session)
			if u != nil {
				_target = u.ID
			}
		}
	}

	player, err := ps.GetByDiscordID(_target)
	if err != nil {
		session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: strPtr("That user is not linked to an ingame account"),
		})
		return
	}

	bal, err := ws.GetBalance(player.ID)
	if err != nil {
		session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: strPtr("Could not retrieve balance"),
		})
		return
	}

	var message string
	if bal == 0 {
		if _target == getIDSafe(interaction) {
			message = "You are gay n poor"
		} else {
			message = player.Name + " is gay n poor"
		}
	} else {
		if _target == getIDSafe(interaction) {
			message = "Your balance is"
		} else {
			message = player.Name + "'s balance is"
		}
	}

	session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
		Content: strPtr(fmt.Sprintf("💰 %s: %s%s", message, cfg.Gambling.Currency, utils.FormatMoney(bal))),
	})
}

func handleLink(
	ps *players.Service,
	ls *links.Service,

	session *discordgo.Session,
	interaction *discordgo.InteractionCreate,
	data discordgo.ApplicationCommandInteractionData,
) {
	session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})

	discordID := getIDSafe(interaction)

	var code string
	for _, opt := range data.Options {
		if opt.Name == "code" {
			code = opt.StringValue()
		}
	}

	if code == "" {
		session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: strPtr("Invalid code"),
		})
		return
	}

	_, err := ps.GetByDiscordID(discordID)
	if err == nil {
		session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: strPtr("Your account is already linked"),
		})
		return
	}

	playerID, err := ls.GetIDByCode(code)
	if err != nil {
		session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: strPtr("Invalid or expired code"),
		})
		return
	}

	player, err := ps.GetByID(playerID)
	if err != nil {
		session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: strPtr("Player not found"),
		})
		return
	}

	if err := ps.UpdateDiscordID(player.ID, discordID); err != nil {
		session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
			Content: strPtr("Failed to link account"),
		})
		return
	}

	_ = ls.DeleteByCode(code)
	_ = ls.DeleteByID(player.ID)

	session.InteractionResponseEdit(interaction.Interaction, &discordgo.WebhookEdit{
		Content: strPtr("✅ Successfully linked to **" + player.Name + "**"),
	})
}

// Helper functions

func getIDSafe(i *discordgo.InteractionCreate) string {
	if i.Member != nil && i.Member.User != nil {
		return i.Member.User.ID
	}
	if i.User != nil {
		return i.User.ID
	}
	return ""
}

func strPtr(s string) *string {
	return &s
}
