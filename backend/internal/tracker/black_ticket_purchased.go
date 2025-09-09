package tracker

import (
	"backend/internal/logger"
	"backend/internal/storage"
	"errors"
	"math"
	"slices"

	"github.com/tonkeeper/tonapi-go"
	"github.com/tonkeeper/tongo"
	"github.com/tonkeeper/tongo/boc"
	"github.com/tonkeeper/tongo/tlb"
	"github.com/tonkeeper/tongo/ton"
	"go.uber.org/zap"
)

func (t *Tracker) collectActionsBlackTicketPurchasedInternal(userAddress string, lastBlackTicketPurchasedLtByUser int64, raffleStartedLt int64) ([]*storage.UserAction, error) {
	logger.Debug("black ticket purchased: processing collect pendingActions")

	var pendingActions = make([]*storage.UserAction, 0)

	var transactionLt int64 = 0
	var maxTransactionLt int64 = 0
	var beforeLt int64 = 0

	blackTicketCollectionAccountID, err := ton.ParseAccountID(t.blackTicketCollectionAddress)
	if err != nil {
		logger.Fatal("black ticket purchased: collection address is invalid", zap.Error(err))
		return nil, errors.New("black ticket collection address is invalid")
	}

	for {
		accountTracesResult, err := infinityRateLimitRetry(
			func() (*tonapi.TraceIDs, error) {
				return t.client.GetAccountTraces(t.ctx, tonapi.GetAccountTracesParams{
					AccountID: userAddress,
					Limit:     tonapi.NewOptInt(GlobalLimitWindowSize),
					BeforeLt: tonapi.OptInt64{
						Value: beforeLt,
						Set:   beforeLt > 0,
					},
				})
			},
		)

		if err != nil {
			logger.Fatal("black ticket purchased: cannot get source account info, skip")
			return nil, err
		}

		for _, traceID := range accountTracesResult.GetTraces() {
			logger.Debug("black ticket purchased: collect trace details... iteration", zap.String("trace id", traceID.GetID()))
			trace, err := infinityRateLimitRetry(
				func() (*tonapi.Trace, error) {
					return t.client.GetTrace(t.ctx, tonapi.GetTraceParams{TraceID: traceID.GetID()})
				},
			)

			if err != nil {
				logger.Debug("black ticket purchased: collect trace details... failed", zap.Error(err))
				break
			}

			transactionLt = trace.Transaction.GetLt()
			maxTransactionLt = max(maxTransactionLt, transactionLt)

			if transactionLt <= lastBlackTicketPurchasedLtByUser {
				logger.Debug("black ticket purchased: last transaction at reached")
				break
			}

			beforeLt = walkTracesBlackTicketPurchased(trace, func(inner *tonapi.Trace) {
				logger.Debug("black ticket purchased: found trace", zap.Int64("transaction at", transactionLt))

				transactionLt = inner.Transaction.GetLt()
				maxTransactionLt = max(maxTransactionLt, transactionLt)

				if transactionLt <= lastBlackTicketPurchasedLtByUser {
					return
				}

				transactionHash, address, ok := t.processBlackTicketPurchasedTrace(inner, &blackTicketCollectionAccountID)
				if ok && transactionLt > lastBlackTicketPurchasedLtByUser {
					logger.Debug("black ticket purchased: append action", zap.String("address", address))

					pendingActions = append(pendingActions, &storage.UserAction{
						ActionType:      storage.BlackTicketPurchasedActionType,
						Address:         address,
						TransactionLt:   transactionLt,
						TransactionHash: transactionHash,
					})
				} else {
					logger.Debug("black ticket purchased: trace cannot be processed, skip")
				}
			}, lastBlackTicketPurchasedLtByUser, raffleStartedLt)

			if beforeLt < raffleStartedLt {
				logger.Debug("black ticket purchased: raffle start time reached, finalize traces results...")
				break
			}

			logger.Debug("black ticket purchased: process user account trace... iteration done")
		}

		if len(accountTracesResult.GetTraces()) < GlobalLimitWindowSize || beforeLt <= raffleStartedLt || transactionLt <= lastBlackTicketPurchasedLtByUser {
			logger.Debug("black ticket purchased: out of trace, finalize traces results...")
			break
		}

		logger.Debug("black ticket purchased: collect user account traces... iteration done")
	}

	if maxTransactionLt > lastBlackTicketPurchasedLtByUser {
		pendingUserActionTouch := &storage.UserActionTouch{
			ActionType:    storage.BlackTicketPurchasedActionType,
			Address:       userAddress,
			TransactionLt: maxTransactionLt,
		}
		err = t.storage.UpdateUserActionTouch(pendingUserActionTouch)
		if err != nil {
			logger.Fatal("black ticket purchased: failed to update last action transaction state")
			return nil, err
		}
	}

	return pendingActions, nil
}

