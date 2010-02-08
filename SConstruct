import os

env = Environment(ENV = os.environ)

ocaml_builder = Builder(action = 'ocamlopt -ccopt -L$OCAMLPATH -I $OCAMLPATH $OCAMLLIBS $SOURCE -o $TARGET',
        suffix = '',
        src_suffix = '.ml')

env.Append(OCAMLPATH = ['lib/extlib-1.5'])
env.Append(OCAMLLIBS = ['extLib.cmxa'])
env.Append(BUILDERS = {'OcamlProgram' : ocaml_builder})

env.BuildDir('build', 'src')
env.Install('.', env.OcamlProgram('build/lbxreader'))
