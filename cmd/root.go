package cmd

import (
	"crypto/hmac"
	"crypto/sha1"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"cunicu.li/go-iso7816/drivers/pcsc"
	"cunicu.li/go-ykoath/v2"
	"github.com/MarkusZoppelt/yubikey-otp/internal/clipboard"
	"github.com/ebfe/scard"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/term"

	"github.com/MarkusZoppelt/yubikey-otp/internal/accounts"
	"github.com/MarkusZoppelt/yubikey-otp/internal/devices"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "yubikey-otp",
	Short: "Search, display and copy YubiKey OTP codes",
	Long: `Search, display and copy YubiKey OTP codes.
A YubiKey is required to use this tool. After connecting the YubiKey, run the
yubiky-otp command to display the OTP codes. The codes are displayed in a fuzzy
searchable list. Select the code you want to copy to the clipboard.`,
	Run: func(_ *cobra.Command, _ []string) {
		devices, err := devices.GetDevices()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if len(devices) == 0 {
			fmt.Println("No YubiKey devices found. Please make sure your YubiKey is connected.")
			os.Exit(1)
		}

		device := devices[0]
		if viper.GetString("device") != "" {
			device = viper.GetString("device")
		}

		if len(devices) > 1 {
			prompt := pterm.DefaultInteractiveSelect.WithDefaultText("Select YubiKey device").WithOptions(devices)
			device, err = prompt.Show()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		accounts, err := accounts.GetOTPAccountsFromYubiKey(device)
		if err != nil {
			fmt.Println("Error getting OATH accounts:", err)
			os.Exit(1)
		}

		idx, err := fuzzyfinder.Find(
			accounts,
			func(i int) string {
				return accounts[i]
			},
		)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Printf("Using your YubiKey to generate the OTP code for %s...\n", accounts[idx])
		otpCode, err := generateOTPCode(device, accounts[idx])
		if err != nil {
			fmt.Println("Error generating OTP code:", err)
			os.Exit(1)
		}

		// copy the OTP code to the clipboard
		err = clipboard.WriteAll(otpCode)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("Copied OTP code for %s to clipboard.\n", accounts[idx])
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func generateOTPCode(device, account string) (string, error) {
	ctx, err := scard.EstablishContext()
	if err != nil {
		return "", err
	}
	defer ctx.Release()

	card, err := pcsc.NewCard(ctx, device, true)
	if err != nil {
		return "", err
	}
	defer card.Close()

	ykoathCard, err := ykoath.NewCard(card)
	if err != nil {
		return "", err
	}
	defer ykoathCard.Close()

	selectResp, err := ykoathCard.Select()
	if err != nil {
		return "", err
	}

	// Check if OATH applet requires authentication
	if selectResp.Challenge != nil {
		fmt.Print("Enter OATH passphrase: ")
		passphrase, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println() // Add newline after password input
		if err != nil {
			return "", fmt.Errorf("failed to read passphrase: %v", err)
		}

		// Derive key from passphrase using PBKDF2 with device ID as salt
		// Use first 16 bytes as specified in OATH protocol
		key := pbkdf2.Key(passphrase, selectResp.Name, 1000, 16, sha1.New)
		
		// Calculate response to challenge
		mac := hmac.New(sha1.New, key)
		mac.Write(selectResp.Challenge)
		response := mac.Sum(nil)

		// Validate with the challenge response
		err = ykoathCard.Validate(response)
		if err != nil {
			return "", fmt.Errorf("failed to authenticate with OATH applet (incorrect passphrase?): %v", err)
		}
	}

	code, err := ykoathCard.Calculate(account)
	if err != nil {
		return "", err
	}

	return code, nil
}

func init() {
	// if yubikey-agent is installed. we might need to kill it
	// otherwise the YubiKey connection might fail
	_, err := exec.LookPath("yubikey-agent")
	if err == nil {
		err = exec.Command("killall", "-HUP", "yubikey-agent").Run()
		if err != nil {
			fmt.Println("Error killing yubikey-agent:", err)
		}
	}

	flag.String("device", "", "YubiKey device ID")
	flag.Bool("verbose", false, "Enable verbose logging")

	flag.Parse()

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)
}
