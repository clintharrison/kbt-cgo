project(
  'go-bt-test',
  'cpp',
  version: 'v0.0.1',
  default_options: ['cpp_std=c++17'],
  meson_version: '>=1.1',
)
add_languages('c')

go_sources = files(
  './cmd/main.go',
)

cc = meson.get_compiler('c')
ldflags = ' '.join(cc.get_supported_link_arguments([]))
cflags = ' '.join(cc.get_supported_arguments([]))
gotool = find_program('go')
pkg_config = find_program('pkg-config')
pkg_config_path = get_option('pkg_config_path')

goenv = environment()
goenv.set('CC', ' '.join(cc.cmd_array()))
goenv.set('CGO_ENABLED', '1')
goenv.set('CGO_CFLAGS', cflags)
goenv.set('CGO_LDFLAGS', ldflags)
goenv.set('PKG_CONFIG', pkg_config.full_path())
goenv.set('PKG_CONFIG_PATH', ':'.join(pkg_config_path))

goenv.set('GOOS', host_machine.system())
goenv.set('GOARCH', host_machine.cpu_family())

if build_machine.cpu_family() == 'arm'
  goenv.set('GOARM', '7')
endif
libace_dep = dependency('ace_bt', method: 'pkg-config')

custom_target(
  'go-bt-test',
  output: 'go-bt-test',
  env: goenv,
  build_always_stale: true,
  input: go_sources,
  # TODO: build -a unfortunately needed due to linking with C libraries that might change.
  #   Is that actually needed? It's a copy/paste from someone on GitHub...
  #   ...and it's so much faster not to do that. Run it manually if you need to!
  command: [gotool, 'build', '-o', '@OUTPUT@', '@INPUT@'],
  build_always_stale: true,
  build_by_default: true,
)