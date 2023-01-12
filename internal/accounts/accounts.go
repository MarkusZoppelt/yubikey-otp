package accounts

import (
	"bufio"
	"os/exec"
	"strings"
)

func GetOTPAccountsFromYubiKey() (accounts []string, err error) {
	out, err := exec.Command("ykman", "oath", "accounts", "list").Output()
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		accounts = append(accounts, scanner.Text())
	}

	return accounts, nil
}
