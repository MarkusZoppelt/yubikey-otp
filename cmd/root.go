package cmd

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

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

		// get serial number of YubiKey (last word of device string)
		deviceID := device[strings.LastIndex(device, " ")+1:]

		accounts, err := accounts.GetOTPAccountsFromYubiKey(deviceID)
		if err != nil {
			fmt.Println("Please make sure that the YubiKey is connected and that ykman is installed.")
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

		cmdStr := fmt.Sprintf("ykman --device %s oath accounts code '%s'", deviceID, accounts[idx])

		if viper.GetBool("verbose") {
			fmt.Println(cmdStr)
		}

		fmt.Printf("Using your YubiKey to generate the OTP code for %s...\n", accounts[idx])
		out, err := exec.Command("sh", "-c", cmdStr).Output()
		if err != nil {
			fmt.Println("Error executing command:", string(out))
			fmt.Println(err)
			os.Exit(1)
		}

		// get the OTP code from the output
		otpCode := out[strings.LastIndex(string(out), " ")+1:]

		// copy the OTP code to the clipboard
		err = clipboard.WriteAll(string(otpCode))
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

func init() {
	_, err := exec.LookPath("ykman")
	if err != nil {
		fmt.Println("ykman not found. Please install ykman.")
		os.Exit(1)
	}

	// if yubikey-agent is installed. we might need to kill it
	// otherwise the ykman command will fail
	_, err = exec.LookPath("yubikey-agent")
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
