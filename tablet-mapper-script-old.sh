#!/bin/sh
set -x
#Change DVI-I-1 to what monitor you want from running command: xrandr

echo "Going to map the tablet to the primary monitor"
MONITOR=$(xrandr | grep " connected" | grep primary | cut -d' ' -f1)
PAD_NAME='HUION H420 Pad pad'

#undo
xsetwacom --set "$PAD_NAME" Button 1 "key +ctrl +z -z -ctrl" 

#define next 2 however you like, I have mine mapped for erase in krita
xsetwacom --set "$PAD_NAME" Button 2 "key e"
xsetwacom --set "$PAD_NAME" Button 3 "key h"

ID_STYLUS=`xinput | grep "Pen stylus" | cut -f 2 | cut -c 4-5`


xinput map-to-output $ID_STYLUS $MONITOR
xsetwacom --set "$PAD_NAME" MapToOutput "$MONITOR"
xsetwacom --set "$ID_STYLUS" MapToOutput "HEAD-0"
xsetwacom --set "$PAD_NAME" MapToOutput "HEAD-0"

xsetwacom --set "$ID_STYLUS" Mode Absolute
xsetwacom --set "$PAD_NAME" Mode Absolute

echo configured pad: "'"$PAD_NAME"'" and stylus: "'"$ID_STYLUS"'" to use monitor: "'"$MONITOR"'"
exit 0
