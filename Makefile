all: deps lbxreader.native spritereader.native

.PHONY: deps

lbxreader.native: src/*.ml
	ocamlbuild -j 2 -Is src,lib/extlib-1.5 lbxreader.native

# You have to build allegro and copy the following files to _build
# dll_alleg_stubs.so
# lib_alleg_stubs.a
# allegro.a
# allegro.cmxa
spritereader.native: src/*.ml
	ocamlbuild -j 2 -lflag -ccopt -lflag -L. -Is src,lib/extlib-1.5,lib/ocaml-allegro-20080222 -libs unix,allegro spritereader.native

clean:
	ocamlbuild -clean

count:
	wc -l src/*.ml

deps:
	mkdir -p _build/; cp save/* _build/
