.PHONY: all.zip

all.zip:
	$(MAKE) -C .. wasm
	cp ../magic.wasm .
	zip all.zip go1.23.0-wasm.js index.html magic.wasm
