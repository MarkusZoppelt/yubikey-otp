package cmd

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/pkg/errors"
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
	Run: func(cmd *cobra.Command, args []string) {
		devices, err := devices.GetDevices()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		device := devices[0]
		if viper.GetString("device") != "" {
			device = viper.GetString("device")
		} else if len(devices) > 1 {
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

			// if the user has yubikey-agent installed, the error might be related to that
			// so we check if yubikey-agent is installed and if so, we print a hint
			_, err = exec.LookPath("yubikey-agent")
			if err == nil {
				fmt.Println(err)
				fmt.Println()
				fmt.Println("If you are using yubikey-agent, you may try to kill it with")
				fmt.Println("`killall -HUP yubikey-agent` and try again.")
			}
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

		var cmdStr string
		if runtime.GOOS == "darwin" {
			// if we are on macOS, we use pbcopy to copy the OTP code to the clipboard
			cmdStr = fmt.Sprintf("ykman --device %s oath accounts code %s | awk 'NF>1{print $NF}' | pbcopy", deviceID, accounts[idx])
		} else if runtime.GOOS == "linux" {
			// if running on linux, use xclip instead of pbcopy
			cmdStr = fmt.Sprintf("ykman --device %s oath accounts code %s | awk 'NF>1{print $NF}' | xclip -selection clipboard", deviceID, accounts[idx])
		} else {
			fmt.Println("Unsupported OS")
			os.Exit(1)
		}

		if viper.GetBool("verbose") {
			fmt.Println(cmdStr)
		}

		fmt.Printf("Using your YubiKey to generate the OTP code for %s...\n", accounts[idx])
		err = exec.Command("sh", "-c", cmdStr).Run()
		if err != nil {
			fmt.Println("Error copying to clipboard: ", errors.WithStack(err))
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

	// TODO: remove awk dependency
	_, err = exec.LookPath("awk")
	if err != nil {
		fmt.Println("awk not found. Please install awk.")
		os.Exit(1)
	}

	flag.String("device", "", "YubiKey device ID")
	flag.Bool("verbose", false, "Enable verbose logging")

	flag.Parse()

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)
}
