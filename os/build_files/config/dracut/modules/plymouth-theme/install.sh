#!/bin/bash
# Install Plymouth connect theme into initramfs
mkdir -p ${initdir}/usr/share/plymouth/themes/connect
cp -a /usr/share/plymouth/themes/connect/* ${initdir}/usr/share/plymouth/themes/connect/
