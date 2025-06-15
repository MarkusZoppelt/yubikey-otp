package accounts

import (
	"cunicu.li/go-iso7816/drivers/pcsc"
	"cunicu.li/go-ykoath/v2"
	"github.com/ebfe/scard"
)

func GetOTPAccountsFromYubiKey(device string) (accounts []string, err error) {
	ctx, err := scard.EstablishContext()
	if err != nil {
		return nil, err
	}
	defer ctx.Release()

	card, err := pcsc.NewCard(ctx, device, true)
	if err != nil {
		return nil, err
	}
	defer card.Close()

	ykoathCard, err := ykoath.NewCard(card)
	if err != nil {
		return nil, err
	}
	defer ykoathCard.Close()

	_, err = ykoathCard.Select()
	if err != nil {
		return nil, err
	}

	names, err := ykoathCard.List()
	if err != nil {
		return nil, err
	}

	for _, name := range names {
		accounts = append(accounts, name.Name)
	}

	return accounts, nil
}
