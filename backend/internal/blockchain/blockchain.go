package blockchain

import (
	"github.com/tonkeeper/tongo/boc"
	"github.com/tonkeeper/tongo/tlb"
	"github.com/tonkeeper/tongo/ton"
	"github.com/tonkeeper/tongo/wallet"
)

type RaffleSetConditionMessage struct {
	AttachedTon          tlb.Grams
	RaffleAddress        ton.AccountID
	UserAddress          ton.AccountID
	WhiteTicketMinted    uint8
	BlackTicketPurchased uint8
}

type RaffleSetConditionMessageBody struct {
	UserAddress tlb.MsgAddress
	Conditions  tlb.Uint256
}

func (m RaffleSetConditionMessage) ToInternal() (tlb.Message, uint8, error) {

	cell := boc.NewCell()
	if err := cell.WriteUint(0x13370011, 32); err != nil {

		return tlb.Message{}, 0, err
	}
	if err := tlb.Marshal(cell, m.UserAddress.ToMsgAddress()); err != nil {
		return tlb.Message{}, 0, err
	}
	if err := cell.WriteUint(uint64(m.WhiteTicketMinted), 8); err != nil {
		return tlb.Message{}, 0, err
	}
	if err := cell.WriteUint(uint64(m.BlackTicketPurchased), 8); err != nil {
		return tlb.Message{}, 0, err
	}
	if err := cell.WriteUint(0, 256-8-8); err != nil {
		return tlb.Message{}, 0, err
	}

	message := wallet.Message{
		Amount:  m.AttachedTon,
		Address: m.RaffleAddress,
		Bounce:  true,
		Mode:    wallet.DefaultMessageMode,
		Body:    cell,
	}

	return message.ToInternal()
}
