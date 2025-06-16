package devices

import (
	"fmt"
	"runtime"
	"strings"

	"cunicu.li/go-iso7816/drivers/pcsc"
	"cunicu.li/go-ykoath/v2"
	"github.com/ebfe/scard"
)

func GetDevices() (devices []string, err error) {
	ctx, err := scard.EstablishContext()
	if err != nil {
		return nil, enhanceError(err, "Failed to establish PC/SC context")
	}
	defer ctx.Release()

	readers, err := ctx.ListReaders()
	if err != nil {
		return nil, enhanceError(err, "Failed to list smart card readers")
	}

	if len(readers) == 0 {
		return nil, fmt.Errorf("no smart card readers found%s", getMacOSHelp())
	}

	// Debug info for macOS
	if runtime.GOOS == "darwin" {
		fmt.Printf("Found %d readers: %v\n", len(readers), readers)
	}

	for _, reader := range readers {
		if runtime.GOOS == "darwin" {
			fmt.Printf("Trying reader: %s\n", reader)
		}
		
		card, err := pcsc.NewCard(ctx, reader, true)
		if err != nil {
			if runtime.GOOS == "darwin" {
				fmt.Printf("  Failed to connect to card: %v\n", err)
			}
			continue
		}

		ykoathCard, err := ykoath.NewCard(card)
		if err != nil {
			if runtime.GOOS == "darwin" {
				fmt.Printf("  Failed to create OATH card: %v\n", err)
			}
			card.Close()
			continue
		}

		_, err = ykoathCard.Select()
		if err != nil {
			if runtime.GOOS == "darwin" {
				fmt.Printf("  Failed to select OATH applet: %v\n", err)
			}
			ykoathCard.Close()
			continue
		}

		if runtime.GOOS == "darwin" {
			fmt.Printf("  Successfully found YubiKey OATH device\n")
		}
		devices = append(devices, fmt.Sprintf("%s", reader))
		ykoathCard.Close()
	}

	if len(devices) == 0 {
		return nil, fmt.Errorf("no YubiKey OATH devices found%s", getMacOSDebugHelp())
	}

	return devices, nil
}

func enhanceError(err error, context string) error {
	if err == nil {
		return nil
	}
	
	baseMsg := fmt.Sprintf("%s: %v", context, err)
	
	if runtime.GOOS == "darwin" {
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "cannot find a smart card reader") || 
		   strings.Contains(errStr, "no readers available") ||
		   strings.Contains(errStr, "scard") {
			return fmt.Errorf("%s%s", baseMsg, getMacOSHelp())
		}
	}
	
	return fmt.Errorf("%s", baseMsg)
}

func getMacOSHelp() string {
	if runtime.GOOS != "darwin" {
		return ""
	}
	
	return `

macOS Troubleshooting:
1. Ensure your YubiKey is properly inserted
2. Try pairing your YubiKey with macOS:
   sc_auth pairing_ui -s enable
   (Then insert your YubiKey and follow prompts)
3. Verify smart card detection:
   sc_auth identities
   system_profiler SPSmartCardsDataType
4. Restart the PC/SC daemon:
   sudo launchctl stop com.apple.securityd
   sudo launchctl start com.apple.securityd`
}

func getMacOSDebugHelp() string {
	if runtime.GOOS != "darwin" {
		return ""
	}
	
	return `

The YubiKey was detected but doesn't support OATH or has no OATH credentials.

Possible issues:
1. YubiKey doesn't have OATH applet enabled
2. No OATH credentials configured on the YubiKey
3. Another application is using the YubiKey

Try:
1. Reconnect the YubiKey
2. Kill conflicting processes: killall yubikey-agent
3. Restart PC/SC daemon:
   sudo launchctl stop com.apple.securityd
   sudo launchctl start com.apple.securityd`
}