func (t *Tracker) collectBlackTicketPurchasedActions(raffleStartedAt int64) error {

	actions := make([]*storage.UserAction, 0)
	candidateAddressesActions, err := t.storage.GetUserActions(storage.CandidateRegistrationActionType)
	if err != nil {
		panic("failed to get candidate registration actions")
	}

	for _, candidateAddressAction := range candidateAddressesActions {
		logger.Debug("get latest black ticket purchased at")
		lastBlackTicketPurchasedByUserAt, err := t.storage.GetUserActionTouchByAddress(storage.BlackTicketPurchasedActionType, candidateAddressAction.Address)
		if err != nil {
			panic("failed to get max black ticket purchased at")
		}

		pendingActions, err := t.collectActionsBlackTicketPurchasedInternal(candidateAddressAction.Address, lastBlackTicketPurchasedByUserAt, raffleStartedAt)
		if err != nil {
			return err
		}

		actions = append(actions, pendingActions...)
	}

	if len(actions) > 0 {
		err = t.storage.UpdateUserActions(actions)
		if err != nil {
			panic("failed to update pending black ticket purchased actions")
		}
	}

	return nil
}

func walkTracesBlackTicketPurchased(trace *tonapi.Trace, callback func(*tonapi.Trace), lastBlackTicketPurchasedAt int64, raffleStartedAt int64) int64 {
	if trace == nil {
		return math.MaxInt64
	}

	callback(trace)

	transactionLt := trace.Transaction.GetLt()
	for i := range trace.Children {
		if transactionLt < raffleStartedAt || transactionLt < lastBlackTicketPurchasedAt {
			break
		}

		transactionLt = min(transactionLt, walkTracesBlackTicketPurchased(&trace.Children[i], callback, lastBlackTicketPurchasedAt, raffleStartedAt))
	}

	return transactionLt
}

