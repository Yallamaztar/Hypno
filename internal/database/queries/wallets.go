package queries

const (
	CreateWallet = `INSERT INTO wallets (player_id, balance, created_at) VALUES (?, ?, CURRENT_TIMESTAMP);`
	DeleteWallet = `DELETE FROM wallets WHERE id = ?;`
	WalletExists = `SELECT COUNT(1) FROM wallets WHERE player_id = ?;`

	GetWalletByID       = `SELECT id, player_id, balance, created_at FROM wallets where id = ?;`
	GetWalletByPlayerID = `SELECT id, player_id, balance, created_at FROM wallets WHERE player_id = ?;`

	GetWalletBalanceByPlayerID = `SELECT balance FROM wallets WHERE player_id = ?;`

	SetWalletBalance   = `UPDATE wallets SET balance = ? WHERE player_id = ?;`
	DepositToWallet    = `UPDATE wallets SET balance = balance + ? WHERE player_id = ?;`
	WithdrawFromWallet = `UPDATE wallets SET balance = balance - ? WHERE player_id = ?;`

	GetTopWallets    = `SELECT w.player_id, w.balance, p.name FROM wallets w JOIN players p ON p.id = w.player_id ORDER BY w.balance DESC LIMIT ?;`
	GetBottomWallets = `SELECT w.player_id, w.balance, p.name FROM wallets w JOIN players p ON p.id = w.player_id ORDER BY w.balance ASC LIMIT ?;`
)
