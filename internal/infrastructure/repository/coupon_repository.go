package repository

import (
	"coupon-service/internal/domain"
	"coupon-service/internal/infrastructure/entity"
	"errors"
	"fmt"
	"gorm.io/gorm"
)

type CouponRepository struct {
	db *gorm.DB
}

func NewCouponRepository(db *gorm.DB) *CouponRepository {
	return &CouponRepository{
		db: db,
	}
}

func (r *CouponRepository) FindOne(id string) (*domain.Coupon, error) {
	var couponEntity entity.CouponEntity
	err := r.db.Where(
		"id = ? AND deleted_at IS NULL", id,
	).First(&couponEntity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("coupon not found")
	}
	if err != nil {
		fmt.Println(err)
		return nil, errors.New(fmt.Sprintf("occurred an error when find a coupon by id(%s)", id))
	}

	return &domain.Coupon{
		ID:          couponEntity.ID,
		Name:        couponEntity.Name,
		IssueAmount: couponEntity.IssueAmount,
		IssuedAt:    couponEntity.IssuedAt,
		ExpiresAt:   couponEntity.ExpiresAt,
		CreatedAt:   couponEntity.CreatedAt,
		ModifiedAt:  couponEntity.ModifiedAt,
	}, nil
}
