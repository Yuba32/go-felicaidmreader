package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/ebfe/scard"
)

var (
	apdu_get_idm = []byte{0xff, 0xca, 0x00, 0x00, 0x00}
)

func main() {

	ctx, err := scard.EstablishContext()
	if err != nil {
		//fmt.Println(err)
		os.Exit(1)
	}
	defer ctx.Release()

	readers, err := ctx.ListReaders()
	if err != nil {
		//fmt.Println(err)
		os.Exit(1)
	}

	if len(readers) > 0 {

		index, err := waitUntilCardPresent(ctx, readers)
		if err != nil {
			//fmt.Println(err)
			os.Exit(1)
		}

		st, card := connectCard(ctx, readers[index])
		if st != 0 {
			//fmt.Println("カードの取得に失敗しました")
			os.Exit(1)
		}
		defer card.Disconnect(scard.ResetCard)

		st, idm := readidm(card)
		if st != 0 {
			//fmt.Println("IDmの取得に失敗しました")
			os.Exit(1)
		}

		fmt.Print(idm)
		os.Exit(0)

	}

}

func readidm(card *scard.Card) (int, string) {

	resp, err := card.Transmit(apdu_get_idm)
	if err != nil {
		fmt.Println(err)
		return -1, ""
	}
	//IDmの取得
	rawidm := resp[0:8]

	idm := parseIdm(rawidm)
	idm = strings.ToUpper(idm)
	return 0, idm
}

func waitUntilCardPresent(ctx *scard.Context, readers []string) (int, error) {
	rs := make([]scard.ReaderState, len(readers))
	for i := range rs {
		rs[i].Reader = readers[i]
		rs[i].CurrentState = scard.StateUnaware
	}

	for {
		for i := range rs {
			if rs[i].EventState&scard.StatePresent != 0 {
				return i, nil
			}
			rs[i].CurrentState = rs[i].EventState

		}
		err := ctx.GetStatusChange(rs, -1)
		if err != nil {
			return -1, err
		}
	}
}

func connectCard(ctx *scard.Context, reader string) (int, *scard.Card) {
	card, err := ctx.Connect(reader, scard.ShareExclusive, scard.ProtocolAny)
	if err != nil {
		return -1, nil
	}
	return 0, card
}

func parseIdm(raw []byte) string {
	return fmt.Sprintf("%02x%02x%02x%02x%02x%02x%02x%02x", raw[0], raw[1], raw[2], raw[3], raw[4], raw[5], raw[6], raw[7])
}
