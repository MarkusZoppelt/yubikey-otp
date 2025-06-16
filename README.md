# YubiKey OTP

A command line tool that lets you search for TOTP (oath) secrets on your
YubiKey with fuzzy search and easily copies them to your clipboard after
selecting.

### Installation:

#### Prerequisites

**Linux:**
```bash
# Ubuntu/Debian
sudo apt-get install libpcsclite-dev wl-clipboard # or xclip for x11

# Arch Linux
sudo pacman -S pcsclite wl-clipboard # or xclip for x11
```

**macOS:**

PC/SC framework is built-in, clipboard support is built-in, so no additional
dependencies required

#### Install the tool:

    go install github.com/MarkusZoppelt/yubikey-otp@latest

### Usage

    Search, display and copy YubiKey OTP codes.
    A YubiKey is required to use this tool. After connecting the YubiKey, run the
    yubiky-otp command to display the OTP codes. The codes are displayed in a fuzzy
    searchable list. Select the code you want to copy to the clipboard.

    Usage:
      yubikey-otp [flags]

    Flags:
      -h, --help        help for yubikey-otp
      --device string   YubiKey device ID
      --verbose         Enable verbose logging


### Motivation:

This tool provides a streamlined way to access TOTP secrets from your YubiKey
with minimal system dependencies. It uses Go to communicate directly with your 
YubiKey via PC/SC, offering fuzzy search and automatic clipboard copying for a 
better user experience.

### Known issues

#### OATH Password Authentication

Currently not supported due to go-ykoath library limitations. Remove password with `ykman oath access change --remove` or use ykman directly.

#### Conflict with yubikey-agent

[`yubikey-agent` takes a persistent transaction so the YubiKey will cache the PIN after first use](https://github.com/FiloSottile/yubikey-agent#conflicts-with-gpg-agent-and-yubikey-manager).
To mitigate that issue, `yubikey-otp` will run `killall -HUP yubikey-agent`
during init.

Don't worry, `yubikey-agent` will restart the next time you want to use it.
