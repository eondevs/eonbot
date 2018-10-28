package pkg

import "fmt"

const (
	Name        = "EonBot"
	VersionCode = "v2.0.0"
	VersionName = "Arya"
)

func FullVersion() string {
	return fmt.Sprintf("%s %s '%s'", Name, VersionCode, VersionName)
}
