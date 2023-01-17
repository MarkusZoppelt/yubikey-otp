package devices

import (
	"bufio"
	"os/exec"
	"strings"
)

func GetDevices() (devices []string, err error) {
	out, err := exec.Command("ykman", "list").Output()
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		devices = append(devices, scanner.Text())
	}

	return devices, nil
}
