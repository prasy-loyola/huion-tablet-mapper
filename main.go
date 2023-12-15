package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path"
	"sort"
	"strconv"
	"strings"

	gui "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type Input struct {
	id       int
	name     string
	selected bool
	config   InputConfig
}

type InputConfig struct {
	Buttons     map[string]string `json:"buttons"`
	CoordMatrix [4]float32        `json:"coordMatrix"`
}

const configFileName = ".tablet-mapper.conf"

func readConfig() (TabletMapperConfig, error) {
	if user, err := user.Current(); err != nil {
		log.Printf("INFO: couldn't read current user %s ", err.Error())
	} else {
		confPath := path.Join(user.HomeDir, configFileName)
		log.Printf("INFO: reading from file %s", confPath)
		if file, err := os.Open(confPath); err != nil {
			log.Printf("INFO: couldn't read from config file %s. %s", confPath, err.Error())
		} else {
			defer file.Close()
			if buf, err := io.ReadAll(file); err != nil {
				log.Printf("INFO: read config %s", err.Error())
			} else {
				var config TabletMapperConfig
				if err = json.Unmarshal(buf, &config); err != nil {
					log.Printf("INFO: couldn't read config file %s. %s", confPath, err.Error())
				} else {
					log.Printf("INFO: read config %v", config)
					return config, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("ERROR: Couldn't read config file %s", configFileName)

}

func writeConfig(config TabletMapperConfig) {
	if user, err := user.Current(); err != nil {
		log.Printf("INFO: couldn't read current user %s ", err.Error())
	} else {
		confPath := path.Join(user.HomeDir, configFileName)
		log.Printf("INFO: writing to file %s", confPath)
		if file, err := os.Create(confPath); err != nil {
			log.Printf("INFO: couldn't write to config file %s. %s", confPath, err.Error())
		} else {
			defer file.Close()
			buf, _ := json.MarshalIndent(config, "", "  ")
			if _, err = file.Write(buf); err != nil {
				log.Printf("INFO: couldn't write to config file %s. %s", confPath, err.Error())
			}
            //log.Printf("INFO: wrote config: %s",buf )
			file.Sync()
		}
	}
}

type TabletMapperConfig map[string]InputConfig

func (input Input) mapButtons() error {

	for button, key := range input.config.Buttons {
		setButtonCmd := exec.Command("xsetwacom", "--set", strconv.Itoa(input.id),
			"Button", button, "key "+key)
		defer setButtonCmd.Wait()
		if out, err := setButtonCmd.CombinedOutput(); err != nil {
			log.Printf("ERROR: %s", out)
			return err
		}
	}
	return nil
}

func (input Input) mapToArea(matrix [4]float32) error {

	c := strings.Split(fmt.Sprintf("%f %f %f %f", matrix[0], matrix[1], matrix[2], matrix[3]), " ")
	setCoordMatrixCmd := exec.Command("xinput", "set-prop", input.name, "--type=float",
		"Coordinate Transformation Matrix", c[0], "0", c[1], "0", c[2], c[3], "0", "0", "1")
	defer setCoordMatrixCmd.Wait()
	if _, err := setCoordMatrixCmd.Output(); err != nil {
		return fmt.Errorf("Couldn't map inputs %w", err)
	}

	return nil
}

func getCoordMappingFromCurrentWindow() [4]float32 {
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

	return [4]float32{c0, c1, c2, c3}
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
				id:       id,
				name:     input,
				selected: true,
			})
		}
	}
	xinputListCmd.Wait()
	return tablets, err
}

func main() {
	var inputs []Input
	var err error

	var config TabletMapperConfig

	if config, err = readConfig(); err != nil {
		log.Printf("WARN: Couldn't load config %s", err.Error())
	}
	if inputs, err = getInputs(); err != nil {
		log.Fatalf("ERROR: Couldn't read inputs %s", err.Error())
	}

	for i := 0; i < len(inputs); i++ {
		inputs[i].config = config[inputs[i].name]
	}

	log.Printf("INFO: inputs '%v'", inputs)
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(800, 800, "Tablet Mapper")
	defer rl.CloseWindow()

	font := rl.LoadFont("fonts/pixantiqua.ttf")
	defer rl.UnloadFont(font)

	rl.SetTargetFPS(60)

	type inputDialog struct {
		input   *Input
		display bool
	}

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
			inputs[i].selected = gui.CheckBox(rl.NewRectangle(x, y, 20, 20), inputs[i].name, inputs[i].selected)

            keys := make([]string, 0, len(inputs[i].config.Buttons))
            for k := range inputs[i].config.Buttons {
                keys = append(keys, k)
            }
            sort.Strings(keys)

            for _, key := range keys {
			    y += 25.0
                gui.Label(rl.NewRectangle(x + 10, y, 100, 20), fmt.Sprintf("Button %s: '%s'", key, inputs[i].config.Buttons[key] ))
            }
		}
		y += 30.0

		if loadConfig := gui.Button(rl.NewRectangle(x, y, 200, 40), "Load Config"); loadConfig {
            config, _ := readConfig()
            for i, _ := range inputs {
                input := &inputs[i]
				if input.selected {
                    input.config = config[input.name]
					if err := input.mapToArea(input.config.CoordMatrix); err != nil {
					}
					if err := input.mapButtons(); err != nil {
					}
				}
			}
		}
		y += 50.0

		if mapArea := gui.Button(rl.NewRectangle(x, y, 200, 40), "Map Current Area"); mapArea {
            for _, input := range inputs {
				if input.selected {
					if err := input.mapToArea(getCoordMappingFromCurrentWindow()); err != nil {
					}
					if err := input.mapButtons(); err != nil {
					}
				}
			}
		}
		y += 50

		if saveConfig := gui.Button(rl.NewRectangle(x, y, 200, 40), "Save Current Config"); saveConfig {
			coordMatrix := getCoordMappingFromCurrentWindow()
			for i := 0; i < len(inputs); i++ {
				if inputs[i].selected {
					if err := inputs[i].mapToArea(coordMatrix); err != nil {
					}
                    inputs[i].config.CoordMatrix = coordMatrix
					config[inputs[i].name] = inputs[i].config
				}
			}
			writeConfig(config)
		}
		rl.EndDrawing()
	}
}
