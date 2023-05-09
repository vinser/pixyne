#!/bin/bash

dpkg --verify msitools &>/dev/null
if [ "$?" -gt 0 ] ; then
    sudo apt apt install msitools # not installed
fi

dpkg --verify wixl &>/dev/null
if [ "$?" -gt 0 ] ; then
    sudo apt apt install wixl # not installed
fi

# build windows installer msi
if [ -n "$1" ] ; then
    sed -i -r 's/windows-(amd64|386)/windows-'$1'/g' sh/app2msi.wsx
    wixl -v -o fyne-cross/pixyne-windows-$1.msi sh/app2msi.wsx
fi