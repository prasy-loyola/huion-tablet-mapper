package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
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
	Buttons     map[string]string  `json:"buttons"`
	CoordMatrix CoordinationMatrix `json:"coordMatrix"`
}

const configFileName = ".tablet-mapper.conf"

func readConfigFromFile(confPath string) (TabletMapperConfig, error) {
	if file, err := os.Open(confPath); err != nil {
		log.Printf("ERROR: couldn't read from config file '%s'. %s", confPath, err.Error())
	} else {
		defer file.Close()
		if buf, err := io.ReadAll(file); err != nil {
			log.Printf("ERROR: read config %s", err.Error())
		} else {
			var config TabletMapperConfig
			if err = json.Unmarshal(buf, &config); err != nil {
				log.Printf("ERROR: couldn't read config file '%s'. %s", confPath, err.Error())
			} else {
				//log.Printf("INFO: read config %v", config)
				return config, nil
			}
		}
	}
	return nil, fmt.Errorf("ERROR: Couldn't read config file '%s'", confPath)

}

func getDefaultConfpath() (string, error) {
	if user, err := user.Current(); err != nil {
		log.Printf("ERROR: couldn't read current user %s ", err.Error())
	} else {
		confPath := path.Join(user.HomeDir, configFileName)
		log.Printf("INFO: reading from file %s", confPath)
		return confPath, nil
	}
	return "", fmt.Errorf("ERROR: Couldn't get default config file %s", configFileName)

}

