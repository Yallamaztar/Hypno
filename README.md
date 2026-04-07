# Hypno - Plugin Made For Plutonium T6 (WIP)

[Hypno Banner](.github/HypnoBanner.png)

## Abouts
**Hypno** Aka **Hypnosis Plugin** is a server-side plugin made for **[Plutonium T6](https://plutonium.pw/)** dedicated servers. It adds custom chat commands, economy, gambling, and shop systems, along with **[IW4M-Admin](https://github.com/RaidMax/IW4M-Admin)** and Discord integrations.

## Features
- Custom chat commands for enhanced server interaction.
- Economy system to manage in-game currency.
- Gambling and shop systems for player engagement.
- Integration with **IW4M-Admin** for advanced server management.
- Discord integration for seamless communication.

## Setup
### 1. Clone the repository
```
git clone https://github.com/Yallamaztar/Hypno.git
```

### 2. Build the plugin
Navigate to the repository folder:
```
cd Hypno
```

#### Option 1: Makefile
```
make build
```

#### Option 2: Build with Go directly
```
go build cmd/plugin/hypno.go -o hypno_plugin.exe
```

### 3. Run the plugin
```
hypno_plugin.exe
```

