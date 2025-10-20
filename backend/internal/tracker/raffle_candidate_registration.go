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

func (t *Tracker) collectCandidateRegistrationActions(raffleAddress string, raffleDeployedLt int64) error {

	logger.Debug("raffle candidate registration: get last candidate registered at")

	var actions = make([]*storage.UserAction, 0)
	lastCandidateRegistrationLt, err := t.storage.GetUserActionTouch(storage.CandidateRegistrationActionType)
	if err != nil {
		panic("failed to get max candidate registration at")
	}

	var transactionLt int64 = 0
	var transactionUnixTime int64 = 0
	var maxTransactionLt int64 = 0
	var beforeLt int64 = 0

	for {
		logger.Debug("raffle candidate registration: collect traces... iteration", zap.Int64("current beforeLt", beforeLt))
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
			logger.Fatal("raffle candidate registration: collect traces... failed", zap.Error(err))
			return err
		}

		for _, traceID := range accountTracesResult.GetTraces() {
			logger.Debug("raffle candidate registration: collect trace details... iteration", zap.String("trace id", traceID.GetID()))
			trace, err := infinityRateLimitRetry(
				func() (*tonapi.Trace, error) {
					return t.client.GetTrace(t.ctx, tonapi.GetTraceParams{TraceID: traceID.GetID()})
				},
			)

			if err != nil {
				logger.Debug("raffle candidate registration: collect trace details... failed", zap.Error(err))
				break
			}

			transactionLt = trace.Transaction.Lt
			transactionUnixTime = trace.Transaction.Utime
			maxTransactionLt = max(maxTransactionLt, transactionLt)

			if transactionLt <= lastCandidateRegistrationLt {
				logger.Debug("raffle candidate registration: last transaction at reached")
				break
			}

			beforeLt = walkTracesCandidateRegistration(trace, func(inner *tonapi.Trace) {
				transactionLt = inner.Transaction.Lt
				transactionUnixTime = inner.Transaction.Utime
				transactionHash, processedUserAddress, processedCandidateAddress, ok := processRaffleCandidateRegistrationTrace(inner)

				if ok && transactionLt > lastCandidateRegistrationLt {
					logger.Info("raffle candidate registration: append action",
						zap.String("user address", processedUserAddress),
						zap.String("ticket address", processedUserAddress),
					)

					actions = append(actions, &storage.UserAction{
						ActionType:          storage.CandidateRegistrationActionType,
						UserAddress:         processedUserAddress,
						Address:             processedCandidateAddress,
						TransactionLt:       transactionLt,
						TransactionHash:     transactionHash,
						TransactionUnixTime: transactionUnixTime,
					})
				} else {
					logger.Debug("raffle candidate registration: trace cannot be processed, skip")
				}
			}, lastCandidateRegistrationLt, raffleDeployedLt)

			if beforeLt < raffleDeployedLt {
				logger.Debug("raffle candidate registration: raffle start time reached, finalize traces results...")
				break
			}

			logger.Debug("raffle candidate registration: collect trace details... iteration done")
		}

		if len(accountTracesResult.GetTraces()) < GlobalLimitWindowSize || beforeLt < raffleDeployedLt || transactionLt <= lastCandidateRegistrationLt {
			logger.Debug("raffle candidate registration: exit condition reached, finalize traces results...")
			break
		}

		logger.Debug("raffle candidate registration: collect raffle candidate account traces... iteration done")
	}

	if maxTransactionLt > lastCandidateRegistrationLt {

		actionTouch := &storage.UserActionTouch{
			ActionType:    storage.CandidateRegistrationActionType,
			UserAddress:   "-",
			TransactionLt: maxTransactionLt,
		}

		err := t.storage.UpdateUserActionTouch(actionTouch)
		if err != nil {
			logger.Fatal("raffle candidate registration: failed to update last action transaction state", zap.Error(err))
			return err
		}
	}

	if len(actions) > 0 {
		err := t.storage.UpdateUserActions(actions)
		if err != nil {
			panic("failed to update pending candidate registration actions")
		}
	}

	return nil
}

func walkTracesCandidateRegistration(trace *tonapi.Trace, callback func(*tonapi.Trace), lastCandidateRegisteredAt int64, raffleDeployedLt int64) int64 {

	if trace == nil {
		return math.MaxInt64
	}

	callback(trace)

	transactionLt := trace.Transaction.Lt
	for i := range trace.Children {
		if transactionLt < raffleDeployedLt || transactionLt < lastCandidateRegisteredAt {
			break
		}

		transactionLt = min(transactionLt, walkTracesCandidateRegistration(&trace.Children[i], callback, lastCandidateRegisteredAt, raffleDeployedLt))
	}

	return transactionLt
}

func processRaffleCandidateRegistrationTrace(trace *tonapi.Trace) (string, string, string, bool) {

	raffleCandidateInitializeOpCode := "0x13370020"
	message, ok := trace.Transaction.GetInMsg().Get()
	if ok {
		isTargetOpCode := message.OpCode.IsSet() && message.OpCode.Value == raffleCandidateInitializeOpCode
		isDeployed := trace.Transaction.OrigStatus == tonapi.AccountStatusNonexist &&
			trace.Transaction.EndStatus == tonapi.AccountStatusActive

		if isTargetOpCode && isDeployed && trace.Transaction.Success {
			body, err := boc.DeserializeBocHex(message.GetRawBody().Value)
			if err != nil {
				logger.Debug("raffle candidate registration: failed to deserialize trace body")
				return "", "", "", false
			}

			bodyCell := body[0]
			err = bodyCell.Skip(32) //op-code
			if err != nil {
				logger.Debug("raffle candidate registration: trace body cell underflow")
				return "", "", "", false
			}

			var userAccountAddress tlb.MsgAddress
			err = tlb.Unmarshal(bodyCell, &userAccountAddress)
			if err != nil {
				logger.Debug("raffle candidate registration: user account address deserialisation failed")
				return "", "", "", false
			}

			userAccountID, err := tongo.AccountIDFromTlb(userAccountAddress)
			if userAccountID == nil || err != nil {
				logger.Debug("raffle candidate registration: user account address is invalid", zap.Error(err))
				return "", "", "", false
			}

			inMessage, ok := trace.Transaction.InMsg.Get()
			if !ok {
				logger.Debug("raffle candidate registration: invalid trace data")
				return "", "", "", false
			}

			inMessageDestination, ok := inMessage.Destination.Get()
			if !ok {
				logger.Debug("raffle candidate registration: invalid trace in message")
				return "", "", "", false
			}

			candidateAddress, err := tongo.ParseAddress(inMessageDestination.Address)
			if err != nil {
				logger.Debug("raffle candidate registration: invalid candidate address")
				return "", "", "", false
			}

			return message.GetHash(), userAccountID.ToHuman(true, false), candidateAddress.ID.ToHuman(true, false), true
		}
	}

	return "", "", "", false
}
