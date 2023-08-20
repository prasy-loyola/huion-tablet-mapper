#include<stdio.h>
#include "raylib.h"
#include <stdlib.h>
#include <string.h>

int main(void) {

    SetConfigFlags(FLAG_WINDOW_RESIZABLE);
    SetTargetFPS(60);
    InitWindow(800, 450, "Huion Tablet mapper");
    Color bg_color = {
        .a = 255,
        .r = 0x18,
        .g = 0x18,
        .b = 0x18
    };
    Color text_color = {
        .a = 255,
        .r = 0x88,
        .g = 0x28,
        .b = 0x08
    };
    const char* TABLET_STYLUS_NAME = "HUION H420 Pen stylus";
    const char* TABLET_NAME = "HUION H420 Pad pad";
    char text[3000] = {0};
    while(!WindowShouldClose()) {
        if (IsKeyDown(KEY_M)) { 
            int curr_height = GetRenderHeight();
            int curr_width = GetRenderWidth();
            Vector2 position = GetWindowPosition();
            int monitor_id = GetCurrentMonitor();
            int screen_width = GetMonitorWidth(monitor_id);
            int screen_height =GetMonitorHeight(monitor_id);
            //c0 = touch_area_width / total_width
            float c0 = (float)curr_width / screen_width;
            //c2 = touch_area_height / total_height
            float c2 = (float)curr_height / screen_height;
            //c1 = touch_area_x_offset / total_width
            float c1 = position.x / screen_width;
            //c3 = touch_area_y_offset / total_height
            float c3 = position.y / screen_height;

            sprintf(text,"xinput set-prop \"%s\" "
                    "--type=float "
                    "\"Coordinate Transformation Matrix\" "
                    "%f 0 %f 0 %f %f 0 0 1",
                    TABLET_NAME,
                    c0, c1, c2, c3);
            printf("%s\n", text);
            system(text);
            sprintf(text,"xinput set-prop \"%s\" "
                    "--type=float "
                    "\"Coordinate Transformation Matrix\" "
                    "%f 0 %f 0 %f %f 0 0 1",
                    TABLET_STYLUS_NAME,
                    c0, c1, c2, c3);
            printf("%s\n", text);
            //TODO: handle the failure case
            system(text);

        }
        BeginDrawing();
            ClearBackground(bg_color);
            DrawText("Resize the window to cover \n"
                    "whichever area you want to map to the tablet\n"
                    "Press 'M' key when you are happy with the area\n",
                    10, 100, 20, LIGHTGRAY);
            if (text[0]) {
                DrawText("Mapped to the area!! \n", 10, 190, 25, text_color);
            }
            
        EndDrawing();
    }

    CloseWindow();
    return 0;
}
