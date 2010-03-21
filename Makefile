lbxreader.native: src/*.ml
	ocamlbuild -Is src,lib/extlib-1.5 lbxreader.native

clean:
	ocamlbuild -clean

count:
	wc -l src/*.ml
