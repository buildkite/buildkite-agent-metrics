package version

import "fmt"

// Version the library version number
const Version = "2.0.0"

// BuildNumber is the buildkite build number set with ldflags
var BuildNumber string = "dev"

// VerisonString returns the version and the build number
func VersionString() string {
	if BuildNumber == "dev" || BuildNumber == "" {
		return fmt.Sprintf("%s Development Build", Version)
	}
	return fmt.Sprintf("%s, Build %s", Version, BuildNumber)
}
