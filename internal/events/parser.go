package events

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	joinReg = regexp.MustCompile(
		`^(?P<cmd>J);(?P<xuid>-?[A-Fa-f0-9_]{1,32}|bot\d+|0);(?P<client>\d+);(?P<name>.+)$`,
	)

	quitReg = regexp.MustCompile(
		`^Q;(?P<xuid>-?\w+);(?P<client>\d+);(?P<name>.+)$`,
	)

	sayReg = regexp.MustCompile(
		`^(say|sayteam);(?P<xuid>-?\w+);(?P<client>\d+);(?P<name>[^;]+);(?P<msg>.+)$`,
	)

	weaponReg = regexp.MustCompile(
		`^Weapon;(?P<xuid>-?\w+);(?P<client>\d+);(?P<name>[^;]+);(?P<weapon>.+)$`,
	)

	killReg = regexp.MustCompile(
		`^K;(?P<vXuid>-?\w+);(?P<vClient>\d+);(?P<vTeam>\w+);(?P<vName>[^;]+);` +
			`(?P<aXuid>-?\w+);(?P<aClient>\d+);(?P<aTeam>\w+);(?P<aName>[^;]+);` +
			`(?P<weapon>[^;]+);(?P<damage>\d+);(?P<mod>\w+);(?P<hit>.+)$`,
	)
)

func parseEventLine(line string) (event, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, fmt.Errorf("Empty line")
	}

	raw := line
	var ts *time.Time

	if i := strings.IndexByte(line, ' '); i > 0 && strings.Contains(line[:i], ":") {
		if dur, err := parseTimestamp(line[:i]); err == nil {
			t := time.Now().Truncate(time.Second).Add(dur)
			ts = &t
			line = line[i+1:]
		}
	}

	if strings.HasPrefix(line, "say") {
		return parseSayEvent(line, raw, ts)
	}

	switch {
	case strings.HasPrefix(line, "InitGame:"):
		data := parseKeyValuePairs(strings.TrimPrefix(line, "InitGame:"))
		return &serverEvent{
			baseEvent: baseEvent{
				Command:   roundStartCommand,
				Raw:       raw,
				Timestamp: ts,
			},
			event: &data,
		}, nil

	case strings.HasPrefix(line, "ShutdownGame:"):
		return &serverEvent{
			baseEvent: baseEvent{
				Command:   roundEndCommand,
				Raw:       raw,
				Timestamp: ts,
			},
			event: nil,
		}, nil
	}

	if strings.Contains(line, ";") {
		if ev, err := parseJoinEvent(line, raw, ts); err == nil {
			return ev, nil
		}

		if ev, err := parseQuitEvent(line, raw, ts); err == nil {
			return ev, nil
		}
		if ev, err := parseKillEvent(line, raw, ts); err == nil {
			return ev, nil
		}
		if ev, err := parseWeaponEvent(line, raw, ts); err == nil {
			return ev, nil
		}
	}

	return &baseEvent{
		Command:   unknownCommand,
		Raw:       raw,
		Timestamp: ts,
	}, nil
}

// helper function to atoi client num
func atoi(s string) (int, error) {
	i, err := strconv.Atoi(s)
	return i, err
}

// parseJoinEvent
// parses a join event from the log file
//
//	example of a join event:
//	  J;XUID;CLIENTNUM;NAME
func parseJoinEvent(line, raw string, ts *time.Time) (*playerEvent, error) {
	s := joinReg.FindStringSubmatch(line)
	if s == nil {
		return nil, fmt.Errorf("Not a join event")
	}

	cn, err := atoi(s[3])
	if err != nil {
		return nil, fmt.Errorf("Couldnt parse join event: invalid client number %q", s[3])
	}

	return &playerEvent{
		baseEvent: baseEvent{
			Command:   joinCommand,
			Raw:       raw,
			Timestamp: ts,
		},

		name: s[4],
		xuid: s[2],
		cn:   cn,
	}, nil
}

// parseQuitEvent
// parses a quit event from the log file
//
//	example of a quit event:
//	  Q;XUID;CLIENTNUM;NAME
func parseQuitEvent(line, raw string, ts *time.Time) (*playerEvent, error) {
	s := quitReg.FindStringSubmatch(line)
	if s == nil {
		return nil, fmt.Errorf("Not a quit event")
	}

	cn, err := atoi(s[3])
	if err != nil {
		return nil, fmt.Errorf("Couldnt parse join event: invalid client number %q", s[3])
	}

	return &playerEvent{
		baseEvent: baseEvent{
			Command:   quitCommand,
			Raw:       raw,
			Timestamp: ts,
		},

		name: s[4],
		xuid: s[2],
		cn:   cn,
	}, nil
}

