#!/bin/env sh
clang -o tablet-mapper main.c -lm -ldl -lpthread -lX11 -lxcb -lGL -lGLX -lXext -lGLdispatch -lXau -lXdmcp -lraylib 
