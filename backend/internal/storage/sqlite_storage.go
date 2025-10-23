package storage

import (
	"backend/internal/logger"
	"database/sql"
	"log"

	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SqliteStorage struct {
	db *gorm.DB
}

func NewSqliteStorage() *SqliteStorage {

	logger.Debug("initializing database...")
	db, err := gorm.Open(sqlite.Open("persistent.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	err = db.AutoMigrate(
		&UserAction{},
		&UserActionTouch{},
		&UserStatus{},
	)

	if err != nil {
		panic(err)
	}

	return &SqliteStorage{
		db: db,
	}
}

func (s *SqliteStorage) GetUserActions(actionType ActionType) ([]*UserAction, error) {

	var actions []*UserAction
	err := s.db.Where("action_type = ?", actionType).Find(&actions).Error

	if err != nil {
		return nil, err
	}

	return actions, nil
}

func (s *SqliteStorage) UpdateUserActions(actions []*UserAction) error {
	logger.Debug("update pending user actions...")

	if len(actions) == 0 {
		logger.Debug("no pending user actions to persist")
		return nil
	}

	err := s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "action_type"}, {Name: "user_address"}, {Name: "address"}},
		DoUpdates: clause.AssignmentColumns([]string{"transaction_lt", "transaction_hash"}),
	}).CreateInBatches(actions, 100).Error

	if err != nil {
		return err
	}

	logger.Debug("update pending user actions...done")
	return nil
}

func (s *SqliteStorage) GetUserActionTouch(actionType ActionType) (int64, error) {
	logger.Debug("getting last action transaction...")

	var transactionLt int64
	err := s.db.Raw(`
		select coalesce(max(transaction_lt), 0) as transaction_lt
		from user_action_touches
		where action_type = ?
	`, actionType).Scan(&transactionLt).Error

	if err != nil {
		return 0, err
	}

	logger.Debug("getting last action transaction... done", zap.Int64("transactionLt", transactionLt))
	return transactionLt, nil
}

func (s *SqliteStorage) GetUserActionTouchByAddress(actionType ActionType, address string) (int64, error) {
	logger.Debug("getting last action transaction by user", zap.String("action type", actionType), zap.String("address", address))

	var transactionLt int64
	err := s.db.Raw(`
		select coalesce(max(transaction_lt), 0) as transaction_lt
		from user_action_touches
		where action_type = ? and user_address = ?
	`, actionType, address).Scan(&transactionLt).Error

	if err != nil {
		return 0, err
	}

	logger.Debug("getting last action transaction by user... done")
	return transactionLt, nil
}

func (s *SqliteStorage) UpdateUserActionTouch(actionTouch *UserActionTouch) error {
	logger.Debug("updating pending action touch...")

	err := s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "action_type"}, {Name: "user_address"}},
		DoUpdates: clause.AssignmentColumns([]string{"transaction_lt"}),
	}).Create(&actionTouch).Error

	if err != nil {
		return err
	}

	logger.Debug("updating pending action touch...done")
	return nil
}

func (s *SqliteStorage) GetPendingCandidateRegistrationActions() ([]*UserAction, error) {
	logger.Debug("getting pending candidate registration actions...")

	rows, err := s.db.Raw(`
		select a.*
		from user_actions a
			left join user_statuses s on s.user_address = a.user_address
		where a.action_type = ? and s.user_address is null
	`, CandidateRegistrationActionType).Rows()
	if err != nil {
		logger.Fatal(err.Error())
		return nil, err
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			logger.Fatal(err.Error())
		}
	}(rows)

	var actions = make([]*UserAction, 0)
	for rows.Next() {
		var userAction UserAction

		if err := s.db.ScanRows(rows, &userAction); err != nil {
			log.Fatal(err)
			return nil, err
		}

		actions = append(actions, &userAction)
	}

	logger.Debug("getting pending candidate registration actions... done")
	return actions, nil
}

