all: lbxreader.native spritereader.native

lbxreader.native: src/*.ml
	ocamlbuild -Is src,lib/extlib-1.5 lbxreader.native

spritereader.native: src/*.ml
	ocamlbuild -Is src,lib/extlib-1.5 spritereader.native

clean:
	ocamlbuild -clean

count:
	wc -l src/*.ml
