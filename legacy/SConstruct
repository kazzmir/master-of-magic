import os

env = Environment(ENV = os.environ)

if False:
    ocaml_builder = Builder(action = 'ocamlopt -ccopt -L$OCAMLPATH -I $OCAMLINCLUDE $OCAMLLIBS $SOURCE -o $TARGET',
        suffix = '',
        src_suffix = '.ml')

    ocaml_library = Builder(action = 'ocamlopt -a -o $TARGET -I $OCAMLINCLUDE $SOURCES',
        suffix = '.cmxa',
        src_suffix = ['.ml', '.cmi'],
        chdir = 1)

    env.Append(OCAMLINCLUDE = ['lib/extlib-1.5'])
    env.Append(OCAMLPATH = ['lib/'])
    env.Append(OCAMLLIBS = ['extLib.cmxa'])
    env.Append(BUILDERS = {'OcamlProgram' : ocaml_builder,
                       'OcamlLibrary' : ocaml_library})

env.Tool('scons/ocaml', '.')
env['OCAML_CODE'] = 'native'
env['OCAML_PATH'] = ['build-lib/extlib-1.5']
env.BuildDir('build', 'ocaml')
libs = env.SConscript('lib/SConstruct', exports = 'env', build_dir = 'build-lib')
env.Install('build/build-lib/', libs)
# lbxreader = env.OcamlProgram('lbxreader', 'build/lbxreader.ml', OCAML_LIBS = libs, OCAML_PATH = ['lib/extlib-1.5'])
lbxreader = env.OcamlProgram('lbxreader', 'build/lbxreader.ml', OCAML_LIBS = libs, OCAML_PATH = ['build-lib/extlib-1.5'])
#env.Depends(lbxreader, 'lib/extlib.cmxa')
# env.Install('.', lbxreader)
