package domain

import (
	"github.com/google/uuid"
	"time"
)

type Coupon struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	IssueAmount   int64          `json:"issue_amount"`
	IssuedAt      time.Time      `json:"issued_at"`
	ExpiresAt     time.Time      `json:"expires_at"`
	IssuedCoupons []IssuedCoupon `json:"issued_coupons"`
	CreatedAt     time.Time      `json:"created_at"`
	ModifiedAt    time.Time      `json:"modified_at"`
}

func NewCoupon(
	name string,
	issueAmount int64,
	issuedAt time.Time,
	expiresAt time.Time,
) *Coupon {
	now := time.Now()
	return &Coupon{
		ID:          uuid.New().String(),
		Name:        name,
		IssueAmount: issueAmount,
		IssuedAt:    issuedAt,
		ExpiresAt:   expiresAt,
		CreatedAt:   now,
		ModifiedAt:  now,
	}
}
