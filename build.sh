#!/bin/env sh
sudo apt install libx11-dev libxcursor-dev libxrandr-dev libxinerama-dev libxi-dev libgl1-mesa-dev

cd exporter
go run exporter.go
cp -f font.go ../
cd ..
go build -o tablet-mapper .
