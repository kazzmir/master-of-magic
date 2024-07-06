.PHONY: magic

magic:
	go build -o magic ./game/magic

update:
	go get -u ./game/magic
	go mod tidy