// parseSayEvent
// parses a chat event (say or sayteam) from the log file
//
//	example of a say event:
//	  say;XUID;CLIENTNUM;NAME;MESSAGE
//	  sayteam;XUID;CLIENTNUM;NAME;MESSAGE
func parseSayEvent(line, raw string, ts *time.Time) (*playerEvent, error) {
	s := sayReg.FindStringSubmatch(line)
	if s == nil {
		return nil, fmt.Errorf("Not a say event")
	}

	cn, err := atoi(s[3])
	if err != nil {
		return nil, fmt.Errorf("Couldnt parse say event: invalid client number %q", s[3])
	}

	return &playerEvent{
		baseEvent: baseEvent{
			Command:   sayCommand,
			Raw:       raw,
			Timestamp: ts,
		},

		name:    s[4],
		xuid:    s[2],
		cn:      cn,
		message: s[5],
	}, nil
}

// parseWeaponEvent
// parses a weapon event from the log file
//
//	 example of a weapon event:
//		Weapon;XUID;CLIENTNUM;NAME;WEAPON+EXTENSIONS
func parseWeaponEvent(line, raw string, ts *time.Time) (*playerEvent, error) {
	s := weaponReg.FindStringSubmatch(line)
	if s == nil {
		return nil, fmt.Errorf("Not a weapon event")
	}

	cn, err := atoi(s[2])
	if err != nil {
		return nil, fmt.Errorf("Couldnt parse weapon event: invalid client number %q", s[2])
	}

	return &playerEvent{
		baseEvent: baseEvent{
			Command:   weaponCommand,
			Raw:       raw,
			Timestamp: ts,
		},
		xuid:    s[1],
		cn:      cn,
		name:    s[3],
		message: s[4],
	}, nil
}

// parseKillEvent
// parses a kill event from the log file
//
//	example of a kill event:
//	  K;VICTIM_XUID;VICTIM_CLIENTNUM;VICTIM_TEAM;VICTIM_NAME;ATTACKER_XUID;ATTACKER_CLIENTNUM;ATTACKER_TEAM;ATTACKER_NAME;WEAPON;DAMAGE;MOD;HITLOCATION
func parseKillEvent(line, raw string, ts *time.Time) (*killEvent, error) {
	s := killReg.FindStringSubmatch(line)
	if s == nil {
		return nil, fmt.Errorf("Not a kill event")
	}

	vcn, err := atoi(s[2])
	if err != nil {
		return nil, fmt.Errorf("Couldnt parse kill event: invalid client number %q (victim)", s[2])
	}

	acn, err := atoi(s[6])
	if err != nil {
		return nil, fmt.Errorf("Couldnt parse kill event: invalid client number %q (attacker)", s[6])
	}

	return &killEvent{
		baseEvent: baseEvent{
			Command:   killCommand,
			Raw:       raw,
			Timestamp: ts,
		},
		victimXUID: s[1],
		victimCN:   vcn,
		victimTeam: s[3],
		victimName: s[4],

		attackerXUID: s[5],
		attackerCN:   acn,
		attackerTeam: s[7],
		attackerName: s[8],

		weapon:       s[9],
		damage:       s[10],
		meansOfDeath: s[11],
		hitLocation:  s[12],
	}, nil
}

// parseTimestamp
// parses a timestamp string into a time.Duration
//
//	example of a timestamp:
//	  MM:SS
func parseTimestamp(s string) (time.Duration, error) {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return 0, errors.New("invalid timestamp format")
	}

	minutes, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, errors.New("invalid timestamp format (minutes)")
	}

	seconds, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, errors.New("invalid timestamp format (seconds)")
	}

	return time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second, nil
}

// parseKeyValuePairs
// parses a backslash-delimited key-value string into a map
//
//	example of a key-value string:
//	  \key1\value1\key2\value2
func parseKeyValuePairs(s string) map[string]string {
	m := make(map[string]string)
	pairs := strings.Split(s, "\\")
	for i := 1; i < len(pairs); i += 2 {
		if i+1 < len(pairs) {
			m[pairs[i]] = pairs[i+1]
		}
	}
	return m
}
