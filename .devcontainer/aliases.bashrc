alias hms="meson setup builddir_host"
alias xms="meson setup builddir_kindlehf --cross-file ~/x-tools/arm-kindlehf-linux-gnueabihf/meson-crosscompile.txt"
alias hmc="meson compile -C builddir_host"
alias xmc="meson compile -C builddir_kindlehf"