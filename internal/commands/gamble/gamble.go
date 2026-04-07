package gamble

import (
	"fmt"
	"math"
	"math/rand"
	"plugin/internal/bank"
	"plugin/internal/config"
	"plugin/internal/discord/webhook"
	"plugin/internal/players"
	"plugin/internal/stats"
	"plugin/internal/wallet"
	"time"
)

type Result struct {
	Won     bool
	Amount  int
	Balance int
}

func Gamble(
	playerID int,
	playerName string,
	amount int,

	cfg *config.Config,
	players *players.Service,
	wallet *wallet.Service,
	bank *bank.Service,

	playerStats *stats.PlayerStatsService,
	gambleStats *stats.GamblingStatsService,
	walletStats *stats.WalletStatsService,

	webhook *webhook.Webhook,
) (*Result, error) {
	if !cfg.Gambling.Enabled {
		return nil, fmt.Errorf("gambling is disabled")
	}

	if amount <= 0 {
		return nil, fmt.Errorf("Invalid amount %d", amount)
	}

	balance, err := wallet.Balance(playerID)
	if err != nil {
		return nil, err
	}

	if balance < amount {
		return nil, fmt.Errorf("You ^6dont ^7have enough money (missing %s%d)", cfg.Gambling.Currency, (amount - balance))
	}

	if cfg.Economy.MaxBet != 0 && amount > cfg.Economy.MaxBet {
		return nil, fmt.Errorf("Amount exceeds maximum bet limit of %s%d", cfg.Gambling.Currency, cfg.Economy.MaxBet)
	}

	if didWin(cfg.Gambling.WinChance) {
		// withdraw winnings from bank
		if err := bank.Withdraw(amount); err != nil {
			return nil, err
		}

		// deposit winnings to players wallet
		if err := wallet.Deposit(playerID, amount); err != nil {
			_ = bank.Deposit(amount)
			return nil, err
		}

		// update stats
		if err := playerStats.Win(playerID, amount, amount*2); err != nil {
			return nil, err
		}

		if err := gambleStats.RecordGamble(amount, amount*2); err != nil {
			return nil, err
		}

		if err := walletStats.Deposit(playerID, amount); err != nil {
			return nil, err
		}

		if cfg.Discord.Enabled {
			webhook.WinWebhook(playerName, amount)
		}

		return &Result{
			Won:     true,
			Amount:  amount,
			Balance: balance + amount,
		}, nil
	}

	// loss
	if err := wallet.Withdraw(playerID, amount); err != nil {
		return nil, err
	}

	if err := bank.Deposit(amount); err != nil {
		return nil, err
	}

	if err := playerStats.Loss(playerID, amount); err != nil {
		return nil, err
	}

	if err := gambleStats.RecordGamble(amount, 0); err != nil {
		return nil, err
	}

	if err := walletStats.Withdraw(playerID, amount); err != nil {
		return nil, err
	}

	if cfg.Discord.Enabled {
		webhook.LossWebhook(playerName, amount)
	}

	return &Result{
		Won:     false,
		Amount:  amount,
		Balance: balance - amount,
	}, nil
}

// scizophrenic maniac paranoid level of randomness
var rng = rand.New(rand.NewSource(time.Now().UnixNano()))
var paranoia uint64 = uint64(time.Now().UnixNano())

func xorshift64(x uint64) uint64 {
	x ^= x << 13
	x ^= x >> 7
	x ^= x << 17
	return x
}

func mix64(x uint64) uint64 {
	x ^= x >> 30
	x *= 0xbf58476d1ce4e5b9
	x ^= x >> 27
	x *= 0x94d049bb133111eb
	x ^= x >> 31
	return x
}

func paranoidFloat() float64 {
	r := rng.Uint64()
	t := uint64(time.Now().UnixNano())
	paranoia = xorshift64(paranoia + t + r)
	mixed := mix64(r ^ paranoia ^ t)
	return float64(mixed>>11) * (1.0 / (1 << 53))
}

func didWin(winChance float64) bool {
	v := paranoidFloat()
	v = math.Mod(math.Sin(v*1e6+math.Phi)*1e5, 1.0)
	if v < 0 {
		v += 1
	}

	return v < winChance
}
