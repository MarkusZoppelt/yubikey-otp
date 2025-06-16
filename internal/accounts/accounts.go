package accounts

import (
	"fmt"
	"runtime"

	"cunicu.li/go-iso7816/drivers/pcsc"
	"cunicu.li/go-ykoath/v2"
	"github.com/ebfe/scard"
)

func GetOTPAccountsFromYubiKey(device string) (accounts []string, err error) {
	fmt.Printf("Connecting to device: %s\n", device)
	
	ctx, err := scard.EstablishContext()
	if err != nil {
		return nil, fmt.Errorf("failed to establish PC/SC context: %v", err)
	}
	defer ctx.Release()

	card, err := pcsc.NewCard(ctx, device, true)
	if err != nil {
		fmt.Printf("Failed to connect to card: %v\n", err)
		return nil, fmt.Errorf("failed to connect to YubiKey: %v%s", err, getMacOSAccountsHelp())
	}
	defer card.Close()

	ykoathCard, err := ykoath.NewCard(card)
	if err != nil {
		fmt.Printf("Failed to create OATH card: %v\n", err)
		return nil, fmt.Errorf("failed to initialize OATH card: %v%s", err, getMacOSAccountsHelp())
	}
	defer ykoathCard.Close()

	selectResp, err := ykoathCard.Select()
	if err != nil {
		fmt.Printf("Failed to select OATH applet: %v\n", err)
		return nil, fmt.Errorf("failed to select OATH applet: %v%s", err, getMacOSAccountsHelp())
	}

	// Check if OATH applet requires authentication
	if selectResp.Challenge != nil {
		return nil, fmt.Errorf(`YubiKey OATH applet requires authentication, but the current go-ykoath library does not support password authentication properly.

To use this tool, you have two options:

1. Remove the OATH password temporarily:
   ykman oath access change --remove

2. Or use ykman directly for now:
   ykman oath accounts list
   ykman oath accounts code <account-name>

This is a known limitation with the current pure Go OATH implementation.
We will need to either:
- Find/implement a working OATH authentication library
- Use ykman as a subprocess
- Implement the OATH protocol from scratch

Challenge: %x
Device: %s`, selectResp.Challenge, string(selectResp.Name))
	}

	names, err := ykoathCard.List()
	if err != nil {
		if runtime.GOOS == "darwin" {
			fmt.Printf("Failed to list OATH accounts: %v\n", err)
		}
		return nil, fmt.Errorf("failed to list OATH accounts: %v%s", err, getMacOSAccountsHelp())
	}

	if runtime.GOOS == "darwin" {
		fmt.Printf("Found %d OATH accounts\n", len(names))
	}

	for _, name := range names {
		accounts = append(accounts, name.Name)
	}

	if len(accounts) == 0 {
		return nil, fmt.Errorf("no OATH accounts found on YubiKey%s", getMacOSAccountsHelp())
	}

	return accounts, nil
}

func getMacOSAccountsHelp() string {
	if runtime.GOOS != "darwin" {
		return ""
	}
	
	return `

macOS YubiKey OATH Troubleshooting:
1. Try reconnecting the YubiKey
2. Ensure no other applications are using the YubiKey
3. Check if yubikey-agent is running: killall yubikey-agent
4. Restart PC/SC daemon:
   sudo launchctl stop com.apple.securityd
   sudo launchctl start com.apple.securityd`
}
