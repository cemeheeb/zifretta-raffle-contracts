package tracker

import (
	"backend/internal/logger"
	"errors"
	"strconv"
	"strings"

	"github.com/tonkeeper/tonapi-go"
	"github.com/tonkeeper/tongo/boc"
	"github.com/tonkeeper/tongo/ton"
	"go.uber.org/zap"
)

type RaffleConditions struct {
	BlackTicketPurchased uint8
	WhiteTicketMinted    uint8
}

type RaffleAccountData struct {
	MinCandidateQuantity        uint32
	ConditionsDuration          uint32
	Conditions                  RaffleConditions
	MinCandidateReachedLt       uint64
	MinCandidateReachedUnixTime int64
	CandidatesQuantity          uint64
	ParticipantsQuantity        uint64
	WinnersQuantity             uint8
	Winners                     []string
}

func (t *Tracker) GetRaffleAccountData() (*RaffleAccountData, error) {

	_, err := ton.ParseAccountID(t.raffleAddress)
	if err != nil {
		logger.Fatal("get raffle account data: failed to parse raffle address", zap.String("raffle address", t.raffleAddress), zap.Error(err))
		return nil, err
	}

	raffleData, err := infinityRateLimitRetry(
		func() (*tonapi.MethodExecutionResult, error) {
			return t.client.ExecGetMethodForBlockchainAccount(t.ctx, tonapi.ExecGetMethodForBlockchainAccountParams{
				AccountID:  t.raffleAddress,
				MethodName: "raffleData",
				Args:       make([]string, 0),
			})
		})

	if err != nil {
		logger.Fatal("get raffle account data: failed to get raffleData, invalid raffle contract")
		return nil, err
	}

	minCandidateQuantityString, ok := raffleData.GetStack()[0].GetNum().Get()
	if !ok {
		return nil, errors.New("get raffle account data: failed to get raffle account data")
	}
	minCandidateQuantity, err := strconv.ParseInt(strings.TrimLeft(minCandidateQuantityString, "0x"), 16, 32)

	conditionsDurationString, ok := raffleData.GetStack()[1].GetNum().Get()
	if !ok {
		return nil, errors.New("get raffle account data: failed to get raffle account data")
	}
	conditionsDuration, err := strconv.ParseInt(strings.TrimLeft(conditionsDurationString, "0x"), 16, 32)

	conditionsString, ok := raffleData.GetStack()[2].GetCell().Get()
	if !ok {
		return nil, errors.New("get raffle account data: failed to get raffle account data")
	}

	conditions, err := boc.DeserializeBocHex(conditionsString)
	if err != nil {
		return nil, errors.New("get raffle account data: failed to get raffle account data")
	}

	bodyCell := conditions[0]
	blackTicketPurchased, err := bodyCell.ReadInt(8)
	if err != nil {
		return nil, errors.New("get raffle account data: failed to get raffle account data")
	}

	whiteTicketMinted, err := bodyCell.ReadInt(8)
	if err != nil {
		return nil, errors.New("get raffle account data: failed to get raffle account data")
	}

	logger.Debug("conditions", zap.Int64("blackTicketPurchased", blackTicketPurchased), zap.Int64("whiteTicketMinted", whiteTicketMinted))

	minCandidateReachedLtString, ok := raffleData.GetStack()[3].GetNum().Get()
	if !ok {
		return nil, errors.New("get raffle account data: failed to get raffle account data")
	}
	minCandidateReachedLt, err := strconv.ParseInt(strings.TrimLeft(minCandidateReachedLtString, "0x"), 16, 64)

	minCandidateReachedUnixTimeString, ok := raffleData.GetStack()[4].GetNum().Get()
	if !ok {
		return nil, errors.New("get raffle account data: failed to get raffle account data")
	}
	minCandidateReachedUnixTime, err := strconv.ParseInt(strings.TrimLeft(minCandidateReachedUnixTimeString, "0x"), 16, 64)

	candidatesQuantityString, ok := raffleData.GetStack()[5].GetNum().Get()
	if !ok {
		return nil, errors.New("get raffle account data: failed to get raffle account data")
	}
	candidatesQuantity, err := strconv.ParseInt(strings.TrimLeft(candidatesQuantityString, "0x"), 16, 64)

	participantsQuantityString, ok := raffleData.GetStack()[6].GetNum().Get()
	if !ok {
		return nil, errors.New("get raffle account data: failed to get raffle account data")
	}
	participantsQuantity, err := strconv.ParseInt(strings.TrimLeft(participantsQuantityString, "0x"), 16, 64)

	winnersQuantityString, ok := raffleData.GetStack()[6].GetNum().Get()
	if !ok {
		return nil, errors.New("get raffle account data: failed to get raffle account data")
	}
	winnersQuantity, err := strconv.ParseInt(strings.TrimLeft(winnersQuantityString, "0x"), 16, 8)

	return &RaffleAccountData{
		MinCandidateQuantity: uint32(minCandidateQuantity),
		ConditionsDuration:   uint32(conditionsDuration),
		Conditions: RaffleConditions{
			BlackTicketPurchased: uint8(blackTicketPurchased),
			WhiteTicketMinted:    uint8(whiteTicketMinted),
		},
		MinCandidateReachedLt:       uint64(minCandidateReachedLt),
		MinCandidateReachedUnixTime: int64(minCandidateReachedUnixTime),
		CandidatesQuantity:          uint64(candidatesQuantity),
		ParticipantsQuantity:        uint64(participantsQuantity),
		WinnersQuantity:             uint8(winnersQuantity),
		Winners:                     []string{},
	}, nil
}
