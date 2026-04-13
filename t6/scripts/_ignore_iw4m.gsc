/*
 * _ignore_iw4m.gsc
 *
 * This script is intended for servers using IW4M-Admin Game Interface.
 * Place it alongside your other scripts to override the default Game Interface
 * command registration.
 *
 * It replaces the original command registration function so that IW4M-Admin
 * ignores Hypno plugin commands by registering them to IW4m-Admin, preventing 
 * "unknown command" messages from appearing in the chat
 */

main() {
    // Override all game interface commands since they are implemented in Hypno
    ReplaceFunc(scripts\_integration_t6::RegisterClientCommands, ::null);
}

init() {
    waittillframeend;
    level waittill(level.notifyTypes.gameFunctionsInitialized);

    // Developer commands
    scripts\_integration_shared::RegisterScriptCommand("Hypno_printmoney", "printmoney", "print", "Print more money to the bank", "Owner", "T6", false, ::null);
    scripts\_integration_shared::RegisterScriptCommand("Hypno_xuid", "xuid", "info", "Show a players name, xuid and clientnum", "Owner", "T6", false, ::null);
    scripts\_integration_shared::RegisterScriptCommand("Hypno_crash", "crash", "panic", "Make the Hypno Plugin crash", "Owner", "T6", false, ::null);
    scripts\_integration_shared::RegisterScriptCommand("Hypno_rcon", "rcon", "rc", "Execute a RCON command", "Owner", "T6", false, ::null);
    scripts\_integration_shared::RegisterScriptCommand("Hypno_lookup", "lookup", "find", "Lookup a player by name", "Owner", "T6", false, ::null);
    scripts\_integration_shared::RegisterScriptCommand("Hypno_discordinvite", "discordinvite", "invite", "Change the discord invite link", "Owner", "T6", false, ::null);

    // Owner commands
    scripts\_integration_shared::RegisterScriptCommand("Hypno_gambling", "gambling", "gmbl", "Enable/disable gambling or view status", "Owner", "T6", false, ::null);
    scripts\_integration_shared::RegisterScriptCommand("Hypno_maxbet", "maxbet", "mb", "Set the maximum bet amount", "Owner", "T6", false, ::null);
    scripts\_integration_shared::RegisterScriptCommand("Hypno_addowner", "addowner", "ao", "Add a new owner", "Owner", "T6", false, ::null);
    scripts\_integration_shared::RegisterScriptCommand("Hypno_removeowner", "removeowner", "ro", "Remove an owner", "Owner", "T6", false, ::null);
    scripts\_integration_shared::RegisterScriptCommand("Hypno_addadmin", "addadmin", "aa", "Add a new admin", "Owner", "T6", false, ::null);
    scripts\_integration_shared::RegisterScriptCommand("Hypno_removeadmin", "removeadmin", "ra", "Remove an admin", "Owner", "T6", false, ::null);

    // Admin commands
    scripts\_integration_shared::RegisterScriptCommand("Hypno_freeze", "freeze", "fz", "Freeze a player", "SeniorAdmin", "T6", false, ::null);
    scripts\_integration_shared::RegisterScriptCommand("Hypno_setspeed", "setspeed", "ss", "Set a player's speed", "SeniorAdmin", "T6", false, ::null);
    scripts\_integration_shared::RegisterScriptCommand("Hypno_swap", "swap", "swp", "Swap two players", "SeniorAdmin", "T6", false, ::null);
    scripts\_integration_shared::RegisterScriptCommand("Hypno_dropgun", "dropgun", "dg", "Drop a player's gun", "SeniorAdmin", "T6", false, ::null);
    scripts\_integration_shared::RegisterScriptCommand("Hypno_teleport", "teleport", "tp", "Teleport to a player", "SeniorAdmin", "T6", false, ::null);
    scripts\_integration_shared::RegisterScriptCommand("Hypno_sayas", "sayas", "says", "Say as a player", "SeniorAdmin", "T6", false, ::null);
    scripts\_integration_shared::RegisterScriptCommand("Hypno_stealmoney", "stealmoney", "steal", "Steal money from a player", "SeniorAdmin", "T6", false, ::null);
    scripts\_integration_shared::RegisterScriptCommand("Hypno_givemoney", "givemoney", "gi", "Give money to a player", "SeniorAdmin", "T6", false, ::null);
    scripts\_integration_shared::RegisterScriptCommand("Hypno_giveall", "giveall", "ga", "Give money to all players", "SeniorAdmin", "T6", false, ::null);
    scripts\_integration_shared::RegisterScriptCommand("Hypno_setorigin", "setorigin", "so", "Set a player's origin", "SeniorAdmin", "T6", false, ::null);
}

// null function just returns nothing
null(){return;}