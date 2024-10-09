# Tablet mapper

## References

### Map the tablet to screen
- https://wiki.ubuntu.com/X/InputCoordinateTransformation
- https://wiki.archlinux.org/title/Calibrating_Touchscreen#Using_nVidia%27s_TwinView
- https://askubuntu.com/questions/801761/huion-tablet-drawing-area

### Map buttons in the tablet
- https://github.com/DIGImend/digimend-kernel-drivers#enabling-wacom-xorg-driver

### Wayland support
https://unix.stackexchange.com/questions/696514/configuring-xp-pen-graphics-tablet-on-linux-specifically-wayland

path: "/etc/X11/xorg.conf.d/10-tablet.conf"

```
Section "InputClass"
  Identifier "Tablet"
  Driver "wacom"
  MatchDevicePath "/dev/input/event*"
  MatchUSBID "256c:006e"
EndSection
```
