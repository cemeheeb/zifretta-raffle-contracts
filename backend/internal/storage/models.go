package storage

type UserStatus struct {
	UserAddress                     string `gorm:"primaryKey"`
	WhiteTicketMinted               uint8  `gorm:"default:0"`
	BlackTicketPurchased            uint8  `gorm:"default:0"`
	CandidateRegistrationLt         int64  `gorm:"not null"`
	WhiteTicketMintedProcessedLt    int64  `gorm:"default:0"`
	BlackTicketPurchasedProcessedLt int64  `gorm:"default:0"`
	ParticipantRegistrationLt       int64  `gorm:"default:0"`
	LastDeployedUnixTime            int64  `gorm:"default:0"`
}

type UserAction struct {
	ID                  int64      `gorm:"primaryKey"`
	ActionType          ActionType `gorm:"uniqueIndex:idx_unique_action"`
	UserAddress         string     `gorm:"uniqueIndex:idx_unique_action"`
	Address             string     `gorm:"uniqueIndex:idx_unique_action"`
	TransactionHash     string     `gorm:"not null"`
	TransactionLt       int64      `gorm:"not null"`
	TransactionUnixTime int64      `gorm:"not null"`
}

type UserActionTouch struct {
	ActionType    ActionType `gorm:"primaryKey"`
	UserAddress   string     `gorm:"primaryKey"`
	TransactionLt int64      `gorm:"not null"`
}
