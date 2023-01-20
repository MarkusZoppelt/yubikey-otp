# YubiKey OTP

A command line tool that lets you search for TOTP (oath) secrets on your
YubiKey with fuzzy search and easily copies them to your clipboard after
selecting.

### Installation:

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

[`ykman`](https://github.com/Yubico/yubikey-manager) is a powerful and useful
tool, but running `ykman oath accounts list` and `ykman oath accounts code
<Account:user>` just for getting TOTP secrets feels long and convoluted. And
even then you have to select the TOTP code and copy it manually... like an
animal! `yubikey-otp` has a nicer UX imho. Try it out! ;)

### Known issues

#### Conflict with yubikey-agent

[`yubikey-agent` takes a persistent transaction so the YubiKey will cache the PIN after first use](https://github.com/FiloSottile/yubikey-agent#conflicts-with-gpg-agent-and-yubikey-manager).
In that case, kill `yubikey-agent` and try again.

    killall -HUP yubikey-agent

Don't worry, `yubikey-agent` will restart the next time you want to use it.
