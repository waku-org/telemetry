package telemetry

import (
	"database/sql"
	"time"
)

type WakuMessage struct {
	ID             int    `json:"id"`
	WalletAddress  string `json:"walletAddress"`
	PeerIDSender   string `json:"peerIdSender"`
	PeerIDReporter string `json:"peerIdReporter"`
	SequenceHash   string `json:"sequenceHash"`
	SequenceTotal  uint64 `json:"sequenceTotal"`
	SequenceIndex  uint64 `json:"sequenceIndex"`
	ContentTopic   string `json:"contentTopic"`
	PubsubTopic    string `json:"pubsubTopic"`
	Timestamp      int64  `json:"timestamp"`
	CreatedAt      int64  `json:"createdAt"`
}

// func queryWakuMessagesBetween(db *sql.DB, startsAt time.Time, endsAt time.Time) ([]*WakuMessage, error) {
// 	rows, err := db.Query(fmt.Sprintf("SELECT id, sequenceHash, sequenceNumber, contentTopic, pubsubTopic, createdAt FROM wakuMessages WHERE createdAt BETWEEN %d and %d", startsAt.Unix(), endsAt.Unix()))
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	var wakuMessages []*WakuMessage
// 	for rows.Next() {
// 		var wakuMessage WakuMessage
// 		err = rows.Scan(
// 			&wakuMessage.ID,
// 			&wakuMessage.SequenceHash,
// 			&wakuMessage.SequenceNumber,
// 			&wakuMessage.ContentTopic,
// 			&wakuMessage.PubsubTopic,
// 			&wakuMessage.CreatedAt,
// 		)
// 		if err != nil {
// 			return nil, err
// 		}
// 		wakuMessages = append(wakuMessages, &wakuMessage)
// 	}
// 	return wakuMessages, nil
// }

func (r *WakuMessage) put(db *sql.DB) error {
	stmt, err := db.Prepare("INSERT INTO wakuMessages (walletAddress, peerIdSender, peerIdReporter, sequenceHash, sequenceTotal, sequenceIndex, contentTopic, pubsubTopic, timestamp, createdAt) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id;")
	if err != nil {
		return err
	}

	r.CreatedAt = time.Now().Unix()
	lastInsertId := 0
	err = stmt.QueryRow(r.WalletAddress, r.PeerIDSender, r.PeerIDReporter, r.SequenceHash, r.SequenceTotal, r.SequenceIndex, r.ContentTopic, r.PubsubTopic, r.Timestamp, r.CreatedAt).Scan(&lastInsertId)
	if err != nil {
		return err
	}
	r.ID = lastInsertId

	return nil
}
