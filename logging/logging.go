package logging

import (
	"log"
	"os"
)

var WarnLog log.Logger = *log.New(os.Stdout, "[WARN]", log.LstdFlags)
