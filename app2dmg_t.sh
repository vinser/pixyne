#!/bin/bash

echo "start app2dmg_t.sh"
dpkg --verify hfsprogs &>/dev/null
if [ "$?" -gt 0 ] ; then
echo "pkg \"hfsprogs\" not found"
# sudo apt install hfsprogs
fi
# cp /tmp/Pixyne.dmg fyne-cross/dist/darwin-amd64/Pixyne.dmg
echo "finish app2dmg_t.sh"
