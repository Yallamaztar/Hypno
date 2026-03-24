package queries

const (
	CreateBank       = `INSERT INTO bank (balance, created_at) VALUES (?, CURRENT_TIMESTAMP);`
	GetBankBalance   = `SELECT balance FROM bank LIMIT 1;`
	SetBankBalance   = `UPDATE bank SET balance = ?;`
	DepositToBank    = `UPDATE bank SET balance = balance + ?;`
	WithdrawFromBank = `UPDATE bank SET balance = balance - ?;`
)
