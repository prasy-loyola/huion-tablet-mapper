#include "raylib.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int main(void) {

  Color bg_color = {.a = 255, .r = 0x18, .g = 0x18, .b = 0x18};
  const char *TABLET_STYLUS_NAME = "HUION H420 Pen stylus";
  const char *TABLET_NAME = "HUION H420 Pad pad";

  SetConfigFlags(FLAG_WINDOW_RESIZABLE);
  SetTargetFPS(60);
  InitWindow(800, 450, "Huion Tablet mapper");
  char text[3000] = {0};
  char info[3000] = {0};
  while (!WindowShouldClose()) {
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

      sprintf(text,
              "xinput set-prop \"%s\" "
              "--type=float "
              "\"Coordinate Transformation Matrix\" "
              "%f 0 %f 0 %f %f 0 0 1",
              TABLET_NAME, c0, c1, c2, c3);
      printf("%s\n", text);
      int result = system(text);
      if (result != 0) {
        DrawText("Error", 10, 190, 25, RED);
      }

      sprintf(text,
              "xinput set-prop \"%s\" "
              "--type=float "
              "\"Coordinate Transformation Matrix\" "
              "%f 0 %f 0 %f %f 0 0 1",
              TABLET_STYLUS_NAME, c0, c1, c2, c3);
      printf("%s\n", text);
      // TODO: handle the failure case
      system(text);
    }
    BeginDrawing();
    ClearBackground(bg_color);
    DrawText("Resize the window to cover \n"
             "whichever area you want to map to the tablet\n"
             "Press 'M' key when you are happy with the area\n",
             10, 100, 20, LIGHTGRAY);
    if (text[0]) {
        DrawText("Mapped to the area!! \n", 10, 190, 25, GREEN);
    }

    EndDrawing();
  }

  CloseWindow();
  return 0;
}