func (s *SqliteStorage) GetPendingParticipantRegistrationActions() ([]*UserAction, error) {
	logger.Debug("getting pending participant registration actions...")

	rows, err := s.db.Raw(`
		select a.*
		from user_actions a
			left join user_statuses s on s.user_address = a.user_address
		where a.action_type = ?
	`, ParticipantRegistrationActionType).Rows()
	if err != nil {
		logger.Fatal(err.Error())
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			logger.Fatal(err.Error())
		}
	}(rows)

	var actions = make([]*UserAction, 0)
	for rows.Next() {
		var userAction UserAction

		if err := s.db.ScanRows(rows, &userAction); err != nil {
			log.Fatal(err)
			return nil, err
		}

		actions = append(actions, &userAction)
	}

	logger.Debug("getting pending participant registration actions... done")
	return actions, nil
}

func (s *SqliteStorage) GetPendingWhiteTicketMintedActions() ([]*UserAction, error) {
	logger.Debug("getting pending white ticket minted actions...")

	rows, err := s.db.Raw(`
		select a.*
		from user_actions a
				 left join user_statuses s on s.user_address = a.user_address
		where a.action_type = ?
		  and (s.white_ticket_minted_processed_lt < a.transaction_lt or s.white_ticket_minted_processed_lt = 0)
	`, WhiteTicketMintedActionType).Rows()

	if err != nil {
		logger.Fatal(err.Error())
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			logger.Fatal(err.Error())
		}
	}(rows)

	var actions = make([]*UserAction, 0)
	for rows.Next() {
		var userAction UserAction

		if err := s.db.ScanRows(rows, &userAction); err != nil {
			log.Fatal(err)
			return nil, err
		}

		actions = append(actions, &userAction)
	}

	logger.Debug("getting pending white ticket minted actions... done")
	return actions, nil
}

func (s *SqliteStorage) GetPendingBlackTicketPurchasedActions() ([]*UserAction, error) {
	logger.Debug("getting pending black ticket purchased actions...")

	rows, err := s.db.Raw(`
		select a.*
		from user_actions a
			left join user_statuses s on s.user_address = a.user_address
		where a.action_type = ?
		  and (s.black_ticket_purchased_processed_lt > a.transaction_lt or s.black_ticket_purchased_processed_lt = 0)
	`, BlackTicketPurchasedActionType).Rows()

	if err != nil {
		logger.Fatal(err.Error())
		return nil, err
	}

	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			logger.Fatal(err.Error())
		}
	}(rows)

	var actions = make([]*UserAction, 0)
	for rows.Next() {
		var userAction UserAction

		if err := s.db.ScanRows(rows, &userAction); err != nil {
			log.Fatal(err)
			return nil, err
		}

		actions = append(actions, &userAction)
	}

	logger.Debug("getting pending black ticket purchased actions... done")
	return actions, nil
}

func (s *SqliteStorage) GetUserStatusByAddress(address string) (*UserStatus, error) {

	var userStatus UserStatus
	err := s.db.Where("user_address = ?", address).First(&userStatus).Error
	if err != nil {
		return nil, err
	}

	return &userStatus, nil
}

func (s *SqliteStorage) GetUserStatusesByAddresses(addresses []string) ([]*UserStatus, error) {

	var userStatuses []*UserStatus
	tx := s.db.Where("user_address in ?", addresses).Find(&userStatuses)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return userStatuses, nil
}

func (s *SqliteStorage) UpdateUserStatus(action *UserStatus) error {
	logger.Debug("updating user status...")

	tx := s.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_address"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"white_ticket_minted",
			"white_ticket_minted_processed_lt",
			"black_ticket_purchased",
			"black_ticket_purchased_processed_lt",
			"last_deployed_unix_time",
		}),
	}).Create(&action)
	if tx.Error != nil {
		logger.Fatal(tx.Error.Error())
		return tx.Error
	}

	logger.Debug("updating user status...done")
	return nil
}

func (s *SqliteStorage) UpdateUserStatuses(userStatuses []*UserStatus) error {
	logger.Debug("update user statuses...")

	if len(userStatuses) == 0 {
		logger.Debug("no user statuses to persist")
		return nil
	}

	err := s.db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_address"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"white_ticket_minted",
			"white_ticket_minted_processed_lt",
			"black_ticket_purchased",
			"black_ticket_purchased_processed_lt",
			"last_deployed_unix_time",
		}),
	}).CreateInBatches(userStatuses, 100).Error
	if err != nil {
		return err
	}

	logger.Debug("update user statuses... done")
	return nil
}
