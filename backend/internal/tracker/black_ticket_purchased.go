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

func (t *Tracker) collectActionsBlackTicketPurchasedInternal(userAddress string, lastBlackTicketPurchasedLtByUser int64, raffleDeployedAt int64) ([]*storage.UserAction, error) {
	logger.Debug("black ticket purchased: processing collect actions")

	var actions = make([]*storage.UserAction, 0)

	var transactionLt int64 = 0
	var transactionUnixTime int64 = 0
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
			logger.Fatal("black ticket purchased: cannot get source account info, ...exiting", zap.Error(err))
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

			transactionLt = trace.Transaction.Lt
			transactionUnixTime = trace.Transaction.Utime
			maxTransactionLt = max(maxTransactionLt, transactionLt)

			if transactionLt <= lastBlackTicketPurchasedLtByUser {
				logger.Debug("black ticket purchased: last transaction logic time reached")
				break
			}

			beforeLt = walkTracesBlackTicketPurchased(trace, func(inner *tonapi.Trace) {
				transactionLt = inner.Transaction.Lt
				transactionUnixTime = inner.Transaction.Utime
				maxTransactionLt = max(maxTransactionLt, transactionLt)

				transactionHash, processedUserAddress, processedTicketAddress, ok := t.processBlackTicketPurchasedTrace(inner, &blackTicketCollectionAccountID)
				if ok && transactionLt > lastBlackTicketPurchasedLtByUser {
					logger.Debug("black ticket purchased: append action", zap.String("user address", processedUserAddress), zap.String("ticket address", processedTicketAddress))

					actions = append(actions, &storage.UserAction{
						ActionType:          storage.BlackTicketPurchasedActionType,
						UserAddress:         processedUserAddress,
						Address:             processedTicketAddress,
						TransactionLt:       transactionLt,
						TransactionHash:     transactionHash,
						TransactionUnixTime: transactionUnixTime,
					})
				} else {
					logger.Debug("black ticket purchased: no need to process, skip")
				}
			}, lastBlackTicketPurchasedLtByUser, raffleDeployedAt)

			logger.Debug("black ticket purchased: process user account trace... iteration done")
		}

		if len(accountTracesResult.GetTraces()) < GlobalLimitWindowSize || beforeLt < raffleDeployedAt || transactionLt <= lastBlackTicketPurchasedLtByUser {
			logger.Debug("black ticket purchased: out of trace, finalize traces results...")
			break
		}

		logger.Debug("black ticket purchased: collect user account traces... iteration done")
	}

	if maxTransactionLt > lastBlackTicketPurchasedLtByUser {
		pendingUserActionTouch := &storage.UserActionTouch{
			ActionType:    storage.BlackTicketPurchasedActionType,
			UserAddress:   userAddress,
			TransactionLt: maxTransactionLt,
		}
		err = t.storage.UpdateUserActionTouch(pendingUserActionTouch)
		if err != nil {
			logger.Fatal("black ticket purchased: failed to update last action transaction state")
			return nil, err
		}
	}

	return actions, nil
}

func (t *Tracker) collectBlackTicketPurchasedActions(raffleDeployedAt int64) error {

	actions := make([]*storage.UserAction, 0)
	candidateAddressesActions, err := t.storage.GetUserActions(storage.CandidateRegistrationActionType)

	if err != nil {
		logger.Fatal("black ticket purchased: failed to get user actions", zap.Error(err))
		panic("failed to get candidate registration actions")
	}

	for _, candidateAddressAction := range candidateAddressesActions {
		logger.Debug("get latest black ticket purchased at")
		lastBlackTicketPurchasedByUserAt, err := t.storage.GetUserActionTouchByAddress(storage.BlackTicketPurchasedActionType, candidateAddressAction.UserAddress)
		if err != nil {
			logger.Fatal("failed to get last black ticket purchased at", zap.Error(err))
			panic("failed to get max black ticket purchased at")
		}

		pendingActions, err := t.collectActionsBlackTicketPurchasedInternal(candidateAddressAction.UserAddress, lastBlackTicketPurchasedByUserAt, raffleDeployedAt)
		if err != nil {
			logger.Fatal("black ticket purchased at", zap.Error(err))
			return err
		}

		actions = append(actions, pendingActions...)
	}

	if len(actions) > 0 {
		err = t.storage.UpdateUserActions(actions)
		if err != nil {
			logger.Fatal("black ticket purchased at", zap.Error(err))
			panic("failed to update pending black ticket purchased actions")
		}
	}

	return nil
}

