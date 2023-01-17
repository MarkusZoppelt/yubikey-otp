package accounts

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"
)

func GetOTPAccountsFromYubiKey(device string) (accounts []string, err error) {
	out, err := exec.Command("ykman", "--device", device, "oath", "accounts", "list").Output()
	if err != nil {
		fmt.Println("Cmd that failed: ", "ykman", "--device", device, "oath", "accounts", "list")
		return nil, err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		accounts = append(accounts, scanner.Text())
	}

	return accounts, nil
}
