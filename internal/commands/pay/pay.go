package pay

import (
	"errors"
	"fmt"
	"plugin/internal/config"
	"plugin/internal/discord/webhook"
	"plugin/internal/players"
	"plugin/internal/stats"
	"plugin/internal/wallet"
)

type Result struct {
	FromPlayer  int
	FromBalance int
	ToPlayer    int
	ToBalance   int

	Amount  int
	Message string
}

func Pay(
	fromPlayerID int,
	toPlayerID int,
	amount int,

	cfg *config.Config,
	player *players.Service,
	wallet *wallet.Service,
	walletStats *stats.WalletStatsService,

	webhook *webhook.Webhook,
) (*Result, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("invalid amount %d", amount)
	}

	from, err := player.GetByID(fromPlayerID)
	if err != nil {
		return nil, fmt.Errorf("error occurred, please try again later")
	}

	to, err := player.GetByID(toPlayerID)
	if err != nil {
		return nil, fmt.Errorf("receiver (%d) doesnt exists", toPlayerID)
	}

	fromBalance, err := wallet.GetBalance(fromPlayerID)
	if err != nil {
		return nil, err
	}

	if fromBalance < amount {
		return nil, errors.New("You dont have enough money")
	}

	if err := wallet.Withdraw(fromPlayerID, amount); err != nil {
		return nil, err
	}

	if err := wallet.Deposit(toPlayerID, amount); err != nil {
		_ = wallet.Deposit(fromPlayerID, amount)
		return nil, err
	}

	toBalance, err := wallet.GetBalance(toPlayerID)
	if err != nil {
		return nil, err
	}

	walletStats.Pay(fromPlayerID, amount)
	walletStats.Receive(toPlayerID, amount)

	if cfg.Discord.Enabled {
		webhook.PayWebhook(from.Name, to.Name, amount)
	}

	return &Result{
		FromPlayer:  fromPlayerID,
		ToPlayer:    toPlayerID,
		Amount:      amount,
		FromBalance: fromBalance - amount,
		ToBalance:   toBalance,
		Message: fmt.Sprintf(
			"You paid %s%d to player %d",
			cfg.Gambling.Currency,
			amount,
			toPlayerID,
		),
	}, nil
}
