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

func (r *CouponRepository) Save(domain *domain.Coupon) error {
	return r.db.Save(&entity.CouponEntity{
		ID:          domain.ID,
		Name:        domain.Name,
		IssueAmount: domain.IssueAmount,
		IssuedAt:    domain.IssuedAt,
		ExpiresAt:   domain.ExpiresAt,
		CreatedAt:   domain.CreatedAt,
		ModifiedAt:  domain.ModifiedAt,
		DeletedAt:   nil,
	}).Error
}

func (r *CouponRepository) Delete(id string) error {
	result := r.db.Where("id = ?", id).Delete(id)
	if result.Error != nil {
		return result.Error
	}
	return nil
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
