package storage

type UserStatus struct {
	Address                         string `gorm:"primaryKey"`
	WhiteTicketMinted               uint8  `gorm:"default:0"`
	BlackTicketPurchased            uint8  `gorm:"default:0"`
	CandidateRegistrationLt         int64  `gorm:"not null"`
	WhiteTicketMintedProcessedLt    int64  `gorm:"default:0"`
	BlackTicketPurchasedProcessedLt int64  `gorm:"default:0"`
	ParticipantRegistrationLt       int64  `gorm:"default:0"`
	LastDeployedUnixTime            int64  `gorm:"default:0"`
}

type UserAction struct {
	ID              int64      `gorm:"primaryKey"`
	ActionType      ActionType `gorm:"uniqueIndex:idx_action_type_address"`
	Address         string     `gorm:"uniqueIndex:idx_action_type_address"`
	TransactionHash string     `gorm:"not null"`
	TransactionLt   int64      `gorm:"not null"`
}

type UserActionTouch struct {
	ActionType    ActionType `gorm:"primaryKey"`
	Address       string     `gorm:"primaryKey"`
	TransactionLt int64      `gorm:"not null"`
}
