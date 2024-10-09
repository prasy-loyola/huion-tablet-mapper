package windows

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"tablet_mapper/inputs"
)

type Window struct {
	Id          string
	DesktopId   int
	Xoffset     int
	Yoffset     int
	Width       int
	Height      int
	MachineName string
	Title       string
	AppName     string
}

func GetWindowList() []Window {

	cmd := exec.Command("wmctrl", "-l", "-G")
	defer cmd.Wait()
	var out []byte
	var err error
	if out, err = cmd.CombinedOutput(); err != nil {
		log.Printf("ERROR: %s", out)
	}

	windowList := make([]Window, 0)
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
		w := Window{}
		id, rest := readNextWord(strings.TrimSpace(rest))
		desktopId, rest := readNextWord(strings.TrimSpace(rest))
		xoffset, rest := readNextWord(strings.TrimSpace(rest))
		yoffset, rest := readNextWord(strings.TrimSpace(rest))
		width, rest := readNextWord(strings.TrimSpace(rest))
		height, rest := readNextWord(strings.TrimSpace(rest))
		machineName, rest := readNextWord(strings.TrimSpace(rest))
		title := strings.TrimSpace(rest)

		w.Id = id
		w.DesktopId, _ = strconv.Atoi(desktopId)
		w.Xoffset, _ = strconv.Atoi(xoffset)
		w.Yoffset, _ = strconv.Atoi(yoffset)
		w.Width, _ = strconv.Atoi(width)
		w.Height, _ = strconv.Atoi(height)
		w.MachineName = machineName
		w.Title = title
		chunks := strings.Split(title, "-")
		w.AppName = chunks[len(chunks)-1]
		windowList = append(windowList, w)
	}

	return windowList
}
func (win Window) GetCoordMappingForWindow() inputs.CoordinationMatrix {
	x := win.Xoffset
	y := win.Yoffset
	w := win.Width
	h := win.Height
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

	return inputs.CoordinationMatrix{{c0, 0.0, c1}, {0.0, c2, c3}, {0.0, 0.0, 1.0}}
}

func GetCoordMappingFromCurrentWindow() inputs.CoordinationMatrix {
	curr_height := rl.GetRenderHeight()
	curr_width := rl.GetRenderWidth()
	window_pos := rl.GetWindowPosition()
	window := Window{
		Xoffset: int(window_pos.X),
		Yoffset: int(window_pos.Y),
		Width:   curr_width,
		Height:  curr_height,
	}
	return window.GetCoordMappingForWindow()
}
