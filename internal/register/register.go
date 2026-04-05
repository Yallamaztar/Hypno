package register

import (
	"fmt"
	"plugin/internal/config"
	"plugin/internal/logger"
	"plugin/internal/players"
	"plugin/internal/rcon"
	"strings"
	"sync"
)

type Handler func(
	clientNum uint8,
	playerID int,
	playerName string,
	xuid string,
	level int,
	args []string,
)

type Command struct {
	Name     string
	Aliases  []string
	MinLevel int
	Help     string
	MinArgs  uint8
	Handler  Handler
}

type commands map[string]*Command
type clients map[string]uint8

type Register struct {
	commands commands
	clients  clients

	rc      *rcon.RCON
	cfg     *config.Config
	players *players.Service

	log *logger.Logger

	mu sync.RWMutex
}

func New(cfg *config.Config, rc *rcon.RCON, players *players.Service, log *logger.Logger) *Register {
	return &Register{
		commands: make(commands),
		clients:  make(clients),

		rc:      rc,
		cfg:     cfg,
		players: players,
		log:     log,

		mu: sync.RWMutex{},
	}
}
func (r *Register) RegisterCommand(cmd *Command) {
	if cmd.Handler == nil {
		panic("command " + cmd.Name + " registered with nil handler")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	c := cmd
	r.commands[strings.ToLower(c.Name)] = c
	for _, alias := range c.Aliases {
		r.commands[strings.ToLower(alias)] = c
	}

	r.log.Printf("Successfully registered command %s %v (level: %d)", c.Name, c.Aliases, c.MinLevel)
}

func (r *Register) Execute(
	clientNum uint8,
	playerID int,
	playerName string,
	xuid string,
	level int,
	command string,
	args []string,
) {
	r.log.Infoln("Executing command:", command)
	r.mu.RLock()
	cmd, ok := r.commands[strings.ToLower(command)]
	r.mu.RUnlock()

	if !ok {
		r.log.Infoln("Command not found:", command)
		return
	}

	if cmd.Handler == nil {
		r.log.Infoln("Command handler is nil:", cmd.Name)
		return
	}

	if !r.hasPermission(level, cmd.MinLevel) {
		r.log.Infoln("Player does not have permission for command:", cmd.Name)
		r.tell(clientNum, fmt.Sprintf(
			"You ^1don't ^7have permission for !%s",
			cmd.Name,
		))
		return
	}

	if len(args) < int(cmd.MinArgs) {
		r.log.Infoln("Not enough arguments for command:", cmd.Name)
		if cmd.Help != "" {
			r.tell(clientNum, cmd.Help)
		}
		return
	}

	go cmd.Handler(clientNum, playerID, playerName, xuid, level, args)
}

func (r *Register) SetClientNum(xuid string, clientNum uint8) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clients[xuid] = clientNum
}

func (r *Register) GetClientNum(xuid string) (uint8, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cn, ok := r.clients[xuid]
	return cn, ok
}

func (r *Register) RemoveClientNum(xuid string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.clients, xuid)
}

func (r *Register) hasPermission(level, required int) bool {
	return level >= required
}

func (r *Register) tell(clientNum uint8, msg string) {
	if r.rc != nil {
		r.rc.Tell(clientNum, msg)
	}
}

type PlayerInfo struct {
	Name      string
	GUID      *string
	ClientNum *int
}

func (r *Register) FindPlayerPartial(partial string) *PlayerInfo {
	name := strings.ToLower(strings.TrimSpace(partial))

	status, err := r.rc.Status()
	if err != nil {
		return nil
	}

	for _, p := range status.Players {
		target := strings.ToLower(p.Name)

		if target == name {
			return &PlayerInfo{
				Name:      p.Name,
				GUID:      &p.GUID,
				ClientNum: &p.ClientNum,
			}
		}

		if strings.Contains(target, name) {
			return &PlayerInfo{
				Name:      p.Name,
				GUID:      &p.GUID,
				ClientNum: &p.ClientNum,
			}
		}

		p, err := r.players.GetByGUID(p.GUID)
		if err == nil {
			cn := r.rc.ClientNumByGUID(p.GUID)
			if cn == -1 {
				return nil
			}

			return &PlayerInfo{
				Name:      p.Name,
				GUID:      &p.GUID,
				ClientNum: &cn,
			}
		}
	}
	return nil
}
