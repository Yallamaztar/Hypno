ENTRY=cmd\plugin\hypno.go
OUTPUT=hypno_plugin.exe
FLAGS=-ldflags="-s -w" -trimpath

build:
	go build $(FLAGS) -o $(OUTPUT) $(ENTRY)

run:
	go run $(ENTRY)