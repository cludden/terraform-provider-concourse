package concourse

import (
	"os"
	"testing"
)

func TestFlyRC_ImportConfigFromEnv(t *testing.T) {
	flyrc := "./testdata/flyrc.yml"
	n := 2

	os.Setenv("FLYRC", flyrc)

	var rc FlyRc

	err := rc.ImportConfig()
	if err != nil {
		t.Fatalf("failed to import config from %s: %s", flyrc, err)
	}

	if rc.Filename != flyrc {
		t.Fatalf("filename stored in FlyRc (%s) is not the same as the test file (%s)", rc.Filename, flyrc)
	}

	if len(rc.Targets) != n {
		t.Fatalf("expected %d targets, but counted %d", n, len(rc.Targets))
	}

}
