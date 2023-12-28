package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"
)

func main() {

	filepath := "../fonts/roboto/Roboto-Medium.ttf"
	var content []byte
	var err error
	if content, err = os.ReadFile(filepath); err != nil {
		log.Fatalf("ERROR: Couldn't read file %s", filepath)
	}

	builder := strings.Builder{}
	builder.WriteString(`
    package main
    var FontAsBytes []byte = []byte {
    `,
	)

	for i, b := range content {
		builder.WriteString(fmt.Sprintf("%d,", uint8(b))) 
		if i > 1 && i%100 == 0 {
			builder.WriteString("\n")
		}
	}

    builder.WriteString("}")

	os.WriteFile("font.go", []byte(builder.String()), fs.ModePerm)

}
