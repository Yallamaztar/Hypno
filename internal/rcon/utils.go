package rcon

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func normalizeRCON(s string) string {
	if s == "" {
		return s
	}
	s = strings.ReplaceAll(s, "\r\n", "\n")
	for {
		changed := false
		if strings.HasPrefix(s, "\xFF\xFF\xFF\xFF") {
			s = s[4:]
			changed = true
		}
		if strings.HasPrefix(s, "print\n") {
			s = s[6:]
			changed = true
		}
		if !changed {
			break
		}
	}

	s = strings.ReplaceAll(s, "\n\xFF\xFF\xFF\xFF", "\n")
	s = strings.ReplaceAll(s, "\nprint\n", "\n")
	return strings.TrimSpace(s)
}

func splitNonEmptyLines(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, "\n")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func stripColorCodes(s string) string {
	if s == "" {
		return s
	}
	// Remove ^<code> where code is an alphanumeric (covers ^0-^9 and potential ^a-^z variants)
	re := regexp.MustCompile(`\^[0-9A-Za-z]`)
	return re.ReplaceAllString(s, "")
}

func populateServerInfo(info *GetInfo, kvPairs map[string]string) {
	info.NetFieldChk = parseIntFromKVSafe(kvPairs, "netfieldchk")
	info.Protocol = int(parseIntFromKVSafe(kvPairs, "protocol"))
	info.SessionMode = int(parseIntFromKVSafe(kvPairs, "sessionmode"))
	info.Hostname = kvPairs["hostname"]
	info.MapName = kvPairs["mapname"]
	info.IsInGame = parseBoolSafe(kvPairs, "isInGame")
	info.MaxClients = int(parseIntFromKVSafe(kvPairs, "com_maxclients"))
	info.GameType = kvPairs["gametype"]
	info.HW = int(parseIntFromKVSafe(kvPairs, "hw"))
	info.Mod = parseBoolSafe(kvPairs, "mod")
	info.Voice = parseBoolSafe(kvPairs, "voice")
	info.SecKey = kvPairs["seckey"]
	info.SecID = kvPairs["secid"]
	info.HostAddr = kvPairs["hostaddr"]
}

func parseIntFromKVSafe(kvPairs map[string]string, key string) int {
	val, ok := kvPairs[key]
	if !ok {
		return 0
	}
	n, _ := strconv.ParseInt(val, 10, 64)
	return int(n)
}

func parseBoolSafe(kvPairs map[string]string, key string) bool {
	val, ok := kvPairs[key]
	if !ok {
		return false
	}
	b, _ := strconv.ParseBool(val)
	return b
}

func parseKeyValueResponse(lines []string) map[string]string {
	var dataLine string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.EqualFold(trimmed, infoResponseMarker) {
			continue
		}
		if strings.HasPrefix(trimmed, kvDelimiter) || strings.Contains(trimmed, kvDelimiter) {
			dataLine += trimmed
		}
	}

	if dataLine == "" && len(lines) > 0 {
		dataLine = strings.TrimSpace(lines[len(lines)-1])
	}

	parts := strings.Split(dataLine, kvDelimiter)
	if len(parts) > 0 && parts[0] == "" {
		parts = parts[1:]
	}

	kvPairs := make(map[string]string, len(parts)/2)
	for i := 0; i < len(parts)-1; i += 2 {
		key := strings.TrimSpace(parts[i])
		if key == "" {
			continue
		}
		val := strings.TrimSpace(parts[i+1])
		kvPairs[key] = stripColorCodes(val)
	}

	return kvPairs
}

func extractMapName(lines []string) string {
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(line), "map:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
			break
		}
	}
	return ""
}

func parsePlayerList(lines []string) ([]Player, error) {
	headerIdx := findPlayerHeaderIndex(lines)
	if headerIdx < 0 || headerIdx >= len(lines) {
		return nil, nil
	}

	pattern := regexp.MustCompile(playerLinePattern)
	var players []Player

	for _, line := range lines[headerIdx:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Skip lines that don't start with a digit
		if len(line) == 0 || line[0] < '0' || line[0] > '9' {
			continue
		}

		match := pattern.FindStringSubmatch(line)
		if match == nil {
			continue
		}

		player := parsePlayerLine(pattern, match)
		players = append(players, player)
	}

	return players, nil
}

func findPlayerHeaderIndex(lines []string) int {
	headerRegex := regexp.MustCompile(statusHeaderPattern)
	for i, line := range lines {
		if headerRegex.MatchString(strings.TrimSpace(line)) {
			return i + 1
		}
	}
	return 0
}

func parsePlayerLine(regex *regexp.Regexp, match []string) Player {
	group := func(name string) string {
		for i, n := range regex.SubexpNames() {
			if n == name && i < len(match) {
				return match[i]
			}
		}
		return ""
	}

	ip, port := parsePlayerIP(group("ipport"))

	return Player{
		ClientNum: parseIntSafe(group("num"), 0),
		Name:      group("name"),
		Ping:      parsePing(group("ping")),
		Score:     parseIntSafe(group("score"), 0),
		IP:        ip,
		Port:      port,
		QPort:     parseIntSafe(group("qport"), 0),
		GUID:      group("guid"),
		LastMsg:   parseIntSafe(group("lastmsg"), 0),
		Rate:      parseIntSafe(group("rate"), 0),
	}
}

func parsePlayerIP(ipport string) (string, int) {
	ip, portStr, ok := strings.Cut(ipport, ":")
	if !ok {
		return ipport, 0
	}
	port := parseIntSafe(portStr, 0)
	return ip, port
}

