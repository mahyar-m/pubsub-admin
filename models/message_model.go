package models

import (
	"database/sql"
	"encoding/json"

	"example/pubsub_manager/db"

	"cloud.google.com/go/pubsub"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

type MessageModel struct {
}

type MessageRow struct {
	Uuid            sql.NullString
	MessageId       sql.NullString
	Subscription    sql.NullString
	Data            sql.NullString
	Attribute       sql.NullString
	PublishTime     sql.NullTime
	DeliveryAttempt sql.NullInt32
	OrderingKey     sql.NullString
}

func (messageModel *MessageModel) Select(config db.MysqlConfig, selectQuery string) ([]MessageRow, error) {
	db, err := sql.Open("mysql", config.GetConnString())
	if err != nil {
		return nil, err
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(selectQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messageRows []MessageRow

	for rows.Next() {
		var message MessageRow
		if err := rows.Scan(&message.Uuid, &message.MessageId, &message.Subscription, &message.Data, &message.Attribute, &message.PublishTime, &message.DeliveryAttempt, &message.OrderingKey); err != nil {
			return nil, err
		}
		messageRows = append(messageRows, message)
	}

	return messageRows, nil
}

func (messageModel *MessageModel) Insert(config db.MysqlConfig, subID string, msg *pubsub.Message) error {
	db, err := sql.Open("mysql", config.GetConnString())
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return err
	}

	stmtIns, err := db.Prepare("INSERT INTO message (uuid, message_id, subscription, data, attribute, publish_time, delivery_attempt, ordering_key) VALUES(?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmtIns.Close()

	var attributes []byte = nil
	if msg.Attributes != nil {
		attributes, _ = json.Marshal(msg.Attributes)
	}
	_, err = stmtIns.Exec(uuid.New().String(), msg.ID, subID, msg.Data, attributes, msg.PublishTime, msg.DeliveryAttempt, msg.OrderingKey)
	if err != nil {
		return err
	}

	return nil
}

func (messageModel *MessageModel) IsDuplicate(config db.MysqlConfig, subID string, msg *pubsub.Message) (bool, error) {
	db, err := sql.Open("mysql", config.GetConnString())
	if err != nil {
		return false, err
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return false, err
	}

	row := db.QueryRow("SELECT count(uuid) FROM message WHERE message_id = ? AND subscription = ?", msg.ID, subID)

	var count int
	if err := row.Scan(&count); err != nil {
		return false, err
	}

	return count > 0, nil
}
