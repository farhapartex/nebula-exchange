package idempotency

import "time"

type keyStatus string

const (
	statusInProgress keyStatus = "IN_PROGRESS"
	statusCompleted  keyStatus = "COMPLETED"
)

type idempotencyKeyRecord struct {
	Scope               string    `gorm:"primaryKey"`
	IdempotencyKey      string    `gorm:"primaryKey"`
	RequestHash         []byte    `gorm:"not null"`
	Status              keyStatus `gorm:"type:idempotency_key_status;not null"`
	ResponseStatusCode  *int
	ResponseContentType string
	ResponseBody        []byte
	CreatedAt           time.Time `gorm:"<-:create;default:now()"`
	CompletedAt         *time.Time
}

func (idempotencyKeyRecord) TableName() string {
	return "idempotency_keys"
}

func (storedKey idempotencyKeyRecord) toRecord() Record {
	statusCode := 0
	if storedKey.ResponseStatusCode != nil {
		statusCode = *storedKey.ResponseStatusCode
	}
	return Record{
		RequestHash: storedKey.RequestHash,
		IsCompleted: storedKey.Status == statusCompleted,
		Response: StoredResponse{
			StatusCode:  statusCode,
			ContentType: storedKey.ResponseContentType,
			Body:        storedKey.ResponseBody,
		},
		CreatedAt: storedKey.CreatedAt,
	}
}
