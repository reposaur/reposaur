package build

import (
	"runtime/debug"
)

// Version is dynamically set by the toolchain.
var Version = "DEV"

func init() {
	if Version == "DEV" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "(devel)" {
			Version = info.Main.Version
		}
	}
}
