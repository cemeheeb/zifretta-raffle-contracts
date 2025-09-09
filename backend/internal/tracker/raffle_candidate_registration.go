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

func (t *Tracker) collectCandidateRegistrationActions(raffleAddress string, raffleStartedLt int64) error {

	var actions = make([]*storage.UserAction, 0)

	logger.Debug("raffle candidate registration: get last candidate registered at")
	lastCandidateRegistrationLt, err := t.storage.GetUserActionTouch(storage.CandidateRegistrationActionType)
	if err != nil {
		panic("failed to get max candidate registration at")
	}

	var transactionLt int64 = 0
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

			transactionLt = trace.Transaction.GetLt()
			var transactionBt = uint64(transactionLt)
			logger.Debug("transactionBt", zap.Uint64("transactionBt", transactionBt))
			maxTransactionLt = max(maxTransactionLt, transactionLt)

			if transactionLt <= lastCandidateRegistrationLt {
				logger.Debug("raffle candidate registration: last transaction at reached")
				break
			}

			beforeLt = walkTracesCandidateRegistration(trace, func(inner *tonapi.Trace) {
				var transactionHash string
				var address string
				var ok bool

				transactionLt, transactionHash, address, ok = processRaffleCandidateRegistrationTrace(inner)
				maxTransactionLt = max(maxTransactionLt, transactionLt)

				if ok && transactionLt > lastCandidateRegistrationLt {
					actions = append(actions, &storage.UserAction{
						ActionType:      storage.CandidateRegistrationActionType,
						Address:         address,
						TransactionLt:   transactionLt,
						TransactionHash: transactionHash,
					})
				} else {
					logger.Debug("raffle candidate registration: trace cannot be processed, skip")
				}
			}, lastCandidateRegistrationLt, raffleStartedLt)

			if beforeLt < raffleStartedLt {
				logger.Debug("raffle candidate registration: raffle start time reached, finalize traces results...")
				break
			}

			logger.Debug("raffle candidate registration: collect trace details... iteration done")
		}

		if len(accountTracesResult.GetTraces()) < GlobalLimitWindowSize || beforeLt < raffleStartedLt || transactionLt < lastCandidateRegistrationLt {
			logger.Debug("raffle candidate registration: exit condition reached, finalize traces results...")
			break
		}

		logger.Debug("raffle candidate registration: collect raffle candidate account traces... iteration done")
	}

	if maxTransactionLt > lastCandidateRegistrationLt {

		actionTouch := &storage.UserActionTouch{
			ActionType:    storage.CandidateRegistrationActionType,
			Address:       raffleAddress,
			TransactionLt: maxTransactionLt,
		}

		err := t.storage.UpdateUserActionTouch(actionTouch)
		if err != nil {
			logger.Fatal("raffle candidate registration: failed to update last action transaction state")
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

func walkTracesCandidateRegistration(trace *tonapi.Trace, callback func(*tonapi.Trace), lastCandidateRegisteredAt int64, raffleStartedAt int64) int64 {

	if trace == nil {
		return math.MaxInt64
	}

	callback(trace)

	beforeLt := trace.Transaction.GetLt()
	for i := range trace.Children {
		if beforeLt < raffleStartedAt || beforeLt < lastCandidateRegisteredAt {
			break
		}
		beforeLt = min(beforeLt, walkTracesCandidateRegistration(&trace.Children[i], callback, lastCandidateRegisteredAt, raffleStartedAt))
	}

	return beforeLt
}

func processRaffleCandidateRegistrationTrace(trace *tonapi.Trace) (int64, string, string, bool) {

	raffleCandidateInitializeOpCode := tonapi.OptString{Value: "0x13370020", Set: true}
	message, ok := trace.Transaction.GetInMsg().Get()
	if ok {
		isTargetOpCode := message.GetOpCode() == raffleCandidateInitializeOpCode
		isDeployed := trace.Transaction.OrigStatus == tonapi.AccountStatusNonexist &&
			trace.Transaction.EndStatus == tonapi.AccountStatusActive

		if isTargetOpCode && isDeployed && trace.Transaction.Success {
			body, err := boc.DeserializeBocHex(message.GetRawBody().Value)
			if err != nil {
				return 0, "", "", false
			}

			bodyCell := body[0]
			err = bodyCell.Skip(32) //op-code
			if err != nil {
				return 0, "", "", false
			}

			err = bodyCell.Skip(64) // telegramID
			if err != nil {
				return 0, "", "", false
			}

			var userAccountAddress tlb.MsgAddress
			err = tlb.Unmarshal(bodyCell, &userAccountAddress)
			if err != nil {
				return 0, "", "", false
			}

			userAccountID, err := tongo.AccountIDFromTlb(userAccountAddress)
			if userAccountID == nil || err != nil {
				return 0, "", "", false
			}

			return message.GetCreatedLt(), message.GetHash(), userAccountID.ToHuman(true, false), true
		}
	}

	return 0, "", "", false
}
