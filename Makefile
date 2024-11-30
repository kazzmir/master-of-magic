.PHONY: magic magic.wasm lbxdump wasm itch.io test

magic:
	go mod tidy
	go build -o magic ./game/magic

wasm: magic.wasm

magic.wasm:
	env GOOS=js GOARCH=wasm go build -o magic.wasm ./game/magic

lbxdump:
	go build -o lbxdump ./util/lbxdump

itch.io: wasm
	cp magic.wasm itch.io
	butler push itch.io kazzmir/magic:html

combat-simulator-itch.io:
	env GOOS=js GOARCH=wasm go build -o ./util/combat-simulator/itch.io/combat-simulator.wasm ./util/combat-simulator
	butler push ./util/combat-simulator/itch.io kazzmir/magic-combat-simulator:html

update:
	go get -u ./game/magic
	go mod tidy

test:
	go test ./...
