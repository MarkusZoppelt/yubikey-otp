package devices

import (
	"fmt"

	"cunicu.li/go-iso7816/drivers/pcsc"
	"cunicu.li/go-ykoath/v2"
	"github.com/ebfe/scard"
)

func GetDevices() (devices []string, err error) {
	ctx, err := scard.EstablishContext()
	if err != nil {
		return nil, err
	}
	defer ctx.Release()

	readers, err := ctx.ListReaders()
	if err != nil {
		return nil, err
	}

	for _, reader := range readers {
		card, err := pcsc.NewCard(ctx, reader, true)
		if err != nil {
			continue
		}

		ykoathCard, err := ykoath.NewCard(card)
		if err != nil {
			card.Close()
			continue
		}

		_, err = ykoathCard.Select()
		if err != nil {
			ykoathCard.Close()
			continue
		}

		devices = append(devices, fmt.Sprintf("%s", reader))
		ykoathCard.Close()
	}

	return devices, nil
}
