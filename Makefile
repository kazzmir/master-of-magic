.PHONY: magic magic.wasm lbxdump wasm

magic:
	go build -o magic ./game/magic

wasm: magic.wasm

magic.wasm:
	env GOOS=js GOARCH=wasm go build -o magic.wasm ./game/magic

lbxdump:
	go build -o lbxdump ./util/lbxdump

update:
	go get -u ./game/magic
	go mod tidy
