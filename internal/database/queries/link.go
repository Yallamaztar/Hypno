package queries

const (
	CreateLink         = `INSERT INTO discord_links (player_id, code, expires_at) VALUES (?, ?, ?);`
	DeleteLinkByPlayer = `DELETE FROM discord_links WHERE player_id = ?;`
	DeleteLinkByCode   = `DELETE FROM discord_links WHERE code = ?;`
	GetIDByCode        = `SELECT player_id FROM discord_links WHERE code = ? AND expires_at > NOW();`
	GetCodeByID        = `SELECT code FROM discord_links WHERE player_id = ? AND expires_at > NOW();`
)
