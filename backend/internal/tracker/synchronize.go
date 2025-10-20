package tracker

import (
	"backend/internal/blockchain"
	"backend/internal/logger"
	"backend/internal/storage"
	"time"

	"github.com/tonkeeper/tongo/ton"
)

func (t *Tracker) synchronizePendingCandidateRegistrationActions() error {
	pendingActions, err := t.storage.GetPendingCandidateRegistrationActions()
	if err != nil {
		logger.Debug("cannot get pending candidate registration actions, exiting...")
		return err
	}

	var userStatuses = make([]*storage.UserStatus, len(pendingActions))
	for i, action := range pendingActions {
		userStatuses[i] = &storage.UserStatus{
			UserAddress:             action.UserAddress,
			CandidateRegistrationLt: action.TransactionLt,
		}
	}

	err = t.storage.UpdateUserStatuses(userStatuses)
	if err != nil {
		logger.Debug("cannot update user statuses action, exiting...")
		return err
	}

	return nil
}

func (t *Tracker) synchronizePendingWhiteTicketMintedActions() error {
	pendingActions, err := t.storage.GetPendingWhiteTicketMintedActions()
	if err != nil {
		logger.Debug("cannot get pending white ticket minted actions, exiting...")
		return err
	}

	addressPendingActionMap := make(map[string]*storage.UserAction)
	for _, action := range pendingActions {
		addressPendingActionMap[action.UserAddress] = action
	}

	addressQuantityMap := make(map[string]uint8)
	for _, action := range pendingActions {
		addressQuantityMap[action.UserAddress] = addressQuantityMap[action.UserAddress] + 1
	}

	addresses := make([]string, 0, len(addressQuantityMap))
	for address := range addressQuantityMap {
		addresses = append(addresses, address)
	}

	userStatuses, err := t.storage.GetUserStatusesByAddresses(addresses)
	if err != nil {
		return err
	}

	for _, userStatus := range userStatuses {
		userStatus.WhiteTicketMinted = min(userStatus.WhiteTicketMinted+addressQuantityMap[userStatus.UserAddress], 2)
		userStatus.WhiteTicketMintedProcessedLt = addressPendingActionMap[userStatus.UserAddress].TransactionLt

		err = t.storage.UpdateUserStatus(userStatus)
		if err != nil {
			logger.Debug("Cannot update user statuses, exiting...")
			return err
		}

		err := t.invalidateConditions(userStatus)
		if err != nil {
			logger.Debug("cannot invalidate conditions, exiting...")
			return err
		}
	}

	return nil
}

func (t *Tracker) synchronizePendingBlackTicketPurchasedActions() error {
	pendingActions, err := t.storage.GetPendingBlackTicketPurchasedActions()
	if err != nil {
		logger.Debug("cannot get pending black ticket purchased actions, exiting...")
		return err
	}

	addressPendingActionMap := make(map[string]*storage.UserAction)
	for _, action := range pendingActions {
		addressPendingActionMap[action.UserAddress] = action
	}

	addressQuantityMap := make(map[string]uint8)
	for _, action := range pendingActions {
		addressQuantityMap[action.UserAddress] = addressQuantityMap[action.UserAddress] + 1
	}

	addresses := make([]string, 0, len(addressQuantityMap))
	for address := range addressQuantityMap {
		addresses = append(addresses, address)
	}

	userStatuses, err := t.storage.GetUserStatusesByAddresses(addresses)
	if err != nil {
		return err
	}

	for _, userStatus := range userStatuses {
		userStatus.BlackTicketPurchased = min(userStatus.BlackTicketPurchased+addressQuantityMap[userStatus.UserAddress], 2)
		userStatus.BlackTicketPurchasedProcessedLt = addressPendingActionMap[userStatus.UserAddress].TransactionLt

		err = t.storage.UpdateUserStatus(userStatus)
		if err != nil {
			logger.Debug("synchronize black ticket purchased: cannot update user statuses, exiting...")
			return err
		}

		err := t.invalidateConditions(userStatus)
		if err != nil {
			logger.Debug("synchronize black ticket purchased: cannot invalidate conditions, exiting...")
			return err
		}
	}

	return nil
}

func (t *Tracker) synchronizePendingParticipantRegistrationActions() error {
	pendingActions, err := t.storage.GetPendingParticipantRegistrationActions()
	if err != nil {
		logger.Debug("cannot get pending participant registration actions, exiting...")
		return err
	}

	var userStatuses = make([]*storage.UserStatus, len(pendingActions))
	for i, action := range pendingActions {
		userStatuses[i] = &storage.UserStatus{
			UserAddress:               action.UserAddress,
			ParticipantRegistrationLt: action.TransactionLt,
		}
	}

	err = t.storage.UpdateUserStatuses(userStatuses)
	if err != nil {
		logger.Debug("cannot update user statuses action, exiting...")
		return err
	}

	return nil
}

func (t *Tracker) invalidateConditions(status *storage.UserStatus) error {

	raffleAccountID, err := ton.ParseAccountID(t.raffleAddress)
	if err != nil {
		logger.Debug("invalidate conditions: cannot parse raffle account id, exiting...")
		return err
	}

	userAccountID, err := ton.ParseAccountID(status.UserAddress)
	if err != nil {
		logger.Debug("invalidate conditions: cannot parse user account id, exiting...")
		return err
	}

	if status.WhiteTicketMinted == 2 && status.BlackTicketPurchased == 2 && ((status.LastDeployedUnixTime + GlobalDeployedTimeout) < time.Now().Unix()) {
		err := t.sendSetConditions(raffleAccountID, userAccountID, status.WhiteTicketMinted, status.BlackTicketPurchased)
		if err != nil {
			logger.Debug("invalidate conditions: cannot send set conditions to blockchain, exiting...")
			return err
		}
		status.LastDeployedUnixTime = time.Now().Unix()
		err = t.storage.UpdateUserStatus(status)
		if err != nil {
			logger.Debug("invalidate conditions: cannot update user status, exiting...")
			return err
		}
	}

	return nil
}

func (t *Tracker) sendSetConditions(raffleAccountID ton.AccountID, userAccountID ton.AccountID, whiteTicketMinted uint8, blackTicketPurchased uint8) error {

	err := t.wallet.Send(t.ctx, blockchain.RaffleSetConditionMessage{
		AttachedTon:          5_000_000_0, // 0.05 ton
		RaffleAddress:        raffleAccountID,
		UserAddress:          userAccountID,
		WhiteTicketMinted:    whiteTicketMinted,
		BlackTicketPurchased: blackTicketPurchased,
	})

	if err != nil {
		return err
	}

	return nil
}

func (t *Tracker) synchronize() error {

	err := t.synchronizePendingCandidateRegistrationActions()
	if err != nil {
		return err
	}

	err = t.synchronizePendingParticipantRegistrationActions()
	if err != nil {
		return err
	}

	err = t.synchronizePendingWhiteTicketMintedActions()
	if err != nil {
		return err
	}

	err = t.synchronizePendingBlackTicketPurchasedActions()
	if err != nil {
		return err
	}

	return nil
}
