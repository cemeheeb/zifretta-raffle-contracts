package blockchain

import (
	"github.com/tonkeeper/tongo/tlb"
	"github.com/tonkeeper/tongo/ton"
)

type RaffleSetConditionMessage struct {
	Amount               tlb.Grams
	Address              ton.AccountID
	UserAddress          ton.AccountID
	WhiteTicketMinted    uint8
	BlackTicketPurchased uint8
}

type RaffleSetConditionMessageBody struct {
	UserAddress tlb.MsgAddress
	Conditions  tlb.Uint256
}
