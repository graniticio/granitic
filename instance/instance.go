package instance

import "os"

const FrameworkPrefix = "grnc"

func ExitError() {
	os.Exit(1)
}

func ExitNormal() {
	os.Exit(0)
}
