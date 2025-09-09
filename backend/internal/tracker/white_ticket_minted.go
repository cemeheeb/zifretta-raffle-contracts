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

func (t *Tracker) collectWhiteTicketMintedActions(raffleStartedLt int64) error {
	logger.Debug("collect white ticket minted actions...")
	var actions = make([]*storage.UserAction, 0)

	logger.Debug("get latest white ticket minted at")
	lastWhiteTicketMintedLt, err := t.storage.GetUserActionTouch(storage.WhiteTicketMintedActionType)
	if err != nil {
		panic("failed to get max white ticket minted at")
	}

	var transactionLt int64 = 0
	var maxTransactionLt int64 = 0
	var beforeLt int64 = 0

	for {
		accountTracesResult, err := infinityRateLimitRetry(
			func() (*tonapi.TraceIDs, error) {
				return t.client.GetAccountTraces(t.ctx, tonapi.GetAccountTracesParams{
					AccountID: t.whiteTicketCollectionAddress,
					Limit:     tonapi.OptInt{Value: GlobalLimitWindowSize, Set: true},
					BeforeLt: tonapi.OptInt64{
						Value: beforeLt,
						Set:   beforeLt > 0,
					},
				})
			})

		if err != nil {
			return nil
		}

		for _, traceID := range accountTracesResult.GetTraces() {
			trace, err := infinityRateLimitRetry(
				func() (*tonapi.Trace, error) {
					return t.client.GetTrace(t.ctx, tonapi.GetTraceParams{TraceID: traceID.GetID()})
				})

			if err != nil {
				logger.Fatal("white ticket minted: cannot get white ticket collection account trace", zap.Error(err))
				return err
			}

			transactionLt = trace.Transaction.GetLt()
			maxTransactionLt = max(maxTransactionLt, transactionLt)

			if transactionLt <= lastWhiteTicketMintedLt {
				logger.Debug("white ticket minted: last processed transaction at reached")
				break
			}

			beforeLt = walkTracesWhiteTicketMinted(trace, func(inner *tonapi.Trace, hasMintOpCode bool) {
				logger.Debug("white ticket minted: found trace", zap.Int64("transaction at", transactionLt))

				transactionLt = inner.Transaction.GetLt()
				maxTransactionLt = max(maxTransactionLt, transactionLt)

				if transactionLt <= lastWhiteTicketMintedLt {
					return
				}

				transactionHash, address, ok := processCollectWhiteTicketMintedTrace(inner, hasMintOpCode)
				if ok && transactionLt > lastWhiteTicketMintedLt {
					logger.Debug("white ticket minted: append action", zap.String("address", address))
					actions = append(actions, &storage.UserAction{
						ActionType:      storage.WhiteTicketMintedActionType,
						Address:         address,
						TransactionLt:   transactionLt,
						TransactionHash: transactionHash,
					})
				} else {
					logger.Debug("white ticket minted: trace cannot be processed, skip")
				}
			}, false, lastWhiteTicketMintedLt, raffleStartedLt)

			if beforeLt < raffleStartedLt {
				logger.Debug("white ticket minted: raffle start time reached, finalize traces results...")
				break
			}

			logger.Debug("white ticket minted: process account trace... iteration done")
		}

		if len(accountTracesResult.GetTraces()) < GlobalLimitWindowSize || beforeLt <= raffleStartedLt || transactionLt <= lastWhiteTicketMintedLt {
			logger.Debug("white ticket minted: exit condition reached, finalize traces results...")
			break
		}

		logger.Debug("white ticket minted: collect nft collection account traces... iteration done")
	}

	if maxTransactionLt > lastWhiteTicketMintedLt {
		actionTouch := &storage.UserActionTouch{
			ActionType:    storage.WhiteTicketMintedActionType,
			Address:       t.whiteTicketCollectionAddress,
			TransactionLt: maxTransactionLt,
		}
		err := t.storage.UpdateUserActionTouch(actionTouch)
		if err != nil {
			logger.Fatal("white ticket minted: failed to update last action transaction state")
			return err
		}
	}

	if len(actions) > 0 {
		err := t.storage.UpdateUserActions(actions)
		if err != nil {
			panic("failed to update pending white ticket minted actions: " + err.Error())
		}
	}

	return nil
}

func walkTracesWhiteTicketMinted(trace *tonapi.Trace, callback func(*tonapi.Trace, bool), hasNFTMintOpCode bool, lastWhiteTicketMintedAt, raffleStartedAt int64) int64 {
	if trace == nil {
		logger.Debug("no trace found, stop walk")
		return math.MaxInt64
	}

	logger.Debug("walk through white ticket minted actions...", zap.String("hash", trace.Transaction.GetHash()))

	inMessage, ok := trace.Transaction.GetInMsg().Get()
	if ok {
		opCode, ok := inMessage.GetOpCode().Get()
		if ok && (opCode == "0x00000001" || opCode == "0x00000002") {
			logger.Debug("found white ticket minted opcode, passing information about it through", zap.String("hash", trace.Transaction.GetHash()))
			hasNFTMintOpCode = true
		}
	}

	callback(trace, hasNFTMintOpCode)

	beforeLt := trace.Transaction.Lt
	for i := range trace.Children {
		if beforeLt < raffleStartedAt || beforeLt < lastWhiteTicketMintedAt {
			break
		}
		beforeLt = min(beforeLt, walkTracesWhiteTicketMinted(&trace.Children[i], callback, hasNFTMintOpCode, lastWhiteTicketMintedAt, raffleStartedAt))
	}

	logger.Debug("walk through white ticket minted actions... done", zap.String("hash", trace.Transaction.GetHash()))
	return beforeLt
}

func processCollectWhiteTicketMintedTrace(trace *tonapi.Trace, hasMintOpCode bool) (string, string, bool) {
	logger.Debug("white ticket minted: process collect white ticket minted trace...", zap.String("hash", trace.Transaction.GetHash()))

	message, ok := trace.Transaction.GetInMsg().Get()
	if ok {
		isDeployed := trace.Transaction.OrigStatus == tonapi.AccountStatusNonexist &&
			trace.Transaction.EndStatus == tonapi.AccountStatusActive

		if hasMintOpCode && isDeployed && trace.Transaction.Success {
			body, err := boc.DeserializeBocHex(message.GetRawBody().Value)
			if err != nil {
				logger.Debug("white ticket minted: failed to deserialize boc hex... skip", zap.String("hash", trace.Transaction.GetHash()))
				return "", "", false
			}

			bodyCell := body[0]

			var userAccountAddress tlb.MsgAddress
			err = tlb.Unmarshal(bodyCell, &userAccountAddress)
			if err != nil {
				logger.Debug("white ticket minted: failed to read user address due to address tlb scheme... skip", zap.String("hash", trace.Transaction.GetHash()))
				return "", "", false
			}

			userAccountID, err := tongo.AccountIDFromTlb(userAccountAddress)
			if userAccountID == nil || err != nil {
				logger.Debug("white ticket minted: invalid user address... skip", zap.String("hash", trace.Transaction.GetHash()))
				return "", "", false
			}

			logger.Debug("process collect white ticket minted trace... done", zap.String("hash", trace.Transaction.GetHash()))
			return message.GetHash(), userAccountID.ToHuman(true, false), true
		}
	}

	logger.Debug("process collect white ticket minted trace... skip", zap.String("hash", trace.Transaction.GetHash()))
	return "", "", false
}
