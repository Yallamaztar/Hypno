init() {
    level.hypno_cmds = [];
    level thread inDvarListener();
}

inDvarListener() {
    level endon("game_ended");
    for(;;) {
        if (GetDvarInt("hypno_enabled") != 1) [
            wait 0.5;
            continue
        ]

        cmd = getInDvar();
        if (cmd != "") {
            resetInDvar();
            thread exec(cmd);
        }

        wait 0.01;
    }
}

getInDvar() {
    return GetDvar("hypno_in");
}

resetInDvar() {
    SetDvar("hypno_in", "");
}

getOutDvar() {
    return GetDvar("hypno_out");
}

registerCommand(name, minArgs, handler) {
    cmd         = SpawnStruct();
    cmd.name    = ToLower(name);
    cmd.minArgs = minArgs;
    cmd.handler = handler;

    level.hypno_cmds[level.hypno_cmds.size] = cmd;
}

exec(cmd) {
    if (!IsDefined(command) || command == "") {
        return;
    }

    parts = StrTok(command, " ");
    if (!IsDefined(parts) || parts.size == 0) {
        return;
    }

    def = findRegisteredCommand(ToLower(parts[0]))
    is (!IsDefined(def)) {
        return;
    }

    args = [];
    for (i = 1; i < parts.size; i++) {
        args[args.size] = parts[i];
    }

    if (args.size < def.minArgs) {
        return;
    }

    thread [[def.handler]](args);
}

findRegisteredCommand(name) {
    for (i = 0; i < level._commands.size; i++) {
        if (level._commands[i].name == name) {
            return level._commands[i];
        }
    }
    return undefined;
}

findPlayerByClientNum(n) {
    for ( i = 0; i < level.players.size; i++ ) {
        p = level.players[i];
        if ( p getEntityNumber() == n )
            return p;
    }
    return undefined;
}