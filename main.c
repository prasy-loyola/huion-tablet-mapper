#include "raylib.h"
#include <math.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int main(int argc, char **argv) {

  Color bg_color = DARKGRAY;
  const static char *TABLET_STYLUS_NAME = "HUION H420 Pen stylus";
  const static char *TABLET_NAME = "HUION H420 Pad pad";
  const static char *MAP_KEY = "xsetwacom --set \"%s\" Button %d \"key %s\"";
  const static char *UPDATE_COORD_TRANS_MATRIX =
      "xinput set-prop \"%s\" "
      "--type=float "
      "\"Coordinate Transformation Matrix\" "
      "%f 0 %f 0 %f %f 0 0 1";

  SetConfigFlags(FLAG_WINDOW_RESIZABLE);
  SetTargetFPS(60);
  InitWindow(800, 450, "Huion Tablet mapper");
  char text[3000] = {0};
  char info[3000] = {0};
  bool mapped = false;
  int prev_mapped_width = GetRenderWidth();
  int prev_mapped_height = GetRenderHeight();
  while (!WindowShouldClose()) {
    if (IsWindowResized()) {
      int curr_width = GetRenderWidth();
      int curr_height = GetRenderHeight();
      int new_height = floorf(curr_width * (2.23 / 4));
      prev_mapped_width = curr_width;
      prev_mapped_height = new_height;
      SetWindowSize(curr_width, new_height);
    }
    if (IsKeyDown(KEY_M)) {
      int monitor_count = GetMonitorCount();
      int screen_width = 0;
      int screen_height = 0;
      for (int i = 0; i < monitor_count; ++i) {
        Vector2 monitor_pos = GetMonitorPosition(i);
        if (screen_width < GetMonitorWidth(i) + monitor_pos.x) {
          screen_width = GetMonitorWidth(i) + monitor_pos.x;
        }
        if (screen_height < GetMonitorHeight(i) + monitor_pos.y) {
          screen_height = GetMonitorHeight(i) + monitor_pos.y;
        }
      }

      int curr_height = GetRenderHeight();
      int curr_width = GetRenderWidth();
      Vector2 window_pos = GetWindowPosition();
      // c0 = touch_area_width / total_width
      float c0 = (float)curr_width / screen_width;
      // c2 = touch_area_height / total_height
      float c2 = (float)curr_height / screen_height;
      // c1 = touch_area_x_offset / total_width
      float c1 = window_pos.x / screen_width;
      // c3 = touch_area_y_offset / total_height
      float c3 = window_pos.y / screen_height;

      sprintf(text, UPDATE_COORD_TRANS_MATRIX, TABLET_NAME, c0, c1, c2, c3);
      printf("%s\n", text);
      int result1 = system(text);
      mapped = true;
      if (result1 != 0) {
        mapped = false;
      }

      sprintf(text, UPDATE_COORD_TRANS_MATRIX, TABLET_STYLUS_NAME, c0, c1, c2,
              c3);
      printf("%s\n", text);
      int result2 = system(text);

      if (result2 != 0 || result1 != 0) {
        mapped = false;
        sprintf(info, "ERROR: failed mapping tablet to area");
      } else {
        mapped = true;
      }

      sprintf(text, MAP_KEY, TABLET_NAME, 1, "c");
      system(text);
      sprintf(text, MAP_KEY, TABLET_NAME, 2, "k");
      system(text);
      sprintf(text, MAP_KEY, TABLET_NAME, 3, "e");
      system(text);
      sprintf(text, MAP_KEY, TABLET_STYLUS_NAME, 3, "z");
      system(text);
    }
    BeginDrawing();
    ClearBackground(bg_color);
    DrawText("Resize the window to cover \n"
             "whichever area you want to map to the tablet\n"
             "Press 'M' key when you are happy with the area\n",
             10, 100, 20, LIGHTGRAY);
    if (mapped) {
      DrawText("Mapped to the area!! \n", 10, 190, 25, GREEN);
    } else {
      DrawText(info, 10, 190, 25, RED);
    }
    EndDrawing();
  }

  CloseWindow();
  return 0;
}
