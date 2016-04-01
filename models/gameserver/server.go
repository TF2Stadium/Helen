package gameserver

type ServerRecord struct {
	ID             uint
	Host           string
	LogSecret      string
	ServerPassword string // sv_password
	RconPassword   string // rcon_password
}
