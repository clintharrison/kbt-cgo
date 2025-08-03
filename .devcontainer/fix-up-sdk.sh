#!/bin/bash

apt -y install patchelf

kindle_sysroot=/root/x-tools/arm-kindlehf-linux-gnueabihf/arm-kindlehf-linux-gnueabihf/sysroot/
libace_path=$kindle_sysroot/usr/lib/libace_bt.so

# linking against the system ones is awkward because they don't have a SONAME set??
patchelf --set-soname $(basename $libace_path) $libace_path

cat > $kindle_sysroot/usr/lib/pkgconfig/ace_bt.pc <<'EOF'
prefix=/usr
exec_prefix=${prefix}
includedir=${prefix}/include
libdir=${exec_prefix}/lib

Name: ace_bt
Description: The ACE Bluetooth library
Version: 1.0.0
Cflags: -I${includedir}
Libs: -L${libdir} -lace_bt
EOF

cat > $kindle_sysroot/usr/lib/pkgconfig/ace_osal.pc <<'EOF'
prefix=/usr
exec_prefix=${prefix}
includedir=${prefix}/include
libdir=${exec_prefix}/lib

Name: ace_osal
Description: The ACE OSAL
Version: 1.0.0
Cflags: -I${includedir}
Libs: -L${libdir} -lace_osal
EOF