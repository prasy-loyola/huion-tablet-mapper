# Tablet mapper

## References

### Map the tablet to screen
- https://wiki.ubuntu.com/X/InputCoordinateTransformation
- https://wiki.archlinux.org/title/Calibrating_Touchscreen#Using_nVidia%27s_TwinView
- https://askubuntu.com/questions/801761/huion-tablet-drawing-area

### Map buttons in the tablet
- https://github.com/DIGImend/digimend-kernel-drivers#enabling-wacom-xorg-driver


```
Section "InputClass"
    Identifier "Tablet"
    Driver "wacom"
    MatchDevicePath "/dev/input/event*"
    MatchUSBID "<VID>:<PID>"
EndSection
```
