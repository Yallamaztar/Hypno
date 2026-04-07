package events

import (
	"fmt"
	"plugin/internal/config"
	"plugin/internal/iw4m"
	"plugin/internal/logger"
	"plugin/internal/players"
	"plugin/internal/rcon"
	"plugin/internal/register"
	ss "plugin/internal/stats"
	"plugin/internal/utils"
	"plugin/internal/wallet"
	"strings"
)

func RunLogTailer(
	index int,

	cfg *config.Config,
	rc *rcon.RCON,
	reg *register.Register,
	iw4m *iw4m.IW4MWrapper,

	players *players.Service,
	wallet *wallet.Service,
	walletStats *ss.WalletStatsService,

	log *logger.Logger,
) {
	eventsCh := make(chan event, 128)
	go func() {
		if err := tail(log, cfg.Server[index].LogPath, eventsCh); err != nil {
			log.Fatalf("Failed to tail file: %s: %w", cfg.Server[index].LogPath, err)
		}
		close(eventsCh)
	}()

	for e := range eventsCh {
		switch event := e.(type) {
		case *playerEvent:
			switch event.Command {
			case joinCommand:
				if err := ensurePlayer(event.xuid, uint8(event.cn), event.name, reg, rc, cfg, iw4m, players, wallet, walletStats, log); err != nil {
					log.Errorln("Failed to ensure player: " + err.Error())
					continue
				}

				p, err := players.GetByXUID(event.xuid)
				if err != nil {
					log.Errorln("Failed to get player")
					continue
				}

				if cfg.IW4MAdmin.Enabled {
					if p.IW4MID == nil {
						log.Warnf("IW4MID is nil for %s", p.Name)
						continue
					}
					stats, err := iw4m.Stats(*p.IW4MID, index)
					if err != nil {
						log.Errorln("Failed to get IW4M-Admin stats")
						continue
					}

					reward := utils.CalcJoinReward(stats.TotalSecondsPlayed, stats.Kills, stats.Deaths, cfg.Economy.JoinReward)
					wallet.Deposit(p.ID, reward)
					walletStats.Deposit(p.ID, reward)
					rc.Tell(uint8(event.cn), fmt.Sprintf("^7Spawning bonus: %s%d", cfg.Gambling.Currency, reward))
				} else {
					wallet.Deposit(p.ID, cfg.Economy.JoinReward)
					walletStats.Deposit(p.ID, cfg.Economy.JoinReward)
					rc.Tell(uint8(event.cn), fmt.Sprintf("^7Spawning bonus: %s%d", cfg.Gambling.Currency, cfg.Economy.JoinReward))
				}

				log.Infoln("Successfully processed join event for " + event.xuid)

			case quitCommand:
				if err := ensurePlayer(event.xuid, uint8(event.cn), event.name, reg, rc, cfg, iw4m, players, wallet, walletStats, log); err != nil {
					log.Errorln("Failed to ensure player: " + err.Error())
					continue
				}

				reg.RemoveClientNum(event.xuid)
				statsMu.Lock()
				delete(stats, event.xuid)
				statsMu.Unlock()

			case sayCommand:
				if cmd, ok := strings.CutPrefix(event.message, cfg.Server[index].CommandPrefix); ok {
					if err := ensurePlayer(event.xuid, uint8(event.cn), event.name, reg, rc, cfg, iw4m, players, wallet, walletStats, log); err != nil {
						log.Errorln("Failed to ensure player: " + err.Error())
						continue
					}

					parts := strings.Fields(cmd)
					if len(parts) > 0 {
						args := []string{}
						if len(parts) > 1 {
							args = parts[1:]
						}

						p, err := players.GetByXUID(event.xuid)
						if err != nil {
							continue
						}

						if p == nil {
							continue
						}

						if err := ensurePlayer(event.xuid, uint8(event.cn), event.name, reg, rc, cfg, iw4m, players, wallet, walletStats, log); err != nil {
							log.Errorln("Failed to ensure player: " + err.Error())
							continue
						}

						reg.Execute(uint8(event.cn), p.ID, event.name, event.xuid, p.Level, parts[0], args)
					}
				}
			}

		case *killEvent:
			if event.attackerXUID == event.victimXUID {
				continue
			}

			if err := ensurePlayer(event.attackerXUID, uint8(event.attackerCN), event.attackerName, reg, rc, cfg, iw4m, players, wallet, walletStats, log); err != nil {
				log.Errorln("Failed to ensure attacker: " + err.Error())
				continue
			}

			if err := ensurePlayer(event.victimXUID, uint8(event.victimCN), event.victimName, reg, rc, cfg, iw4m, players, wallet, walletStats, log); err != nil {
				log.Errorln("Failed to ensure victim: " + err.Error())
				continue
			}

			go func() {
				// attacker stats
				if asess := getOrCreateSession(event.attackerXUID); asess != nil {
					asess.mu.Lock()
					asess.kills++
					asess.mu.Unlock()
				}

				// victim stats
				if vsess := getOrCreateSession(event.victimXUID); vsess != nil {
					vsess.mu.Lock()
					vsess.kills++
					vsess.mu.Unlock()
				}

				attacker, err := players.GetByXUID(event.attackerXUID)
				if err != nil {
					log.Warnf("Failed to get attacker %s: %v", event.attackerXUID, err)
					return
				}

				victim, err := players.GetByXUID(event.victimXUID)
				if err != nil {
					log.Warnf("Failed to get victim %s: %v", event.victimXUID, err)
					return
				}

				if err := wallet.Deposit(attacker.ID, cfg.Economy.KillReward); err != nil {
					log.Warnf("Failed to deposit kill reward for %s: %v", event.attackerXUID, err)
					return
				}

				if err := wallet.Withdraw(victim.ID, cfg.Economy.DeathPenalty); err != nil {
					log.Warnf("Failed to withdraw death penalty for %s: %v", event.victimXUID, err)
					return
				}

				if mvp, ok := mostvaluable[event.victimXUID]; ok {
					rc.Tell(uint8(event.attackerCN), fmt.Sprintf(
						"Claimed bounty: %s%d for killing %s",
						cfg.Gambling.Currency, mvp, event.victimName,
					))

					rc.Say(fmt.Sprintf(
						"%s has killed MVP %s and got rewarded: %s%d",
						event.attackerName, event.victimName, cfg.Gambling.Currency, mvp,
					))

					delete(mostvaluable, event.victimXUID)
				}
			}()

		case *serverEvent:
			if event.Command == roundStartCommand {
				status, err := rc.Status()
				if err != nil {
					continue
				}

				if len(status.Players) < 4 {
					continue
				}

				go func() {
					var (
						topKills int
						topXUID  string
						topName  string
					)

					for _, ps := range status.Players {
						p, err := players.GetByGUID(ps.GUID)
						if err != nil {
							continue
						}

						s, ok := stats[p.XUID]
						if !ok {
							continue
						}

						s.mu.Lock()
						if s.kills > topKills {
							topKills = s.kills
							topXUID = p.XUID
							topName = p.Name
						}
						s.mu.Unlock()

						if topKills < 7 {
							continue
						}

						reward := utils.RandomReward()
						mostvaluable[topXUID] = reward

						rc.Say(fmt.Sprintf(
							"^6BOUNTY ACTIVE! ^7Kill %s to ^6claim ^7the bounty!",
							topName,
						))
					}
				}()
			}

			if event.Command == roundEndCommand {
				status, err := rc.Status()
				if err != nil {
					continue
				}

				if len(status.Players) < 4 {
					continue
				}

				go func() {
					for _, ps := range status.Players {
						p, err := players.GetByGUID(ps.GUID)
						if err != nil {
							continue
						}

						s, ok := mostvaluable[p.XUID]
						if !ok {
							continue
						}

						cn := rc.ClientNumByGUID(ps.GUID)
						if cn == -1 {
							break
						}

						reward := int(s / 5)

						wallet.Deposit(p.ID, reward)
						rc.Tell(uint8(cn), fmt.Sprintf("You survived the round as MVP and got rewarded: %s%d", cfg.Gambling.Currency, reward))
					}
				}()
			}
		}
	}
}

