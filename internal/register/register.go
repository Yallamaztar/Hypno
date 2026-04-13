package register

import (
	"fmt"
	"plugin/internal/config"
	"plugin/internal/logger"
	"plugin/internal/players"
	"plugin/internal/rcon"
	"strings"
	"sync"
	"time"
)

// Handler is a function that handles a registered command
type Handler func(
	clientNum uint8,
	playerID int,
	playerName string,
	xuid string,
	level int,
	args []string,
)

// Command represents a registered command
//   - Name: The name of the command
//   - Aliases: List of aliases for the command
//   - MinLevel: The minimum level required to use the command
//   - Help: The help text for the command
//   - MinArgs: The minimum number of arguments required for the command
//   - Handler: The function that handles the command
type Command struct {
	Name     string
	Aliases  []string
	MinLevel int
	Help     string
	MinArgs  uint8
	Handler  Handler
}

type commands map[string]*Command // map command names and/or aliases to Command struct
type clients map[string]uint8     // map player names to client numbers

// Register manages registered commands and client information
type Register struct {
	commands commands
	clients  clients

	rc      *rcon.RCON
	cfg     *config.Config
	players *players.Service

	log *logger.Logger
	mu  sync.RWMutex
}

// New creates a new Register instance
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

// RegisterCommand registers a new command
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

// Execute executes a registered command if found in Register.commands
func (r *Register) Execute(
	clientNum uint8,
	playerID int,
	playerName string,
	xuid string,
	level int,
	command string,
	args []string,
) {
	r.mu.RLock()
	cmd, ok := r.commands[strings.ToLower(command)]
	r.mu.RUnlock()

	if !ok {
		return
	}

	if cmd.Handler == nil {
		r.log.Warnln("Command handler is nil:", cmd.Name)
		return
	}

	if !r.hasPermission(level, cmd.MinLevel) {
		r.log.Infoln("Player does not have permission for command:", cmd.Name)
		r.rc.Tell(clientNum, fmt.Sprintf(
			"You ^1don't ^7have permission for !%s",
			cmd.Name,
		))
		return
	}

	if len(args) < int(cmd.MinArgs) {
		r.log.Infoln("Not enough arguments for command:", cmd.Name)
		if cmd.Help != "" {
			r.rc.Tell(clientNum, cmd.Help)
		}
		return
	}

	r.log.Infof("Executing command: %s for %s (%d)\n", command, playerName, playerID)
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

type PlayerInfo struct {
	Name      string
	GUID      string
	ClientNum int
}

// FindPlayerPartial finds a player by a partial name match
func (r *Register) FindPlayerPartial(partial string) *PlayerInfo {
	name := strings.ToLower(strings.TrimSpace(partial))

	var status *rcon.Status
	var err error

	for i := 0; i < 3; i++ {
		status, err = r.rc.Status()
		if err == nil && status != nil {
			break
		}

		r.log.Errorf("Failed to get status (%d/3)\n", i)
		time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
	}

	if status == nil {
		r.log.Errorln("Failed to get status after 3 attempts")
		return nil
	}

	for _, p := range status.Players {
		target := strings.ToLower(p.Name)

		// exact match
		if target == name {
			return &PlayerInfo{
				Name:      p.Name,
				GUID:      p.GUID,
				ClientNum: p.ClientNum,
			}
		}

		// partial match
		if strings.Contains(target, name) {
			return &PlayerInfo{
				Name:      p.Name,
				GUID:      p.GUID,
				ClientNum: p.ClientNum,
			}
		}
	}

	// fallback to db look up
	player, err := r.players.GetByPartial(name)
	if err != nil {
		r.log.Errorln("Error occurred while fetching player by partial name")
		return nil
	}

	cn := r.rc.ClientNumByGUID(player.GUID)
	if cn == -1 {
		r.log.Errorln("Error occurred while fetching client number by GUID")
		return nil
	}

	return &PlayerInfo{
		Name:      player.Name,
		GUID:      player.GUID,
		ClientNum: cn,
	}
}
