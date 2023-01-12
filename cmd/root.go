package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/MarkusZoppelt/yubikey-otp/internal/accounts"
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
		accounts, err := accounts.GetOTPAccountsFromYubiKey()
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

		// build command to copy to clipboard
		cmdStr := fmt.Sprintf("ykman oath accounts code %s | awk 'NF>1{print $NF}' | pbcopy", accounts[idx])

		// fmt.Println("Running command: ", cmdStr)
		fmt.Printf("Please touch your YubiKey to generate the OTP code for %s\n", accounts[idx])
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
	// check if ykman is installed
	_, err := exec.LookPath("ykman")
	if err != nil {
		fmt.Println("ykman not found. Please install ykman.")
		os.Exit(1)
	}
}
