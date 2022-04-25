package version

import "fmt"

var (
	// GitCommit is the git commit that was compiled.
	GitCommit string

	// Version is the main version at the moment.
	Version = "0.1.1"

	// VersionPrerelease is a marker for the version.
	VersionPrerelease = ""
)

// GetVersion returns a string representation of the version
func GetVersion() string {
	version := Version

	release := VersionPrerelease
	if release != "" {
		version += fmt.Sprintf("-%s", release)

		if GitCommit != "" {
			version += fmt.Sprintf(" (%s)", GitCommit)
		}
	}

	return version
}