func (t *Tracker) processBlackTicketPurchasedTrace(trace *tonapi.Trace, blackTicketCollectionAccountID *ton.AccountID) (string, string, bool) {
	nftTransferOpCode := tonapi.OptString{Value: "0x5fcc3d14", Set: true}

	message, ok := trace.Transaction.GetInMsg().Get()
	if !ok {
		logger.Debug("black ticket purchased: missing incoming message... skip")
		return "", "", false
	}

	if message.OpCode == nftTransferOpCode {
		sourceAccountID, ok := message.Source.Get()
		if !ok {
			logger.Debug("black ticket purchased: cannot get message source address... skip")
			return "", "", false
		}
		sourceAccount, err := infinityRateLimitRetry(
			func() (*tonapi.Account, error) {
				return t.client.GetAccount(t.ctx, tonapi.GetAccountParams{AccountID: sourceAccountID.Address})
			},
		)
		if err != nil {
			logger.Debug("black ticket purchased: cannot get source account info... skip")
			return "", "", false
		}

		if slices.Contains(sourceAccount.GetMethods, "get_sale_data") {
			saleDataResult, err := infinityRateLimitRetry(
				func() (*tonapi.MethodExecutionResult, error) {
					return t.client.ExecGetMethodForBlockchainAccount(t.ctx, tonapi.ExecGetMethodForBlockchainAccountParams{AccountID: sourceAccount.Address, MethodName: "get_sale_data"})
				},
			)
			if err != nil {
				logger.Debug("black ticket purchased: cannot execute get_sale_data method... skip")
				return "", "", false
			}

			marketplaceAddressCellOpt := saleDataResult.GetStack()[3].GetCell() // marketplace address
			marketplaceAddressCellString, ok := marketplaceAddressCellOpt.Get()
			if !ok {
				logger.Debug("black ticket purchased: invalid GetGems get_sale_data output... skip")
				return "", "", false
			}

			marketPlaceAddressCell, err := boc.DeserializeBocHex(marketplaceAddressCellString)
			if err != nil {
				logger.Debug("black ticket purchased: failed to deserialize marketplace boc hex... skip")
				return "", "", false
			}

			var marketplaceAddress tlb.MsgAddress
			err = tlb.Unmarshal(marketPlaceAddressCell[0], &marketplaceAddress)
			if err != nil {
				logger.Debug("black ticket purchased: failed to read marketplace address due to invalid tlb scheme... skip")
				return "", "", false
			}

			marketplaceAccountID, err := tongo.AccountIDFromTlb(marketplaceAddress)
			if marketplaceAccountID == nil || err != nil {
				logger.Debug("black ticket purchased: invalid marketplace address... skip")
				return "", "", false
			}

			if marketplaceAccountID.ToRaw() != MarketplaceAddressRaw {
				logger.Debug("black ticket purchased: purchase from not getgems marketplace, skip")
				return "", "", false
			}
		} else {
			return "", "", false
		}

		inMessageDestination, ok := message.Destination.Get()
		if ok {
			itemResult, err := infinityRateLimitRetry(
				func() (*tonapi.NftItem, error) {
					return t.client.GetNftItemByAddress(t.ctx, tonapi.GetNftItemByAddressParams{
						AccountID: inMessageDestination.Address,
					})
				},
			)

			if err != nil {
				logger.Debug("black ticket purchased: cannot get nft item information... skip")
				return "", "", false
			}

			collectionValue, ok := itemResult.GetCollection().Get()
			if !ok {
				logger.Debug("black ticket purchased: could not extract item collection value... skip")
				return "", "", false
			}

			if collectionValue.Address != blackTicketCollectionAccountID.ToRaw() {
				logger.Debug("black ticket purchased: black ticket collection address not matched... skip")
				return "", "", false
			}
		}

		if trace.Transaction.Success {
			body, err := boc.DeserializeBocHex(message.GetRawBody().Value)
			if err != nil {
				logger.Debug("black ticket purchased: failed to deserialize new owner boc hex... skip")
				return "", "", false
			}

			bodyCell := body[0]
			err = bodyCell.Skip(32)
			if err != nil {
				logger.Debug("black ticket purchased: failed to skip op code... skip")
				return "", "", false
			}

			err = bodyCell.Skip(64)
			if err != nil {
				logger.Debug("black ticket purchased: failed to skip query id... skip")
				return "", "", false
			}

			var newOwnerAddress tlb.MsgAddress
			err = tlb.Unmarshal(bodyCell, &newOwnerAddress)
			if err != nil {
				logger.Debug("black ticket purchased: failed to read new owner address due to invalid tlb scheme... skip")
				return "", "", false
			}

			userAccountID, err := tongo.AccountIDFromTlb(newOwnerAddress)
			if userAccountID == nil || err != nil {
				logger.Debug("black ticket purchased: invalid user account address... skip")
				return "", "", false
			}

			return message.GetHash(), userAccountID.ToHuman(true, false), true
		}
	}

	return "", "", false
}
