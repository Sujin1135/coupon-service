package domain

import "time"

type Coupon struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	IssueAmount int64     `json:"issue_amount"`
	IssuedAt    time.Time `json:"issued_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
	ModifiedAt  time.Time `json:"modified_at"`
}
