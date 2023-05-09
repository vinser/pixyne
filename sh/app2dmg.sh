#!/bin/bash

dpkg --verify hfsprogs &>/dev/null
if [ "$?" -gt 0 ] ; then
    sudo apt install hfsprogs # not installed
fi

# turn macOS .app to dmg
dd if=/dev/zero of=fyne-cross/tmp/darwin-$1/Pixyne.dmg bs=1M count=19 status=progress 
mkfs.hfsplus -v Pixyne fyne-cross/tmp/darwin-$1/Pixyne.dmg

sudo mkdir -pv /mnt/tmp
sudo mount -o loop fyne-cross/tmp/darwin-$1/Pixyne.dmg /mnt/tmp
sudo cp -av fyne-cross/dist/darwin-$1/Pixyne.app /mnt/tmp

sudo umount /mnt/tmp

cp fyne-cross/tmp/darwin-$1/Pixyne.dmg fyne-cross/pixyne-macosx-$1.dmg