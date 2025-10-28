package tracker

import (
	"backend/internal/logger"
	"backend/internal/storage"
	"time"

	"github.com/tonkeeper/tongo/boc"
	"github.com/tonkeeper/tongo/tlb"
	"github.com/tonkeeper/tongo/ton"
	"github.com/tonkeeper/tongo/wallet"
	"go.uber.org/zap"
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
		userStatusNext := &storage.UserStatus{
			UserAddress:                     userStatus.UserAddress,
			WhiteTicketMinted:               min(userStatus.WhiteTicketMinted+addressQuantityMap[userStatus.UserAddress], 2),
			WhiteTicketMintedProcessedLt:    addressPendingActionMap[userStatus.UserAddress].TransactionLt,
			CandidateRegistrationLt:         userStatus.CandidateRegistrationLt,
			BlackTicketPurchasedProcessedLt: userStatus.BlackTicketPurchasedProcessedLt,
			ParticipantRegistrationLt:       userStatus.ParticipantRegistrationLt,
			LastDeployedUnixTime:            userStatus.LastDeployedUnixTime,
		}

		err = t.invalidateConditions(userStatus, userStatusNext)
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
		userStatusNext := &storage.UserStatus{
			UserAddress:                     userStatus.UserAddress,
			WhiteTicketMinted:               min(userStatus.BlackTicketPurchased+addressQuantityMap[userStatus.UserAddress], 2),
			WhiteTicketMintedProcessedLt:    userStatus.WhiteTicketMintedProcessedLt,
			CandidateRegistrationLt:         userStatus.CandidateRegistrationLt,
			BlackTicketPurchasedProcessedLt: addressPendingActionMap[userStatus.UserAddress].TransactionLt,
			ParticipantRegistrationLt:       userStatus.ParticipantRegistrationLt,
			LastDeployedUnixTime:            userStatus.LastDeployedUnixTime,
		}

		err := t.invalidateConditions(userStatus, userStatusNext)
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

func (t *Tracker) invalidateConditions(status *storage.UserStatus, statusNext *storage.UserStatus) error {

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

	logger.Info(
		"invalidate conditions",
		zap.Uint8("status.WhiteTicketMinted", status.WhiteTicketMinted),
		zap.Uint8("status.BlackTicketPurchased", status.BlackTicketPurchased),
		zap.Uint8("statusNext.WhiteTicketMinted", statusNext.WhiteTicketMinted),
		zap.Uint8("statusNext.BlackTicketPurchased", statusNext.BlackTicketPurchased),
	)

	if status.WhiteTicketMinted < statusNext.WhiteTicketMinted || status.BlackTicketPurchased < statusNext.BlackTicketPurchased {
		err := t.sendSetConditions(raffleAccountID, userAccountID, statusNext.WhiteTicketMinted, statusNext.BlackTicketPurchased)
		if err != nil {
			logger.Debug("invalidate conditions: cannot send set conditions to blockchain, exiting...")
			return err
		}

		status.LastDeployedUnixTime = time.Now().Unix()
		err = t.storage.UpdateUserStatus(statusNext)
		if err != nil {
			logger.Debug("invalidate conditions: cannot update user status, exiting...")
			return err
		}
	}

	return nil
}

func (t *Tracker) sendSetConditions(raffleAccountID ton.AccountID, userAccountID ton.AccountID, whiteTicketMinted uint8, blackTicketPurchased uint8) error {
	logger.Debug("sending setting conditions to blockchain...")

	cell := boc.NewCell()

	if err := cell.WriteUint(0x13370011, 32); err != nil {

		return err
	}

	if err := tlb.Marshal(cell, userAccountID.ToMsgAddress()); err != nil {
		return err
	}

	if err := cell.WriteUint(uint64(whiteTicketMinted), 8); err != nil {
		return err
	}

	if err := cell.WriteUint(uint64(blackTicketPurchased), 8); err != nil {
		return err
	}

	if err := cell.WriteUint(0, 240); err != nil {
		return err
	}

	message := wallet.Message{
		Amount:  5_000_000_0,
		Address: raffleAccountID,
		Bounce:  true,
		Mode:    wallet.DefaultMessageMode,
		Body:    cell,
	}

	_, err := t.wallet.SendV2(t.ctx, 60*time.Second, message)

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
