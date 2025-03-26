package entity

import "time"

type CouponEntity struct {
	ID          string     `gorm:"primary_key;type:varchar(36);not null"`
	Name        string     `gorm:"type:varchar(20);not null"`
	IssueAmount int64      `gorm:"type:bigint(20);not null"`
	IssuedAt    time.Time  `gorm:"type:timestamp;not null"`
	ExpiresAt   time.Time  `gorm:"type:timestamp;not null"`
	CreatedAt   time.Time  `gorm:"type:timestamp;not null;default:current_timestamp"`
	ModifiedAt  time.Time  `gorm:"type:timestamp;not null;default:current_timestamp ON UPDATE current_timestamp"`
	DeletedAt   *time.Time `gorm:"type:timestamp"`
}

func (CouponEntity) TableName() string {
	return "coupons"
}

type IssuedCouponEntity struct {
	ID         string     `gorm:"primary_key;type:varchar(36)"`
	CouponID   string     `gorm:"type:varchar(36);not null;index:idx_coupon_code,unique"`
	Code       string     `gorm:"type:varchar(10);not null;index:idx_coupon_code,unique"`
	CreatedAt  time.Time  `gorm:"type:timestamp;not null;default:current_timestamp"`
	ModifiedAt time.Time  `gorm:"type:timestamp;not null;default:current_timestamp ON UPDATE current_timestamp"`
	DeletedAt  *time.Time `gorm:"type:timestamp"`
}

func (IssuedCouponEntity) TableName() string {
	return "issued_coupons"
}
