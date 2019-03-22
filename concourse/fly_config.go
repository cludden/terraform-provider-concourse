package concourse

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"

	yaml "gopkg.in/yaml.v2"
)

// FlyRc is a representation of the configuration file structure that is stored by the
// "fly" command line interface. (Usually to be found in ~/.flyrc)
type FlyRc struct {
	Filename string
	Targets  map[string]FlyRcTarget `yaml:"targets"`
}

// FlyRcTarget describes a target
type FlyRcTarget struct {
	API      string           `yaml:"api"`
	Team     string           `yaml:"team"`
	Insecure bool             `yaml:"insecure,omitempty"`
	Token    FlyRcTargetToken `yaml:"token"`
}

// FlyRcTargetToken describes a token
type FlyRcTargetToken struct {
	Type  string `yaml:"type"`
	Value string `yaml:"value"`
}

// ImportConfig reads in a `flyrc` file and returns a FlyRc struct
func (rc *FlyRc) ImportConfig() error {

	rc.setFlyRcLocation()

	b, err := rc.readFlyConfig()
	if err != nil {
		return err
	}
	return yaml.Unmarshal(b, rc)
}

func (rc *FlyRc) setFlyRcLocation() {
	fallback := ".flyrc" // If all else fails, we'll just return .flyrc in the current directory

	// Check if an ENV var has been set with a path
	// Todo: Find out if this is the correct ENV var, or if it fly even has one.
	if flyrc, ok := os.LookupEnv("FLYRC"); ok {
		if len(flyrc) > 0 {
			rc.Filename = flyrc
		}
		return
	}

	// Otherwise, return the default flyrc location
	cu, err := user.Current()
	if err != nil {
		rc.Filename = fallback
	}
	flyrc := fmt.Sprintf("%s/.flyrc", cu.HomeDir)
	rc.Filename = flyrc
}

// Get the bytes of the flyrc config based on the filepath given
func (rc *FlyRc) readFlyConfig() ([]byte, error) {
	if _, err := os.Stat(rc.Filename); err != nil {
		return nil, fmt.Errorf("unable to stat the flyrc file (%s): %v", rc.Filename, err)
	}
	return ioutil.ReadFile(rc.Filename)
}
