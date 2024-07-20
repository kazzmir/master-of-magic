.PHONY: magic lbxdump

magic:
	go build -o magic ./game/magic

lbxdump:
	go build -o lbxdump ./util/lbxdump

update:
	go get -u ./game/magic
	go mod tidy
