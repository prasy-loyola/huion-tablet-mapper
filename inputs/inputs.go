package inputs

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

const (
	INPUT_MAPPING_COORD_MATRIX = "transformation_matrix"
	INPUT_MAPPING_WINDOW       = "window"
)

type InputMappingType string

type InputConfig struct {
	Buttons     map[string]string  `json:"buttons"`
	CoordMatrix CoordinationMatrix `json:"coordMatrix"`
	WindowName  string             `json:"widowName"`
	Rotation    int                `json:"rotation"`
	MappingType InputMappingType   `json:"mappingType"`
}

func (input Input) MapButtons() error {

	for button, key := range input.Config.Buttons {
		setButtonCmd := exec.Command("xsetwacom", "--set", strconv.Itoa(input.Id),
			"Button", button, key)
		defer setButtonCmd.Wait()
		if out, err := setButtonCmd.CombinedOutput(); err != nil {
			log.Printf("ERROR: %s", out)
			return err
		}
	}
	return nil
}

func (input Input) MapToArea(m CoordinationMatrix) error {

	//xinput set-prop "<input-name>" --type=float "Coordinate Transformation Matrix" %f 0 %f 0 %f %f 0 0 1
	log.Printf("Coordination Matrix: %v", m)
	args := make([]string, 0)
	args = append(args, "set-prop")
	args = append(args, input.Name)
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

func GetCoordinateMatrix(angle int) CoordinationMatrix {
	switch angle {
	case 0:
		return rotationCoordMatrices[0]
	case 90:
		return rotationCoordMatrices[1]
	case 180:
		return rotationCoordMatrices[2]
	case 270:
		return rotationCoordMatrices[3]
	default:
		return rotationCoordMatrices[0]
	}
}

var rotationCoordMatrices []CoordinationMatrix = []CoordinationMatrix{
	{ // 0 degree
		{1.0, 0.0, 0.0},
		{0.0, 1.0, 0.0},
		{0.0, 0.0, 1.0},
	},
	{ // 90 degree
		{0.0, 1.0, 0.0},
		{-1.0, 0.0, 1.0},
		{0.0, 0.0, 1.0},
	},
	{ //180 degree
		{-1.0, 0.0, 1.0},
		{0.0, -1.0, 1.0},
		{0.0, 0.0, 1.0},
	},
	{ // 270 degree
		{0.0, -1.0, 1.0},
		{1.0, 0.0, 0.0},
		{0.0, 0.0, 1.0},
	},
}

type CoordMatrixRow [3]float32
type CoordinationMatrix [3]CoordMatrixRow

func (c CoordinationMatrix) MultiplyCoordMatrices(r CoordinationMatrix) CoordinationMatrix {

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

func GetInputs() ([]Input, error) {

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
				return nil, fmt.Errorf("Couldn't read inputs  for %s %w", input, err)
			}
			var id int
			if id, err = strconv.Atoi(strings.Split(string(output), "\n")[0]); err != nil {
				return nil, fmt.Errorf("Couldn't read id for input %s. %w", input, err)
			}
			tablets = append(tablets, Input{
				Id:       id,
				Name:     input,
				Selected: true,
			})
		}
	}
	xinputListCmd.Wait()
	return tablets, err
}

type Input struct {
	Id       int
	Name     string
	Selected bool
	Config   InputConfig
}
