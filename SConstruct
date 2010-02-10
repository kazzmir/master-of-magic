import os

env = Environment(ENV = os.environ)

ocaml_builder = Builder(action = 'ocamlopt -ccopt -L$OCAMLPATH -I $OCAMLINCLUDE $OCAMLLIBS $SOURCE -o $TARGET',
        suffix = '',
        src_suffix = '.ml')

ocaml_library = Builder(action = 'ocamlopt -a -o $TARGET -I $OCAMLINCLUDE $SOURCES',
        suffix = '.cmxa',
        src_suffixes = ['.ml', '.mli'])

env.Append(OCAMLINCLUDE = ['lib/extlib-1.5'])
env.Append(OCAMLPATH = ['lib/'])
env.Append(OCAMLLIBS = ['extLib.cmxa'])
env.Append(BUILDERS = {'OcamlProgram' : ocaml_builder,
                       'OcamlLibrary' : ocaml_library})

env.SConscript('lib/SConstruct', exports = 'env')

env.BuildDir('build', 'src')
lbxreader = env.OcamlProgram('build/lbxreader')
env.Depends(lbxreader, 'lib/extlib.cmxa')
env.Install('.', lbxreader)
