package commands

import (
	"fmt"
	"plugin/internal/bank"
	"plugin/internal/config"
	"plugin/internal/discord/webhook"
	"plugin/internal/links"
	"plugin/internal/players"
	"plugin/internal/rcon"
	"plugin/internal/register"
	"plugin/internal/stats"
	"plugin/internal/wallet"
	"strings"
)

const (
	LevelUser = iota
	LevelAdmin
	LevelOwner
	LevelDeveloper
)

func RegisterCommands(
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
	registerOwnerCommands(cfg, rc, reg, players, wallet, bank)
	registerAdminCommands(cfg, rc, reg, players, wallet, bank)
	registerClientCommands(cfg, rc, reg, players, wallet, bank, links, playerStats, gambleStats, walletStats, webhook)
}

func resolveClientNum(rcon *rcon.RCON, reg *register.Register, query string) (int, error) {
	query = strings.TrimSpace(query)

	target := reg.FindPlayer(query)
	if target == nil {
		return -1, fmt.Errorf("Couldn't find %s", query)
	}

	cn := rcon.ClientNumByGUID(target.GUID)
	if cn == -1 {
		return -1, fmt.Errorf("Couldn't resolve %s", query)
	}

	return cn, nil
}

func resolveClientNums(
	rcon *rcon.RCON,
	reg *register.Register,
	clientNum uint8,
	args []string,
) (int, int, error) {

	if len(args) == 0 {
		return -1, -1, fmt.Errorf("at least one target is required")
	}

	// first target
	cn1 := int(clientNum)
	var err error

	if len(args) >= 1 {
		cn1, err = resolveClientNum(rcon, reg, args[0])
		if err != nil {
			return -1, -1, err
		}
	}

	// second target
	if len(args) == 1 {
		return int(clientNum), cn1, nil
	}

	cn2, err := resolveClientNum(rcon, reg, args[1])
	if err != nil {
		return -1, -1, err
	}

	if cn1 == cn2 {
		return -1, -1, fmt.Errorf("targets must be different players")
	}

	return cn1, cn2, nil
}
