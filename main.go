package main

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"

	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Input struct {
	id       int
	name     string
	selected bool
}

func (input Input) mapButton(button int, key string) error {
    setButtonCmd := exec.Command("xsetwacom", "--set", strconv.Itoa(input.id), 
    "Button", strconv.Itoa(button), "key " + key)
    defer setButtonCmd.Wait()
    if out, err := setButtonCmd.CombinedOutput(); err != nil {
        log.Printf("ERROR: %s", out)
        return err
    } else {
        log.Printf("INFO: %s", out)
    }
    return nil
}

func (input Input) mapToArea() error {
	monitor_count := rl.GetMonitorCount()
	screen_width := 0
	screen_height := 0
	for i := 0; i < monitor_count; i++ {
		monitor_pos := rl.GetMonitorPosition(i)
		if screen_width < rl.GetMonitorWidth(i)+int(monitor_pos.X) {
			screen_width = rl.GetMonitorWidth(i) + int(monitor_pos.X)
		}
		if screen_height < rl.GetMonitorHeight(i)+int(monitor_pos.Y) {
			screen_height = rl.GetMonitorHeight(i) + int(monitor_pos.Y)
		}
	}

	curr_height := rl.GetRenderHeight()
	curr_width := rl.GetRenderWidth()
	window_pos := rl.GetWindowPosition()
	// c0 = touch_area_width / total_width
	c0 := float32(curr_width) / float32(screen_width)
	// c2 = touch_area_height / total_height
	c2 := float32(curr_height) / float32(screen_height)
	// c1 = touch_area_x_offset / total_width
	c1 := window_pos.X / float32(screen_width)
	// c3 = touch_area_y_offset / total_height
	c3 := window_pos.Y / float32(screen_height)
	//xinput set-prop "<input-name>" --type=float "Coordinate Transformation Matrix" %f 0 %f 0 %f %f 0 0 1

	c := strings.Split(fmt.Sprintf("%f %f %f %f", c0, c1, c2, c3), " ")

	setCoordMatrixCmd := exec.Command("xinput", "set-prop", input.name, "--type=float",
		"Coordinate Transformation Matrix", c[0], "0", c[1], "0", c[2], c[3], "0", "0", "1")
	defer setCoordMatrixCmd.Wait()
	if _, err := setCoordMatrixCmd.Output(); err != nil {
		return fmt.Errorf("Couldn't map inputs %w", err)
	}

	return nil
}

func getInputs() ([]Input, error) {

	xinputListCmd := exec.Command("xinput", "--list", "--name-only")
	defer xinputListCmd.Wait()
	var output []byte
	var err error
	if output, err = xinputListCmd.Output(); err != nil {
		log.Printf("ERROR: error while reading output %s", err.Error())
		return nil, fmt.Errorf("Couldn't read inputs %w", err)
	}
	inputs := strings.Split(string(output), "\n")
	tablets := make([]Input, 0)
	for _, input := range inputs {
		if strings.Contains(input, " stylus") || strings.Contains(input, " pad") {
			xinputFindIdCmd := exec.Command("xinput", "--list", "--id-only", input)
			defer xinputListCmd.Wait()
			if output, err = xinputFindIdCmd.Output(); err != nil {
				return nil, fmt.Errorf("Couldn't read inputs %w", err)
			}
			var id int
			if id, err = strconv.Atoi(strings.Split(string(output), "\n")[0]); err != nil {
				return nil, fmt.Errorf("Couldn't read id for input %s. %w", input, err)
			}
			tablets = append(tablets, Input{
				id:   id,
				name: input,
			})
		}
	}
	log.Printf("INFO: tablets '%v'", tablets)
	xinputListCmd.Wait()
	return tablets, err
}

func main() {
	var inputs []Input
	var err error
	if inputs, err = getInputs(); err != nil {
		log.Fatalf("ERROR: Couldn't read inputs %s", err.Error())
	}
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(800, 800, "Tablet Mapper")
	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

	for !rl.WindowShouldClose() {
		if rl.IsWindowResized() {
			curr_width := rl.GetRenderWidth()
			new_height := int(float32(curr_width) * (float32(2.23) / float32(4.0)))
			rl.SetWindowSize(curr_width, new_height)
		}
		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)

		gui.SetStyle(gui.DEFAULT, gui.TEXT_SIZE, 17)
		var x float32 = 10.0
		var y float32 = 10.0
		for i := 0; i < len(inputs); i++ {
			y += 20.0
			inputs[i].selected = gui.CheckBox(rl.NewRectangle(x, y, 20, 20), inputs[i].name, inputs[i].selected)
		}
		y += 30.0
        if mapArea := gui.Button(rl.NewRectangle(x, y, 200, 40), "Map Current Area"); mapArea {
			for i := 0; i < len(inputs); i++ {
				if inputs[i].selected {
					if err := inputs[i].mapToArea(); err != nil {
					}
					if err := inputs[i].mapButton(1, "q"); err != nil {
					}
				}
			}
		}
		rl.EndDrawing()
	}
}
