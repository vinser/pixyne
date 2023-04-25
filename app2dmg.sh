#!/bin/bash

dpkg --verify hfsprogs &>/dev/null
if [ "$?" -gt 0 ] ; then
    sudo apt install hfsprogs # not installed
fi

# turn macOS .app to dmg
dd if=/dev/zero of=/tmp/Pixyne.dmg bs=1M count=19 status=progress 
mkfs.hfsplus -v Pixyne /tmp/Pixyne.dmg

sudo mkdir -pv /mnt/tmp
sudo mount -o loop /tmp/Pixyne.dmg /mnt/tmp
sudo cp -av fyne-cross/dist/darwin-amd64/Pixyne.app /mnt/tmp

sudo umount /mnt/tmp

cp /tmp/Pixyne.dmg fyne-cross/dist/darwin-amd64/Pixyne.dmg