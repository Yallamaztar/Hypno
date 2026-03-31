#define PLAYERNOTFOUND

init() {
    level.hypno_cmds = [];
    level thread inDvarListener();
}

inDvarListener() {
    level endon("game_ended");
    for(;;) {
        if (GetDvarInt("hypno_enabled") != 1) {
            wait 0.5;
            continue
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

registerClientCommands() {
    registerCommand("plugin_ready", 0, ::on_ready);

    registerCommand("freeze", 2, ::impl_freeze);
    registerCommand("setspeed", 2, ::impl_setspeed);
    registerCommand("swap", 2, ::impl_swap);
    registerCommand("dropgun", 2, ::impl_dropgun);
    registerCommand("teleport", 2, ::impl_teleport);
    registerCommand("setorigin", 4, ::impl_setorigin);
}

on_ready() {
    setOutDvar("success");
}

impl_freeze(args) {
    origin = findPlayerByClientNum(args[0]);
    target = findPlayerByClientNum(args[1]);

    // toggling on/off with persistent value
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
    coords = array(int(args[1]), int(args[2]), int(args[3]));
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
    target [[team]]()
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
    target.startmessagenotifyqueue[self.startmessagenotifyqueue.size] = alert;
    target notify("received award");
}
