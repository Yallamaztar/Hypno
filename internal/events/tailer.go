package events

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type eventType int

const (
	unknownCommand eventType = iota

	joinCommand
	quitCommand

	sayCommand
	killCommand
	weaponCommand

	roundStartCommand
	roundEndCommand
)

type event interface {
	Type() eventType
	String() string
	Time() *time.Time
}

type baseEvent struct {
	Command   eventType
	Raw       string
	Timestamp *time.Time
}

func (b *baseEvent) Type() eventType  { return b.Command }
func (b *baseEvent) String() string   { return b.Raw }
func (b *baseEvent) Time() *time.Time { return b.Timestamp }

type playerEvent struct {
	baseEvent

	name string
	xuid string
	cn   int

	message string
}

type serverEvent struct {
	baseEvent
	event *map[string]string // round end has no event data
}

type killEvent struct {
	baseEvent

	attackerName string
	attackerXUID string
	attackerCN   int
	attackerTeam string

	victimName string
	victimXUID string
	victimCN   int
	victimTeam string

	weapon       string
	damage       string
	meansOfDeath string
	hitLocation  string
}

func tail(path string, eventsCh chan<- event) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("Failed to open log file %s: %w", path, err)
	}

	defer file.Close()

	if _, err := file.Seek(0, io.SeekEnd); err != nil {
		return fmt.Errorf("Failed to seek log file %s: %w", path, err)
	}

	reader := bufio.NewReader(file)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			return err
		}

		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			continue
		}

		ev, err := parseEventLine(line)
		if err != nil {
			fmt.Printf("events: failed to parse line: %v\n", err)
			continue
		}

		// non blocking channel
		select {
		case eventsCh <- ev:
		default:
		}
	}
}
