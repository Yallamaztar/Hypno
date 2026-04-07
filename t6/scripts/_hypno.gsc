init() {
    level.hypno_cmds = [];
    SetDvar("hypno_enabled", 1);
    level registerClientCommands();
    level thread inDvarListener();
}

// polls hypno_in dvar for commands and executes them
inDvarListener() {
    level endon("game_ended");
    for(;;) {
        if (GetDvarInt("hypno_enabled") != 1) {
            wait 0.1;
            continue;
        }

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

setOutDvar(val) {
    SetDvar("hypno_out", val);
}

registerCommand(name, minArgs, handler) {
    cmd         = SpawnStruct();
    cmd.name    = ToLower(name);
    cmd.minArgs = minArgs;
    cmd.handler = handler;

    level.hypno_cmds[level.hypno_cmds.size] = cmd;
}

// splits command, validates arguments and calls the handler
exec(cmd) {
    if (!IsDefined(cmd) || cmd == "") {
        return;
    }
    parts = StrTok(cmd, " ");
    if (!IsDefined(parts) || parts.size == 0) {
        return;
    }

    def = findRegisteredCommand(ToLower(parts[0]));
    if (!IsDefined(def)) {
        return;
    }

    // build args
    args = [];
    for (i = 1; i < parts.size; i++) {
        args[args.size] = parts[i];
    }

    if (args.size < def.minArgs) {
        return;
    }

    // call handler
    thread [[def.handler]](args);
}

findRegisteredCommand(name) {
    for (i = 0; i < level.hypno_cmds.size; i++) {
        if (level.hypno_cmds[i].name == name) {
            return level.hypno_cmds[i];
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

registerClientCommands() {
    registerCommand("plugin_ready", 0, ::on_ready);

    registerCommand("freeze",       2, ::impl_freeze);
    registerCommand("setspeed",     2, ::impl_setspeed);
    registerCommand("swap",         2, ::impl_swap);
    registerCommand("dropgun",      2, ::impl_dropgun);
    registerCommand("teleport",     2, ::impl_teleport);
    registerCommand("setorigin",    4, ::impl_setorigin);
    registerCommand("giveweapon",   2, ::impl_giveweapon);
    registerCommand("takeweapons",  1, ::impl_takeweapons);
    registerCommand("switchteams",  1, ::impl_switchteams);
    registerCommand("hide",         2, ::impl_hide);
    registerCommand("alert",        2, ::impl_alert);
    registerCommand("killplayer",   1, ::impl_killplayer);
    registerCommand("setspectator", 1, ::impl_setspectator);

}

on_ready() {
    setOutDvar("success");
}

impl_freeze(args) {
    origin = findPlayerByClientNum(args[0]);
    target = findPlayerByClientNum(args[1]);

    if (!IsDefined(target.frozen)) {
        target.frozen = false;
    }

    if (!target.frozen) { 
        target FreezeControls(true);
        target.frozen = true;
        origin IPrintLnBold("Froze player " + target.name);
    } else {
        target FreezeControls(false);
        target.frozen = false;
        origin IPrintLnBold("Unfroze player " + target.name);
    }
}

impl_setspeed(args) {
    target = findPlayerByClientNum(args[0]);
    speed  = Float(args[1]);
    if (!IsFloat(speed)) {
        return;
    }

    target SetMoveSpeedScale(speed);
}

impl_swap(args) {
    target1 = findPlayerByClientNum(args[0]);
    target2 = findPlayerByClientNum(args[1]);

    org1 = target1.origin;
    org2 = target2.origin;

    target1 SetOrigin(org2);
    target2 SetOrigin(org1);
}

impl_dropgun(args) {
    target = findPlayerByClientNum(args[0]);
    weap   = target GetCurrentWeapon();
    target DropItem(weap);
}

impl_teleport(args) {
    target1 = findPlayerByClientNum(args[0]);
    target2 = findPlayerByClientNum(args[1]);
    target1 SetOrigin(target2.origin);
}

impl_setorigin(args) {
    target = findPlayerByClientNum(args[0]);
    // use () not array() or it breaks silently
    coords = (int(args[1]), int(args[2]), int(args[3]));
    target SetOrigin(coords);
}

impl_giveweapon(args) {
    target = findPlayerByClientNum(args[0]);
    target GiveWeapon(args[1]);
}

impl_takeweapons(args) {
    target = findPlayerByClientNum(args[0]);
    target TakeAllWeapons();
}

impl_switchteams(args) {
    target = findPlayerByClientNum(args[0]);
    team = level.allies;
    if (target.team == "allies") team = level.axis;
    target [[team]]();
}

impl_hide(args) {
    origin = findPlayerByClientNum(args[0]);
    target = findPlayerByClientNum(args[1]);

    if (!IsDefined(target.hidden)) {
        target.hidden = false;
    }

    if (!target.hidden) {
        target Hide();
        target.hidden = true;
        origin IPrintLnBold("Hidden ^6" + target.name);
    } else {
        target Show();
        target.hidden = false;
        origin IPrintLnBold("Unhidden ^6" + target.name);
    }
}

impl_alert(args) {
    target = findPlayerByClientNum(args[0]);
    
    message = "";
    for (i=1; i<args.size; i++) {
        message += args[i]+" ";
    }

    alert = spawnstruct();
    alert.titletext  = "Alert";
    alert.notifytext = message;
    alert.iconname   = undefined;
    alert.sound      = "mpl_sab_ui_suitcasebomb_timer";
    alert.duration   = 7.5;
    target.startmessagenotifyqueue[target.startmessagenotifyqueue.size] = alert;
    target notify("received award");
}

impl_killplayer(args) {
    target = findPlayerByClientNum(args[0]);
    target Kill();
}

impl_setspectator(args) {
    target = findPlayerByClientNum(args[0]);
    target [[level.spectator]]();
}

impl_bringall(args) {
    target = findPlayerByClientNum(args[0]);
    foreach(player in level.players) {
        player SetOrigin(target.origin);
    }
}

impl_savepos(args) {
    target = findPlayerByClientNum(args[0]);
    target.pers["saved_pos"] = target.origin;
}

impl_loadpos(args) {
    target = findPlayerByClientNum(args[0]);
    if (!IsDefined(target.pers["saved_pos"])) {
        return;
    }

    target SetOrigin(target.pers["saved_pos"]);
}

