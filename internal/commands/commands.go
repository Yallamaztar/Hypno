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

	"github.com/google/shlex"
)

const (
	levelUser = iota
	levelAdmin
	levelOwner
	levelDeveloper
)

type aliases []string

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
	registerDeveloperCommands(cfg, rc, reg, players, wallet, bank)
	registerOwnerCommands(cfg, rc, reg, players)
	registerAdminCommands(cfg, rc, reg, players, wallet, bank)
	registerClientCommands(cfg, rc, reg, players, wallet, bank, links, playerStats, gambleStats, walletStats, webhook)
}

// resolveClientNum finds the client number of given target in args, if none found: return origins clientNum
func resolveClientNum(rcon *rcon.RCON, reg *register.Register, clientNum uint8, args []string) (int, error) {
	if len(args) == 0 {
		return int(clientNum), nil
	}

	query := strings.TrimSpace(strings.Join(args, " "))
	target := reg.FindPlayerPartial(query)
	if target == nil {
		return -1, fmt.Errorf("^6%s ^7couldnt be found", query)
	}

	cn := rcon.ClientNumByGUID(*target.GUID)
	if cn == -1 {
		return -1, fmt.Errorf("^6%s ^7couldnt be found ingame", target.Name)
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
		return -1, -1, fmt.Errorf("At least one target is required")
	}

	resolve := func(query string) (int, error) {
		target := reg.FindPlayerPartial(query)
		if target == nil {
			return -1, fmt.Errorf("Coudlnt find ^6%s", query)
		}

		cn := rcon.ClientNumByGUID(*target.GUID)
		if cn == -1 {
			return -1, fmt.Errorf("^6%s ^7couldnt be found ingame", target.Name)
		}

		return cn, nil
	}

	if len(args) == 1 {
		cn2, err := resolve(strings.TrimSpace(args[0]))
		if err != nil {
			return -1, -1, err
		}
		return int(clientNum), cn2, nil
	}

	cn1, err := resolve(strings.TrimSpace(args[0]))
	if err != nil {
		return -1, -1, err
	}

	cn2, err := resolve(strings.TrimSpace(args[1]))
	if err != nil {
		return -1, -1, err
	}

	if cn1 == cn2 {
		return -1, -1, fmt.Errorf("Targets must be different players")
	}

	return cn1, cn2, nil
}

// parseArgs uses googles shlex string split method instead of strings package,
// since some players names include spaces (e.g. F A Y A Z) and this seems to work
// with it, in previous versions atleast
func parseArgs(args []string) ([]string, error) {
	parsed, err := shlex.Split(strings.Join(args, " "))
	if err != nil {
		return nil, fmt.Errorf("Invalid arguments")
	}

	return parsed, nil
}
