package tracker

import (
	"backend/internal/logger"
	"backend/internal/storage"
	"math"

	"github.com/tonkeeper/tonapi-go"
	"github.com/tonkeeper/tongo"
	"github.com/tonkeeper/tongo/boc"
	"github.com/tonkeeper/tongo/tlb"
	"go.uber.org/zap"
)

func (t *Tracker) collectParticipantRegistrationActions(raffleAddress string, raffleDeployedLt int64) error {

	var actions = make([]*storage.UserAction, 0)

	logger.Debug("raffle participant registration: get last participant registered at")
	lastParticipantRegistrationLt, err := t.storage.GetUserActionTouch(storage.ParticipantRegistrationActionType)
	if err != nil {
		panic(err)
	}

	var transactionLt int64 = 0
	var transactionUnixTime int64 = 0
	var maxTransactionLt int64 = 0
	var beforeLt int64 = 0

	for {
		logger.Debug("raffle participant registration: collect traces... iteration", zap.Int64("current beforeLt", beforeLt))
		accountTracesResult, err := infinityRateLimitRetry(
			func() (*tonapi.TraceIDs, error) {
				return t.client.GetAccountTraces(t.ctx, tonapi.GetAccountTracesParams{
					AccountID: raffleAddress,
					Limit:     tonapi.NewOptInt(GlobalLimitWindowSize),
					BeforeLt: tonapi.OptInt64{
						Value: beforeLt,
						Set:   beforeLt > 0,
					},
				})
			},
		)

		if err != nil {
			logger.Fatal("raffle participant registration: collect traces... failed", zap.Error(err))
			return err
		}

		for _, traceID := range accountTracesResult.GetTraces() {
			logger.Debug("raffle participant registration: collect trace details... iteration", zap.String("trace id", traceID.GetID()))
			trace, err := infinityRateLimitRetry(
				func() (*tonapi.Trace, error) {
					return t.client.GetTrace(t.ctx, tonapi.GetTraceParams{TraceID: traceID.GetID()})
				},
			)

			if err != nil {
				logger.Debug("raffle participant registration: collect trace details... failed", zap.Error(err))
				break
			}

			transactionLt = trace.Transaction.Lt
			transactionUnixTime = trace.Transaction.Utime
			maxTransactionLt = max(maxTransactionLt, transactionLt)

			if transactionLt <= lastParticipantRegistrationLt {
				logger.Debug("raffle participant registration: last transaction at reached")
				break
			}

			beforeLt = walkTracesParticipantRegistration(trace, func(inner *tonapi.Trace) {
				transactionLt = inner.Transaction.Lt
				transactionUnixTime = inner.Transaction.Utime
				transactionHash, processedUserAddress, processedParticipantAddress, ok := processRaffleParticipantRegistrationTrace(inner)

				if ok && transactionLt > lastParticipantRegistrationLt {
					actions = append(actions, &storage.UserAction{
						ActionType:          storage.ParticipantRegistrationActionType,
						UserAddress:         processedUserAddress,
						Address:             processedParticipantAddress,
						TransactionLt:       transactionLt,
						TransactionHash:     transactionHash,
						TransactionUnixTime: transactionUnixTime,
					})
				} else {
					logger.Debug("raffle participant registration: trace cannot be processed, skip")
				}
			}, lastParticipantRegistrationLt, raffleDeployedLt)

			if beforeLt < raffleDeployedLt {
				logger.Debug("raffle participant registration: raffle start time reached, finalize traces results...")
				break
			}

			logger.Debug("raffle participant registration: collect trace details... iteration done")
		}

		if len(accountTracesResult.GetTraces()) < GlobalLimitWindowSize || beforeLt < raffleDeployedLt || transactionLt <= lastParticipantRegistrationLt {
			logger.Debug("raffle participant registration: exit condition reached, finalize traces results...")
			break
		}

		logger.Debug("raffle participant registration: collect raffle participant account traces... iteration done")
	}

	if maxTransactionLt > lastParticipantRegistrationLt {

		actionTouch := &storage.UserActionTouch{
			ActionType:    storage.ParticipantRegistrationActionType,
			UserAddress:   "-",
			TransactionLt: maxTransactionLt,
		}

		err := t.storage.UpdateUserActionTouch(actionTouch)
		if err != nil {
			logger.Fatal("raffle participant registration: failed to update last action transaction state")
			return err
		}
	}

	if len(actions) > 0 {
		err := t.storage.UpdateUserActions(actions)
		if err != nil {
			panic(err)
		}
	}

	return nil
}

func walkTracesParticipantRegistration(trace *tonapi.Trace, callback func(*tonapi.Trace), lastParticipantRegisteredAt int64, raffleDeployedAt int64) int64 {

	if trace == nil {
		return math.MaxInt64
	}

	callback(trace)

	transactionLt := trace.Transaction.Lt
	for i := range trace.Children {
		if transactionLt < raffleDeployedAt || transactionLt < lastParticipantRegisteredAt {
			break
		}

		transactionLt = min(transactionLt, walkTracesParticipantRegistration(&trace.Children[i], callback, lastParticipantRegisteredAt, raffleDeployedAt))
	}

	return transactionLt
}

func processRaffleParticipantRegistrationTrace(trace *tonapi.Trace) (string, string, string, bool) {

	raffleParticipantInitializeOpCode := tonapi.OptString{Value: "0x13370030", Set: true}
	message, ok := trace.Transaction.GetInMsg().Get()
	if ok {
		isTargetOpCode := message.GetOpCode() == raffleParticipantInitializeOpCode
		isDeployed := trace.Transaction.OrigStatus == tonapi.AccountStatusNonexist &&
			trace.Transaction.EndStatus == tonapi.AccountStatusActive

		if isTargetOpCode && isDeployed && trace.Transaction.Success {
			body, err := boc.DeserializeBocHex(message.GetRawBody().Value)
			if err != nil {
				logger.Debug("raffle participant registration: failed to deserialize trace body")
			}

			bodyCell := body[0]
			err = bodyCell.Skip(32) //op-code
			if err != nil {
				logger.Debug("raffle participant registration: trace body cell underflow")
				return "", "", "", false
			}

			var userAccountAddress tlb.MsgAddress
			err = tlb.Unmarshal(bodyCell, &userAccountAddress)
			if err != nil {
				logger.Debug("raffle participant registration: user account address deserialisation failed")
				return "", "", "", false
			}

			userAccountID, err := tongo.AccountIDFromTlb(userAccountAddress)
			if userAccountID == nil || err != nil {
				logger.Debug("raffle participant registration: user account address is invalid")
				return "", "", "", false
			}

			inMessage, ok := trace.Transaction.InMsg.Get()
			if !ok {
				logger.Debug("raffle participant registration: invalid trace data")
				return "", "", "", false
			}

			inMessageDestination, ok := inMessage.Destination.Get()
			if !ok {
				logger.Debug("raffle participant registration: invalid trace in message")
				return "", "", "", false
			}

			participantAddress, err := tongo.ParseAddress(inMessageDestination.Address)
			if err != nil {
				logger.Debug("raffle participant registration: invalid participant address")
				return "", "", "", false
			}

			return message.GetHash(), userAccountID.ToHuman(true, false), participantAddress.ID.ToHuman(true, false), true
		}
	}

	return "", "", "", false
}