func parsePing(pingStr string) any {
	if pingStr == "LOAD" {
		return "LOAD"
	}
	if val, err := strconv.Atoi(pingStr); err == nil {
		return val
	}
	return pingStr
}

func parseIntSafe(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return val
}

func parseStatusKeyValueResponse(lines []string) map[string]string {
	var dataLine string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.EqualFold(trimmed, statusResponseMarker) {
			continue
		}
		if strings.HasPrefix(trimmed, kvDelimiter) || strings.Contains(trimmed, kvDelimiter) {
			dataLine += trimmed
		}
	}

	if dataLine == "" && len(lines) > 0 {
		dataLine = strings.TrimSpace(lines[len(lines)-1])
	}

	parts := strings.Split(dataLine, kvDelimiter)
	if len(parts) > 0 && parts[0] == "" {
		parts = parts[1:]
	}

	kvPairs := make(map[string]string, len(parts)/2)
	for i := 0; i < len(parts)-1; i += 2 {
		key := strings.TrimSpace(parts[i])
		if key == "" {
			continue
		}
		val := strings.TrimSpace(parts[i+1])
		kvPairs[key] = stripColorCodes(val)
	}

	return kvPairs
}

func populateServerStatusInfo(info *GetStatus, kvPairs map[string]string) {
	info.ComMaxClients = parseIntSafe(kvPairs["com_maxclients"], 0)
	info.GameType = kvPairs["g_gametype"]
	info.RandomSeed = parseIntSafe(kvPairs["g_randomSeed"], 0)
	info.GameName = kvPairs["gamename"]
	info.MapName = kvPairs["mapname"]
	info.PlaylistEnabled = parseBoolSafe(kvPairs, "playlist_enabled")
	info.PlaylistEntry = parseIntSafe(kvPairs["playlist_entry"], 0)
	info.Protocol = parseIntSafe(kvPairs["protocol"], 0)
	info.ScrTeamFFType = parseIntSafe(kvPairs["scr_team_fftype"], 0)
	info.ShortVersion = parseBoolSafe(kvPairs, "shortversion")
	info.SvAllowAimAssist = parseBoolSafe(kvPairs, "sv_allowAimAssist")
	info.SvAllowAnonymous = parseBoolSafe(kvPairs, "sv_allowAnonymous")
	info.SvClientFpsLimit = parseIntSafe(kvPairs["sv_clientFpsLimit"], 0)
	info.SvDisableClientConsole = parseBoolSafe(kvPairs, "sv_disableClientConsole")
	info.SvHostname = kvPairs["sv_hostname"]
	info.SvMaxClients = parseIntSafe(kvPairs["sv_maxclients"], 0)
	info.SvMaxPing = parseIntSafe(kvPairs["sv_maxPing"], 0)
	info.SvMinPing = parseIntSafe(kvPairs["sv_minPing"], 0)
	info.SvPatchDSR50 = parseBoolSafe(kvPairs, "sv_patch_dsr50")
	info.SvPrivateClients = parseIntSafe(kvPairs["sv_privateClients"], 0)

	// Handle variant field name for private clients
	if val, ok := kvPairs["sv_privateClientsForClients"]; ok {
		info.SvPrivateClientsForUsers = parseIntSafe(val, 0)
	} else {
		info.SvPrivateClientsForUsers = parseIntSafe(kvPairs["sv_privateClientsForUsers"], 0)
	}

	info.SvPure = parseBoolSafe(kvPairs, "sv_pure")
	info.SvVoice = parseBoolSafe(kvPairs, "sv_voice")
	info.PasswordEnabled = parseBoolSafe(kvPairs, "pswrd")
	info.ModEnabled = parseBoolSafe(kvPairs, "mod")
}

func (r *RCON) compileDvarPatterns(dvar string) []*regexp.Regexp {
	escapedDvar := regexp.QuoteMeta(strings.TrimSpace(dvar))

	pattern1 := regexp.MustCompile(fmt.Sprintf(`(?i)^"?%s"?\s+is:\s+"?(?P<val>.*?)"?(?:\s|$)`, escapedDvar))
	pattern2 := regexp.MustCompile(fmt.Sprintf(`(?i)^"?%s"?\s*[:=]\s*"?(?P<val>.*?)"?$`, escapedDvar))
	return []*regexp.Regexp{pattern1, pattern2}
}

func (r *RCON) queryDvar(dvar string) ([]string, error) {
	packet := r.buildPacket(dvar, true)
	if err := r.sendPacket(packet); err != nil {
		return nil, err
	}

	return r.readResponse()
}

func extractDvarValueFromLines(lines []string, regexes []*regexp.Regexp) string {
	for _, line := range lines {

		clean := strings.TrimSpace(stripColorCodes(line))
		if clean == "" {
			continue
		}

		for _, regex := range regexes {

			match := regex.FindStringSubmatch(clean)
			if match == nil {
				continue
			}

			for i, name := range regex.SubexpNames() {
				if name == "val" && i < len(match) {

					val := stripColorCodes(match[i])
					val = strings.TrimSpace(val)

					return val
				}
			}
		}
	}

	return ""
}

func findFallbackValue(lines []string) string {
	for _, line := range lines {
		clean := strings.TrimSpace(stripColorCodes(line))
		if clean == "" {
			continue
		}

		if !strings.Contains(strings.ToLower(clean), "sv_iw4madmin_in") {
			return clean
		}
	}
	return ""
}

func shouldRetryDvarQuery(lines []string) bool {
	for _, line := range lines {
		if strings.Contains(strings.ToLower(line), "sv_iw4madmin_in") {
			return true
		}
	}
	return false
}
