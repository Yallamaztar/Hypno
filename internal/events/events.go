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
	"time"
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
		if err := tail(cfg.Server[index].LogPath, eventsCh); err != nil {
			log.Fatalf("Failed to tail file: %s: %w", cfg.Server[index].LogPath, err)
		}
		close(eventsCh)
	}()

	for e := range eventsCh {
		if !cfg.Gambling.Enabled {
			time.Sleep(250 * time.Millisecond)
			continue
		}

		switch event := e.(type) {
		case *playerEvent:
			switch event.Command {
			case joinCommand:
				go func() {
					reg.SetClientNum(event.xuid, uint8(event.cn))

					guid := rc.GUIDByClientNum(uint8(event.cn))
					if guid == "" {
						return
					}

					exists, err := players.ExistsByXUID(event.xuid)
					if err != nil {
						return
					}

					if !exists {
						id, err := players.Create(event.name, event.xuid, guid, 0, cfg, iw4m)
						if err != nil {
							return
						}

						if err := wallet.Create(id, cfg.Economy.FirstTimeReward); err != nil {
							return
						}

						walletStats.Init(id)
						walletStats.Deposit(id, cfg.Economy.FirstTimeReward)

						log.Printf("Created wallet for %s (%s) | ID: %d\n", event.name, event.xuid, id)

						rc.Tell(uint8(event.cn),
							fmt.Sprintf(
								"^7Created a wallet with ^6%s%d balance",
								cfg.Gambling.Currency,
								cfg.Economy.FirstTimeReward,
							),
						)
					}

					p, err := players.GetByXUID(event.xuid)
					if err != nil {
						return
					}

					if cfg.IW4MAdmin.Enabled {
						stats, err := iw4m.Stats(*p.IW4MID, index)
						if err != nil {
							return
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
				}()

			case quitCommand:
				go func() {
					reg.RemoveClientNum(event.xuid)
					delete(stats, event.xuid)
				}()

			case sayCommand:
				if cmd, ok := strings.CutPrefix(event.message, cfg.Server[index].CommandPrefix); ok {
					parts := strings.Fields(cmd)
					if len(parts) > 0 {
						args := []string{}
						if len(parts) > 1 {
							args = parts[1:]
						}

						p, err := players.GetByXUID(event.xuid)
						if err != nil {
							return
						}

						if p == nil {
							return
						}

						go reg.Execute(uint8(event.cn), p.ID, event.name, event.xuid, p.Level, parts[0], args)
					}
				}
			}

		case *killEvent:
			// on suicide just ignore
			if event.attackerXUID == event.victimXUID {
				return
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
					return
				}

				if len(status.Players) < 4 {
					return
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
							return
						}

						s, ok := stats[p.XUID]
						if !ok {
							return
						}

						s.mu.Lock()
						if s.kills > topKills {
							topKills = s.kills
							topXUID = p.XUID
							topName = p.Name
						}
						s.mu.Unlock()

						if topKills < 7 {
							return
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
					return
				}

				if len(status.Players) < 4 {
					return
				}

				go func() {
					for _, ps := range status.Players {
						p, err := players.GetByGUID(ps.GUID)
						if err != nil {
							return
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
