package assets

import "testing"

func TestLoadTomlFile(t *testing.T) {
	machines, err := LoadTomlFile("../etc/machines.conf")
	if err != nil {
		t.Fatal(err)
	}
	for _, machine := range machines {
		t.Logf("machine: %v\r\n", machine)
	}
}
