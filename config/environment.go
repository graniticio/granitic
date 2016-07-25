package config

import "os"

func GraniticHome() string {
	return os.Getenv("GRANITIC_HOME")
}
