all: deps lbxreader.native spritereader.native main.native terrain.native

ocamlallegro = lib/ocaml-allegro-20080222

.PHONY: deps

lbxreader.native: src/lbxreader.ml src/utils.ml
	ocamlbuild -j 2 -Is src,lib/extlib-1.5 lbxreader.native

terrain.native: src/terrain.ml
	ocamlbuild -j 2 -Is src,lib/extlib-1.5 terrain.native

main.native: src/graphics.ml src/windows.ml src/main.ml src/gamedata.ml
	ocamlbuild -j 2 -lflag -ccopt -lflag -L. -Is src,lib/extlib-1.5,${ocamlallegro} -libs unix,allegro main.native

# You have to build allegro and copy the following files to _build
# dll_alleg_stubs.so
# lib_alleg_stubs.a
# allegro.a
# allegro.cmxa
spritereader.native: src/lbxreader.ml src/utils.ml src/spritereader.ml
	ocamlbuild -j 2 -lflag -ccopt -lflag -L. -Is src,lib/extlib-1.5,${ocamlallegro} -libs unix,allegro spritereader.native

clean:
	ocamlbuild -clean

count:
	wc -l src/*.ml

deps: save/allegro.cmxa
	mkdir -p _build/; cp save/* _build/

save/allegro.cmxa: ${ocamlallegro}/alleg-wrap.c ${ocamlallegro}/allegro.ml
	cd ${ocamlallegro}; $(MAKE)
