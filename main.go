package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"sort"
	tm_config "tablet_mapper/config"
	tm_inputs "tablet_mapper/inputs"
	"tablet_mapper/windows"

	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	windowList := windows.GetWindowList()
	rl.SetTraceLogLevel(rl.LogNone)
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.SetConfigFlags(rl.FlagMsaa4xHint)

	args := os.Args
	climode := false
	if len(args) > 1 {
		log.Printf("INFO: using cli mode as arguments are passed")
		if args[1] == "-h" {
			fmt.Printf("Usage: \n %s <config-file-path>\n", args[0])
			return
		} else {
			climode = true
		}
	}
    if (climode) {
	    rl.SetConfigFlags(rl.FlagWindowHidden)
    }
	rl.InitWindow(800, 800, "Tablet Mapper")
	defer rl.CloseWindow()

	var inputs []tm_inputs.Input
	var err error

	var config tm_config.TabletMapperConfig

	var confPath string

	if climode {
		confPath = args[1]
	} else {
		if confPath, err = tm_config.GetDefaultConfpath(); err != nil {
			log.Printf("WARN: Couldn't load config %s", err.Error())
		}

	}
	if config, err = tm_config.ReadConfigFromFile(confPath); err != nil {
		log.Printf("WARN: Couldn't load config %s", err.Error())
		config = tm_config.TabletMapperConfig{}
	}
	if inputs, err = tm_inputs.GetInputs(); err != nil {
		log.Fatalf("ERROR: Couldn't read inputs %s", err.Error())
        inputs = make([]tm_inputs.Input, 0)
	}

    log.Printf("Window list %v", windowList)
    log.Printf("Input list %v", inputs)

	for i := 0; i < len(inputs); i++ {
        if config, ok := config[inputs[i].Name]; ok {
            inputs[i].Config = config
            input := inputs[i]
            log.Printf("Input config: %v", input.Config)
            if input.Config.MappingType == tm_inputs.INPUT_MAPPING_COORD_MATRIX {
                input.MapToArea(input.Config.CoordMatrix)
            } else if input.Config.MappingType == tm_inputs.INPUT_MAPPING_WINDOW {
                for _, window := range windowList {
                    if window.AppName == input.Config.WindowName {
                        log.Printf("Mapping to window %s", window)
                        coordMatrix := window.GetCoordMappingForWindow()
                        log.Printf("window coordinates %v", coordMatrix)
                        coordMatrix = coordMatrix.MultiplyCoordMatrices(tm_inputs.GetCoordinateMatrix(input.Config.Rotation))
                        log.Printf("transformed coordinates %v", coordMatrix)
                        input.MapToArea(coordMatrix)
                    }
                }
            }
        }
	}

	if climode {
		return
	}

	fontFilePath := ".temp.ttf"
	err = os.WriteFile(fontFilePath, FontAsBytes, fs.ModePerm)
	font := rl.LoadFontEx(fontFilePath, 30, nil)
	rl.SetTextureFilter(font.Texture, rl.FilterBilinear)
	defer rl.UnloadFont(font)

	rl.SetTargetFPS(60)

	type inputDialog struct {
		input   *tm_inputs.Input
		display bool
	}

	var selectedWindow int32
	windowEditMode := false
	rotate := 0
	rotateOptions := []int{0, 90, 180, 270}

	for !rl.WindowShouldClose() {
		if rl.IsWindowResized() {
			curr_width := rl.GetRenderWidth()
			new_height := int(float32(curr_width) * (float32(2.23) / float32(4.0)))
			rl.SetWindowSize(curr_width, new_height)
		}

		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)
		gui.SetStyle(gui.DEFAULT, gui.TEXT_SIZE, 17)
		gui.SetFont(font)

		var x float32 = 40.0
		var y float32 = 10.0
		for i := 0; i < len(inputs); i++ {
			y += 25.0
			inputs[i].Selected = gui.CheckBox(rl.NewRectangle(x, y, 20, 20), inputs[i].Name, inputs[i].Selected)

			keys := make([]string, 0, len(inputs[i].Config.Buttons))
			for k := range inputs[i].Config.Buttons {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, key := range keys {
				y += 25.0
				gui.Label(rl.NewRectangle(x+10, y, 200, 20), fmt.Sprintf("Button %s: '%s'", key, inputs[i].Config.Buttons[key]))
			}
		}
		y += 30.0

		gui.Label(rl.NewRectangle(x, y, 140, 30), "Rotation (degrees)")
		for i, value := range rotateOptions {
			selected := gui.Toggle(rl.NewRectangle(140+x+float32(i*50), y, 40, 30), fmt.Sprintf("%d", value), rotate == i)
			if selected {
				rotate = i
			}
		}
		y += 40

		if mapArea := gui.Button(rl.NewRectangle(x, y, 200, 40), "Map Current Area"); mapArea {
			for _, input := range inputs {
				if input.Selected {
					coordMatrix := windows.GetCoordMappingFromCurrentWindow()
					coordMatrix = coordMatrix.MultiplyCoordMatrices(tm_inputs.GetCoordinateMatrix(rotateOptions[rotate]))
					if err := input.MapToArea(coordMatrix); err != nil {
					}
					if err := input.MapButtons(); err != nil {
					}
				}
			}
		}

		y += 50
		options := ""
		first := true
		for _, w := range windowList {
			if !first {
				options += "\n"
			}
			first = false
			options += w.AppName
		}

		var x1, y1 = x, y
		var dropdown = func() {
			gui.Unlock()
			if gui.DropdownBox(rl.NewRectangle(x1, y1, 200, 40), options, &selectedWindow, windowEditMode) {
				windowEditMode = !windowEditMode
			}
		}

		if gui.Button(rl.NewRectangle(x+205, y, 40, 40), "(R)") {
			windowList = windows.GetWindowList()
		}

		if gui.Button(rl.NewRectangle(x+250, y, 200, 40), "Map to Window") {
			windowName := windowList[selectedWindow].AppName
			windowList = windows.GetWindowList()

            for _, window := range windowList {
                if window.AppName == windowName {
                    log.Printf("INFO: mapping to window %+v", window)
                    coordMatrix := window.GetCoordMappingForWindow()
                    coordMatrix = coordMatrix.MultiplyCoordMatrices(tm_inputs.GetCoordinateMatrix(rotateOptions[rotate]))
                    for i := 0; i < len(inputs); i++ {
                        if inputs[i].Selected {
                            if err := inputs[i].MapToArea(coordMatrix); err != nil {
                            }
                            inputs[i].Config.CoordMatrix = coordMatrix
                            inputs[i].Config.Rotation = rotateOptions[rotate]
                            inputs[i].Config.MappingType = tm_inputs.INPUT_MAPPING_WINDOW
                            inputs[i].Config.WindowName = window.AppName
                            config[inputs[i].Name] = inputs[i].Config
                            if err := inputs[i].MapButtons(); err != nil {
                            }
                        }
                    }
                }
            }
            
		}
		y += 50.0
		if gui.Button(rl.NewRectangle(x, y, 200, 40), "Load Config") {
			config, _ := tm_config.ReadConfigFromFile(confPath)
			for i := range inputs {
				input := &inputs[i]
				if input.Selected {
					input.Config = config[input.Name]
					if err := input.MapToArea(input.Config.CoordMatrix); err != nil {
					}
					if err := input.MapButtons(); err != nil {
					}
				}
			}
		}

		if gui.Button(rl.NewRectangle(x+250, y, 200, 40), "Save Current Config") {
			tm_config.WriteConfig(config)
		}

		dropdown()
		rl.EndDrawing()
	}

}
