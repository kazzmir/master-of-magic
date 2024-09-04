.PHONY: magic magic.wasm lbxdump wasm itch.io

magic:
	go build -o magic ./game/magic

wasm: magic.wasm

magic.wasm:
	env GOOS=js GOARCH=wasm go build -o magic.wasm ./game/magic

lbxdump:
	go build -o lbxdump ./util/lbxdump

itch.io: wasm
	cp magic.wasm itch.io
	butler push itch.io kazzmir/magic:html

update:
	go get -u ./game/magic
	go mod tidy
