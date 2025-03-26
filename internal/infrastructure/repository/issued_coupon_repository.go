package repository

import (
	"coupon-service/internal/domain"
	"coupon-service/internal/infrastructure/entity"
	"fmt"
	"gorm.io/gorm"
)

type IssuedCouponRepository struct {
	db *gorm.DB
}

func NewIssuedCouponRepository(db *gorm.DB) *IssuedCouponRepository {
	return &IssuedCouponRepository{
		db: db,
	}
}

func (r *IssuedCouponRepository) Save(domain *domain.IssuedCoupon) error {
	return r.db.Save(&entity.IssuedCouponEntity{
		ID:         domain.ID,
		CouponID:   domain.CouponID,
		Code:       domain.Code,
		CreatedAt:  domain.CreatedAt,
		ModifiedAt: domain.ModifiedAt,
		DeletedAt:  nil,
	}).Error
}

func (r *IssuedCouponRepository) FindByCouponId(couponId string) []domain.IssuedCoupon {
	var issuedCouponEntities []entity.IssuedCouponEntity
	err := r.db.Where(
		"coupon_id = ? AND deleted_at IS NULL", couponId,
	).Find(&issuedCouponEntities).Error
	if err != nil {
		fmt.Println(err)
	}

	domains := make([]domain.IssuedCoupon, len(issuedCouponEntities))
	for i, v := range issuedCouponEntities {
		domains[i] = domain.IssuedCoupon{
			ID:         v.ID,
			CouponID:   v.CouponID,
			Code:       v.Code,
			CreatedAt:  v.CreatedAt,
			ModifiedAt: v.ModifiedAt,
		}
	}

	return domains
}
