package tracker

import (
	"backend/internal/logger"
	"errors"

	"github.com/tonkeeper/tonapi-go"
	"github.com/tonkeeper/tongo/ton"
	"go.uber.org/zap"
)

func (t *Tracker) GetRaffleAccountDeployedLt() (int64, error) {
	var lastTraceID *tonapi.TraceID = nil

	raffleAccountID, err := ton.ParseAccountID(t.raffleAddress)
	if err != nil {
		logger.Fatal("verify raffle account: failed to parse raffle address", zap.String("raffle address", t.raffleAddress), zap.Error(err))
		return 0, err
	}

	var beforeLt = tonapi.OptInt64{Value: 0, Set: false}
	for {
		if lastTraceID != nil {
			lastTrace, err := infinityRateLimitRetry(
				func() (*tonapi.Trace, error) {
					return t.client.GetTrace(t.ctx, tonapi.GetTraceParams{TraceID: lastTraceID.GetID()})
				},
			)
			if err != nil {
				return 0, err
			}
			beforeLt = tonapi.NewOptInt64(lastTrace.Transaction.Lt)
		}

		logger.Debug("verify raffle account: search first traceID")
		accountTracesResult, err := infinityRateLimitRetry(
			func() (*tonapi.TraceIDs, error) {
				return t.client.GetAccountTraces(t.ctx, tonapi.GetAccountTracesParams{
					AccountID: raffleAccountID.ToRaw(),
					Limit:     tonapi.NewOptInt(GlobalLimitWindowSize),
					BeforeLt:  beforeLt,
				})
			})

		if err != nil {
			logger.Fatal("verify raffle account: failed to search first traceID", zap.Error(err))
			return 0, err
		}

		if len(accountTracesResult.Traces) > 0 {
			lastTraceID = &accountTracesResult.Traces[len(accountTracesResult.Traces)-1]
		}

		logger.Debug("verify raffle account: check conditions in proper to continue", zap.Int("trace count", len(accountTracesResult.Traces)))
		if len(accountTracesResult.Traces) < GlobalLimitWindowSize {
			break
		}
	}

	if lastTraceID == nil {
		return 0, errors.New("verify raffle account: no traces found")
	}

	lastTrace, err := infinityRateLimitRetry(
		func() (*tonapi.Trace, error) {
			return t.client.GetTrace(t.ctx, tonapi.GetTraceParams{TraceID: lastTraceID.GetID()})
		},
	)

	if err != nil {
		return 0, err
	}

	return lastTrace.Transaction.GetLt(), nil
}
