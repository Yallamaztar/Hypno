package main

/*
 * hypno.go
 *
 * Entry point for the Hypno plugin.
 * This application acts as the central controller for all Hypno systems,
 * coordinating communication between game servers, external integrations,
 * and internal services.
 *
 * The plugin is designed to be modular and concurrent, with each server
 * running in its own goroutine while sharing common services.
 */

import (
	"fmt"
	"os"
	"plugin/internal/bank"
	"plugin/internal/commands"
	"plugin/internal/config"
	"plugin/internal/database"
	"plugin/internal/discord/bot"
	"plugin/internal/discord/webhook"
	"plugin/internal/events"
	"plugin/internal/iw4m"
	"plugin/internal/links"
	"plugin/internal/logger"
	"plugin/internal/players"
	"plugin/internal/rcon"
	"plugin/internal/register"
	"plugin/internal/stats"
	"plugin/internal/wallet"
	"sync"
)

func main() {
	log := logger.New("Hypno [MAIN]", "hypno_log.log")
	log.Infoln("Loading config.yaml")

	cfg, err := config.Setup(log)
	if err != nil {
		log.Errorf("Config setup failed: ", err)
		os.Exit(1)
	}

	log.Infoln("Sucessfully setup and loaded config")
	log.Infoln("Starting database migrations")

	db, err := database.Open()
	if err != nil {
		log.Errorf("Failed to open database: ", err)
		os.Exit(1)
	}

	// do database migrations
	if err := database.Migrate(db); err != nil {
		log.Errorf("Failed database migration")
		os.Exit(1)
	}

	// creating services
	ps := players.New(players.NewRepository(db))
	ws := wallet.New(wallet.NewRepository(db))
	bs := bank.New(bank.NewRepository(db))
	ls := links.New(links.NewRepository(db))

	pStats := stats.NewPlayerStats(stats.NewPlayerStatsRepository(db))
	gStats := stats.NewGamblingStats(stats.NewGamblingStatsRepository(db))
	wStats := stats.NewWalletStats(stats.NewWalletStatsRepository(db))
	log.Infoln("Database migrations done")

	var wh *webhook.Webhook
	// creating webhook and running bot if enabled
	if cfg.Discord.Enabled {
		println()
		log.Infoln("Starting discord integrations")

		if cfg.Discord.WebhookLink == "" {
			log.Errorln("Discord webhook link cant be empty if discord config is enabled")
			os.Exit(1)
		}
		wh = webhook.New(cfg.Discord.WebhookLink, cfg)

		if cfg.Discord.BotToken == "" {
			log.Errorln("Discord bot token cant be empty if discord config is enabled")
			os.Exit(1)
		}

		go func() {
			// Start the discord bot
			if err := bot.Run(cfg, ps, ws, bs, ls, pStats, gStats, wStats, wh); err != nil {
				log.Errorf("Failed to start the discord bot: %+v\n", err)
				os.Exit(1)
			}
		}()
		log.Infoln("Discord bot running")
	} else {
		log.Warnln("Skipping discord integrations (not enabled in config)")
	}

	var iw *iw4m.IW4MWrapper
	// creating IW4M-Admin wrapper if enabled
	if cfg.IW4MAdmin.Enabled {
		println()
		log.Infoln("Starting IW4M-Admin integrations")

		iw = iw4m.New(cfg, log)
		if err := iw.TestConnection(); err != nil {
			log.Errorln(err)
			os.Exit(1)
		}
		log.Infoln("Successfully connected to IW4M-Admin (" + iw.Host + ")")
	} else {
		log.Warnln("Skipping IW4M-Admin integration (not enabled in config)")
	}

	var wg sync.WaitGroup
	for i, s := range cfg.Server {
		println()
		serverLog := logger.New(cfg.Server[i].Host, fmt.Sprintf("hypno_%d_log.log", i))

		serverLog.Infoln("Starting RCON client")

		// creating server specific rcon connections
		rc, err := rcon.New(s.Host, s.Password, cfg, serverLog)
		if err != nil {
			serverLog.Errorf("Couldnt connect to RCON: %+v\n", err)
			continue
		}

		if err = rc.TestConnection(); err != nil {
			serverLog.Errorln(err)
			serverLog.Debugln("Make sure you have the necessary GSC scripts in your server /t6/scripts/")
			continue
		}
		serverLog.Infoln("Successfully connected to RCON & GSC")

		serverLog.Infoln("Registering client commands")
		reg := register.New(cfg, rc, ps, serverLog)
		commands.RegisterCommands(cfg, rc, reg, iw, ps, ws, bs, ls, pStats, gStats, wStats, serverLog, wh)

		wg.Add(1)
		go func(rc *rcon.RCON, slog *logger.Logger) {
			defer wg.Done()
			defer rc.Close()
			slog.Infoln("Hypno Plugin Starting!")
			events.RunLogTailer(i, cfg, rc, reg, iw, ps, ws, wStats, slog)
		}(rc, serverLog)
	}

	wg.Wait()
}