func walkTracesBlackTicketPurchased(trace *tonapi.Trace, callback func(*tonapi.Trace), lastBlackTicketPurchasedAt int64, raffleDeployedAt int64) int64 {
	if trace == nil {
		return math.MaxInt64
	}

	callback(trace)

	transactionLt := trace.Transaction.GetLt()
	for i := range trace.Children {
		if transactionLt < raffleDeployedAt || transactionLt < lastBlackTicketPurchasedAt {
			break
		}

		transactionLt = min(transactionLt, walkTracesBlackTicketPurchased(&trace.Children[i], callback, lastBlackTicketPurchasedAt, transactionLt))
	}

	return transactionLt
}

func (t *Tracker) processBlackTicketPurchasedTrace(trace *tonapi.Trace, blackTicketCollectionAccountID *ton.AccountID) (string, string, string, bool) {
	nftTransferOpCode := "0x5fcc3d14"

	message, ok := trace.Transaction.GetInMsg().Get()
	if !ok {
		logger.Debug("black ticket purchased: missing incoming message... skip")
		return "", "", "", false
	}

	if !trace.Transaction.Success {
		logger.Debug("black ticket purchased: ignore unsuccessful incoming messages... skip")
		return "", "", "", false
	}

	if !message.OpCode.IsSet() || message.OpCode.Value != nftTransferOpCode {
		logger.Debug("black ticket purchased: not NFT transfer op code... skip")
		return "", "", "", false
	}

	sourceAccountID, ok := message.Source.Get()
	if !ok {
		logger.Debug("black ticket purchased: cannot get message source address... skip")
		return "", "", "", false
	}
	sourceAccount, err := infinityRateLimitRetry(
		func() (*tonapi.Account, error) {
			return t.client.GetAccount(t.ctx, tonapi.GetAccountParams{AccountID: sourceAccountID.Address})
		},
	)
	if err != nil {
		logger.Debug("black ticket purchased: cannot get source account info... skip")
		return "", "", "", false
	}

	if !slices.Contains(sourceAccount.GetMethods, "get_sale_data") {
		logger.Debug("black ticket purchased: account contract does not provide sale data method... skip")
		return "", "", "", false
	}

	saleDataResult, err := infinityRateLimitRetry(
		func() (*tonapi.MethodExecutionResult, error) {
			return t.client.ExecGetMethodForBlockchainAccount(t.ctx, tonapi.ExecGetMethodForBlockchainAccountParams{AccountID: sourceAccount.Address, MethodName: "get_sale_data"})
		},
	)
	if err != nil {
		logger.Debug("black ticket purchased: cannot execute get_sale_data method... skip")
		return "", "", "", false
	}

	// marketplace address
	saleDataMarketplaceAddressCellString, ok := saleDataResult.GetStack()[3].GetCell().Get()
	if !ok {
		logger.Debug("black ticket purchased: invalid GetGems get_sale_data output... skip")
		return "", "", "", false
	}

	saleDataMarketplaceAddressCell, err := boc.DeserializeBocHex(saleDataMarketplaceAddressCellString)
	if err != nil {
		logger.Debug("black ticket purchased: failed to deserialize marketplace boc hex... skip")
		return "", "", "", false
	}

	var saleDataMarketplaceAddress tlb.MsgAddress
	err = tlb.Unmarshal(saleDataMarketplaceAddressCell[0], &saleDataMarketplaceAddress)
	if err != nil {
		logger.Debug("black ticket purchased: failed to read marketplace address due to invalid tlb scheme... skip")
		return "", "", "", false
	}

	saleDataMarketplaceAccountID, err := tongo.AccountIDFromTlb(saleDataMarketplaceAddress)
	if saleDataMarketplaceAccountID == nil || err != nil {
		logger.Debug("black ticket purchased: invalid marketplace address... skip")
		return "", "", "", false
	}

	if saleDataMarketplaceAccountID.ToRaw() != MarketplaceAddressRaw {
		logger.Debug("black ticket purchased: purchase from not getgems marketplace, skip")
		return "", "", "", false
	}

	// owner address
	saleDataOwnerAddressCellOpt := saleDataResult.GetStack()[5].GetCell()
	saleDataOwnerAddressCellString, ok := saleDataOwnerAddressCellOpt.Get()
	if !ok {
		logger.Debug("black ticket purchased: invalid GetGems get_sale_data output... skip")
		return "", "", "", false
	}

	saleDataOwnerAddressCell, err := boc.DeserializeBocHex(saleDataOwnerAddressCellString)
	if err != nil {
		logger.Debug("black ticket purchased: failed to deserialize marketplace boc hex... skip")
		return "", "", "", false
	}

	var saleDataOwnerAddress tlb.MsgAddress
	err = tlb.Unmarshal(saleDataOwnerAddressCell[0], &saleDataOwnerAddress)
	if err != nil {
		logger.Debug("black ticket purchased: failed to read marketplace address due to invalid tlb scheme... skip")
		return "", "", "", false
	}

	saleDataOwnerAccountID, err := tongo.AccountIDFromTlb(saleDataOwnerAddress)
	if saleDataOwnerAccountID == nil || err != nil {
		logger.Debug("black ticket purchased: invalid owner address... skip")
		return "", "", "", false
	}

	inMessageDestination, ok := message.Destination.Get()
	if !ok {
		logger.Debug("black ticket purchased: destination account address missing... skip")
		return "", "", "", false
	}

	inMessageDestinationAccountID, err := ton.ParseAccountID(inMessageDestination.Address)
	if err != nil {
		logger.Debug("black ticket purchased: failed to parse destination account address... skip")
		return "", "", "", false
	}

	itemResult, err := infinityRateLimitRetry(
		func() (*tonapi.NftItem, error) {
			return t.client.GetNftItemByAddress(t.ctx, tonapi.GetNftItemByAddressParams{
				AccountID: inMessageDestination.Address,
			})
		},
	)

	if err != nil {
		logger.Debug("black ticket purchased: cannot get nft item information... skip")
		return "", "", "", false
	}

	collectionValue, ok := itemResult.GetCollection().Get()
	if !ok {
		logger.Debug("black ticket purchased: could not extract item collection value... skip")
		return "", "", "", false
	}

	if collectionValue.Address != blackTicketCollectionAccountID.ToRaw() {
		logger.Debug("black ticket purchased: black ticket collection address not matched... skip")
		return "", "", "", false
	}

	body, err := boc.DeserializeBocHex(message.GetRawBody().Value)
	if err != nil {
		logger.Debug("black ticket purchased: failed to deserialize new owner boc hex... skip")
		return "", "", "", false
	}

	bodyCell := body[0]
	err = bodyCell.Skip(32)
	if err != nil {
		logger.Debug("black ticket purchased: failed to skip op code... skip")
		return "", "", "", false
	}

	err = bodyCell.Skip(64)
	if err != nil {
		logger.Debug("black ticket purchased: failed to skip query id... skip")
		return "", "", "", false
	}

	var newOwnerAddress tlb.MsgAddress
	err = tlb.Unmarshal(bodyCell, &newOwnerAddress)
	if err != nil {
		logger.Debug("black ticket purchased: failed to read new owner address due to invalid tlb scheme... skip")
		return "", "", "", false
	}

	userAccountID, err := tongo.AccountIDFromTlb(newOwnerAddress)
	if userAccountID == nil || err != nil {
		logger.Debug("black ticket purchased: invalid user account address... skip")
		return "", "", "", false
	}

	if saleDataOwnerAccountID.ToRaw() == userAccountID.ToRaw() {
		logger.Debug("black ticket purchased: it is not purchase, it is sale cancellation, skip")
		return "", "", "", false
	}

	return trace.Transaction.Hash, userAccountID.ToHuman(true, false), inMessageDestinationAccountID.ToHuman(true, false), true
}
