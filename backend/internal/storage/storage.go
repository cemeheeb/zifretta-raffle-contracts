package storage

type Storage interface {
	// user action
	GetUserActions(actionType ActionType) ([]*UserAction, error)
	UpdateUserActions(actions []*UserAction) error

	// user action touch
	GetUserActionTouch(actionType ActionType) (int64, error)
	GetUserActionTouchByAddress(actionType ActionType, address string) (int64, error)
	UpdateUserActionTouch(actionTouch *UserActionTouch) error

	// pending user action
	GetPendingCandidateRegistrationActions() ([]*UserAction, error)
	GetPendingParticipantRegistrationActions() ([]*UserAction, error)
	GetPendingWhiteTicketMintedActions() ([]*UserAction, error)
	GetPendingBlackTicketPurchasedActions() ([]*UserAction, error)

	// user action
	GetUserStatusByAddress(address string) (*UserStatus, error)
	GetUserStatusesByAddresses(addresses []string) ([]*UserStatus, error)
	GetUserStatusesByConditionsReached() ([]*UserStatus, error)
	UpdateUserStatus(action *UserStatus) error
	UpdateUserStatuses(action []*UserStatus) error
}

type ActionType = string

const (
	CandidateRegistrationActionType   ActionType = "CandidateRegistrationActionType"
	ParticipantRegistrationActionType ActionType = "ParticipantRegistrationActionType"
	WhiteTicketMintedActionType       ActionType = "WhiteTicketMintedActionType"
	BlackTicketPurchasedActionType    ActionType = "BlackTicketPurchasedActionType"
)