func writeConfig(config TabletMapperConfig) {
	if user, err := user.Current(); err != nil {
		log.Printf("ERROR: couldn't read current user %s ", err.Error())
	} else {
		confPath := path.Join(user.HomeDir, configFileName)
		log.Printf("INFO: writing to file %s", confPath)
		if file, err := os.Create(confPath); err != nil {
			log.Printf("ERROR: couldn't write to config file %s. %s", confPath, err.Error())
		} else {
			defer file.Close()
			buf, _ := json.MarshalIndent(config, "", "  ")
			if _, err = file.Write(buf); err != nil {
				log.Printf("ERROR: couldn't write to config file %s. %s", confPath, err.Error())
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
			"Button", button, key)
		defer setButtonCmd.Wait()
		if out, err := setButtonCmd.CombinedOutput(); err != nil {
			log.Printf("ERROR: %s", out)
			return err
		}
	}
	return nil
}

func (input Input) mapToArea(m CoordinationMatrix) error {

	//xinput set-prop "<input-name>" --type=float "Coordinate Transformation Matrix" %f 0 %f 0 %f %f 0 0 1
	args := make([]string, 0)
	args = append(args, "set-prop")
	args = append(args, input.name)
	args = append(args, "--type=float")
	args = append(args, "Coordinate Transformation Matrix")

	for _, row := range m {
		for _, val := range row {
			args = append(args, fmt.Sprintf("%f", val))
		}
	}
	log.Printf("INFO: area %+v", args)

	setCoordMatrixCmd := exec.Command("xinput", args...)
	defer setCoordMatrixCmd.Wait()
	if _, err := setCoordMatrixCmd.Output(); err != nil {
		return fmt.Errorf("Couldn't map inputs %w", err)
	}

	return nil
}

type CoordMatrixRow [3]float32
type CoordinationMatrix [3]CoordMatrixRow

func multiplyCoordMatrices(c, r CoordinationMatrix) CoordinationMatrix {

	result := CoordinationMatrix{}
    for ridx, row := range c {
        for j := 0; j < 3; j++ {
            sum := float32(0)
            for i := 0; i < 3; i++ {
                sum += row[i] * r[i][j]
            }
            result[ridx][j] = sum
        }
    }

	return result
}

func (c CoordinationMatrix) rotateLeft90() CoordinationMatrix {

	// tablet right 90 rotation matrix - screen left 90 rotation 
	r := CoordinationMatrix{
		{0.0, 1.0, 0.0},
		{-1.0, 0.0, 1.0},
		{0.0, 0.0, 1.0},
	}
	return multiplyCoordMatrices(c, r)
}

func (c CoordinationMatrix) rotateRight90() CoordinationMatrix {

	// tablet left 90 rotation matrix - screen right 90 rotation 
	r := CoordinationMatrix{
		{0.0, -1.0, 1.0},
		{1.0, 0.0, 0.0},
		{0.0, 0.0, 1.0},
	}
	return multiplyCoordMatrices(c, r)
}

func getCoordMappingForWindow(x, y, w, h int) CoordinationMatrix {
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

	window_posX := float32(x)
	window_posY := float32(y)
	curr_width := float32(w)
	curr_height := float32(h)

	// c0 = touch_area_width / total_width
	c0 := curr_width / float32(screen_width)
	// c2 = touch_area_height / total_height
	c2 := curr_height / float32(screen_height)
	// c1 = touch_area_x_offset / total_width
	c1 := window_posX / float32(screen_width)
	// c3 = touch_area_y_offset / total_height
	c3 := window_posY / float32(screen_height)

	return CoordinationMatrix{{c0, 0.0, c1}, {0.0, c2, c3}, {0.0, 0.0, 1.0}}
}

func getCoordMappingFromCurrentWindow() CoordinationMatrix {
	curr_height := rl.GetRenderHeight()
	curr_width := rl.GetRenderWidth()
	window_pos := rl.GetWindowPosition()
	return getCoordMappingForWindow(int(window_pos.X), int(window_pos.Y), curr_width, curr_height)
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
		if strings.Contains(input, " stylus") || strings.Contains(input, " pad") || strings.Contains(input, "Tablet") {
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

type window struct {
	id          string
	desktopId   int
	xoffset     int
	yoffset     int
	width       int
	height      int
	machineName string
	title       string
	appName     string
}

func getWindowList() []window {

	cmd := exec.Command("wmctrl", "-l", "-G")
	defer cmd.Wait()
	var out []byte
	var err error
	if out, err = cmd.CombinedOutput(); err != nil {
		log.Printf("ERROR: %s", out)
	}

	windowList := make([]window, 0)
	readNextWord := func(text string) (string, string) {
		word := make([]byte, 0)
		i := 0
		var b byte
		for i, b = range []byte(text) {
			if b == ' ' {
				break
			}
			word = append(word, b)
		}
		return string(word), string(text[i:])
	}

	for _, line := range strings.Split(string(out), "\n") {
		if strings.TrimSpace(line) == "" {
			break
		}
		rest := line
		w := window{}
		id, rest := readNextWord(strings.TrimSpace(rest))
		desktopId, rest := readNextWord(strings.TrimSpace(rest))
		xoffset, rest := readNextWord(strings.TrimSpace(rest))
		yoffset, rest := readNextWord(strings.TrimSpace(rest))
		width, rest := readNextWord(strings.TrimSpace(rest))
		height, rest := readNextWord(strings.TrimSpace(rest))
		machineName, rest := readNextWord(strings.TrimSpace(rest))
		title := strings.TrimSpace(rest)

		w.id = id
		w.desktopId, _ = strconv.Atoi(desktopId)
		w.xoffset, _ = strconv.Atoi(xoffset)
		w.yoffset, _ = strconv.Atoi(yoffset)
		w.width, _ = strconv.Atoi(width)
		w.height, _ = strconv.Atoi(height)
		w.machineName = machineName
		w.title = title
		chunks := strings.Split(title, " ")
		w.appName = chunks[len(chunks)-1]
		windowList = append(windowList, w)
	}

	return windowList
}
func main() {
	windowList := getWindowList()

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

	var inputs []Input
	var err error

	var config TabletMapperConfig

	var confPath string

	if climode {
		confPath = args[1]
	} else {
		if confPath, err = getDefaultConfpath(); err != nil {
			log.Printf("WARN: Couldn't load config %s", err.Error())
		}

	}
	if config, err = readConfigFromFile(confPath); err != nil {
		log.Printf("WARN: Couldn't load config %s", err.Error())
		config = TabletMapperConfig{}
	}
	if inputs, err = getInputs(); err != nil {
		log.Fatalf("ERROR: Couldn't read inputs %s", err.Error())
	}

	for i := 0; i < len(inputs); i++ {
		inputs[i].config = config[inputs[i].name]
	}

	if climode {
		return
	}

	rl.SetTraceLogLevel(rl.LogNone)
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.SetConfigFlags(rl.FlagMsaa4xHint)
	rl.InitWindow(800, 800, "Tablet Mapper")
	defer rl.CloseWindow()
	fontFilePath := ".temp.ttf"
	err = os.WriteFile(fontFilePath, FontAsBytes, fs.ModePerm)
	font := rl.LoadFontEx(fontFilePath, 30, nil)
	rl.SetTextureFilter(font.Texture, rl.FilterBilinear)
	defer rl.UnloadFont(font)

	rl.SetTargetFPS(60)

	type inputDialog struct {
		input   *Input
		display bool
	}

	var selectedWindow int32
	windowEditMode := false
	rotate := false

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
				gui.Label(rl.NewRectangle(x+10, y, 200, 20), fmt.Sprintf("Button %s: '%s'", key, inputs[i].config.Buttons[key]))
			}
		}
		y += 30.0

		rotate = gui.CheckBox(rl.NewRectangle(x, y, 20, 20), "Rotate Right 90deg", rotate)
		y += 40

		if mapArea := gui.Button(rl.NewRectangle(x, y, 200, 40), "Map Current Area"); mapArea {
			for _, input := range inputs {
				if input.selected {
					coordMatrix := getCoordMappingFromCurrentWindow()
					if rotate {
						coordMatrix = coordMatrix.rotateRight90()
					}
					if err := input.mapToArea(coordMatrix); err != nil {
					}
					if err := input.mapButtons(); err != nil {
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
			options += w.appName
		}

		gui.Unlock()
		if gui.DropdownBox(rl.NewRectangle(x, y, 200, 40), options, &selectedWindow, windowEditMode) {
			windowEditMode = !windowEditMode
		}

		if gui.Button(rl.NewRectangle(x+205, y, 40, 40), "(R)") {
			windowList = getWindowList()
		}

		if gui.Button(rl.NewRectangle(x+250, y, 200, 40), "Map to Window") {
			window := windowList[selectedWindow]
			log.Printf("INFO: mapping to window %+v", window)
			coordMatrix := getCoordMappingForWindow(window.xoffset, window.yoffset, window.width, window.height)
			if rotate {
				coordMatrix = coordMatrix.rotateLeft90()
			}
			for i := 0; i < len(inputs); i++ {
				if inputs[i].selected {
					if err := inputs[i].mapToArea(coordMatrix); err != nil {
					}
					inputs[i].config.CoordMatrix = coordMatrix
					config[inputs[i].name] = inputs[i].config
					if err := inputs[i].mapButtons(); err != nil {
					}
				}
			}
		}
		y += 50.0
		if gui.Button(rl.NewRectangle(x, y, 200, 40), "Load Config") {
			config, _ := readConfigFromFile(confPath)
			for i := range inputs {
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

		if gui.Button(rl.NewRectangle(x+250, y, 200, 40), "Save Current Config") {
			writeConfig(config)
		}

		rl.EndDrawing()
	}

}