func ensurePlayer(
	xuid string,
	clientNum uint8,
	name string,
	reg *register.Register,
	rc *rcon.RCON,
	cfg *config.Config,
	iw4m *iw4m.IW4MWrapper,
	players *players.Service,
	wallet *wallet.Service,
	walletStats *ss.WalletStatsService,
	log *logger.Logger,
) error {
	reg.SetClientNum(xuid, uint8(clientNum))
	exists, err := players.ExistsByXUID(xuid)
	if err != nil {
		log.Warnln("Failed to check if player exists")
		return err
	}

	if exists {
		return nil
	}

	var guid string
	for i := 0; i < 5; i++ {
		guid = rc.GUIDByClientNum(uint8(clientNum))
		if guid == "" {
			continue
		} else {
			break
		}
	}

	if guid == "" {
		log.Errorln("Failed to get GUID for player")
		return fmt.Errorf("failed to get GUID for player")
	}

	id, err := players.Create(name, xuid, guid, 0, cfg, iw4m)
	if err != nil {
		log.Errorln("Failed to create player profile " + err.Error())
		return err
	}

	if err := wallet.Create(id, cfg.Economy.FirstTimeReward); err != nil {
		log.Errorln("Failed to create wallet")
		return err
	}

	if err := walletStats.Init(id); err != nil {
		log.Errorln("walletStats.Init failed:", err)
	}
	if err := walletStats.Deposit(id, cfg.Economy.FirstTimeReward); err != nil {
		log.Errorln("walletStats.Deposit failed:", err)
	}

	return rc.Tell(uint8(clientNum),
		fmt.Sprintf(
			"^6Created ^7a wallet with ^6%s%d ^7balance",
			cfg.Gambling.Currency,
			cfg.Economy.FirstTimeReward,
		),
	)
}
