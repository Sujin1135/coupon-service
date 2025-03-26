package application

import (
	"context"
	"coupon-service/internal/domain"
	"coupon-service/internal/infrastructure/cache"
	"coupon-service/internal/infrastructure/repository"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"time"
)

type CouponService struct {
	cache                  cache.Cache
	issuedCouponRepository *repository.IssuedCouponRepository
}

func NewCouponService(
	cacheClient *redis.Client,
	issuedCouponRepository *repository.IssuedCouponRepository,
) *CouponService {
	return &CouponService{
		cache:                  cache.NewCacheClient(cacheClient),
		issuedCouponRepository: issuedCouponRepository,
	}
}

const (
	ValidateJsonUnmarshalError = IssueCouponError("ValidateJsonUnmarshalError")
	CouponNotStartedError      = IssueCouponError("coupon issuance has not started yet")
	CouponExpiredError         = IssueCouponError("the coupon issuance period has expired")
	DuplicatedCouponUserError  = IssueCouponError("coupon already issued to this user")
	CouponAmountRecoveryError  = IssueCouponError("failed to increment coupon amount for recover")
	DeleteRecoveryError        = IssueCouponError("failed to delete coupon for recover")
	AllCouponIssuedError       = IssueCouponError("all coupons has been issued")
	CacheAddUserError          = IssueCouponError("failed to add coupon")
	CouponDecrError            = IssueCouponError("failed to decrement coupon")
	DataKeyNotFoundError       = IssueCouponError("data key not found")
	IssuedCouponCreationError  = IssueCouponError("failed to create coupon code")
)

type IssueCouponError string

func (e IssueCouponError) Error() string { return string(e) }

func (c *CouponService) IssueCoupon(
	ctx context.Context,
	couponId string,
	userId string,
) error {
	dataKey := "coupon:" + couponId + ":data"
	userStoreKey := "coupon:" + couponId + ":users"
	couponKey := "coupon:" + couponId + ":remaining"
	now := time.Now()

	err := c.validateCouponEvent(ctx, dataKey, now)
	if err != nil {
		return err
	}

	err2 := c.controlConcurrent(ctx, userStoreKey, userId, couponKey)
	if err2 != nil {
		return err2
	}

	err3 := c.issuedCouponRepository.Save(domain.NewIssuedCoupon(couponId, now))
	if err3 != nil {
		return IssuedCouponCreationError
	}

	return nil
}

func (c *CouponService) validateCouponEvent(ctx context.Context, dataKey string, now time.Time) error {
	var coupon domain.Coupon
	data, err := c.cache.Get(ctx, dataKey)
	if err != nil {
		return DataKeyNotFoundError
	}
	if err2 := json.Unmarshal(data, &coupon); err2 != nil {
		return ValidateJsonUnmarshalError
	}
	if coupon.IssuedAt.After(now) {
		return CouponNotStartedError
	}
	if coupon.ExpiresAt.Before(now) {
		return CouponExpiredError
	}
	return nil
}

func (c *CouponService) controlConcurrent(
	ctx context.Context,
	userStoreKey string,
	userId string,
	couponKey string,
) error {
	added, err := c.cache.SetAdd(ctx, userStoreKey, userId)
	if err != nil {
		fmt.Println(err.Error())
		return CacheAddUserError
	}
	if added == false {
		return DuplicatedCouponUserError
	}

	count, err := c.cache.Decr(ctx, couponKey)
	if err != nil {
		fmt.Println(err.Error())
		return CouponDecrError
	}

	if count < 0 {
		_, incrErr := c.cache.Incr(ctx, couponKey)
		if incrErr != nil {
			log.Println(incrErr)
			return CouponAmountRecoveryError
		}

		_, delErr := c.cache.SetDel(ctx, userStoreKey, userId)
		if delErr != nil {
			log.Println(delErr)
			return DeleteRecoveryError
		}
		return AllCouponIssuedError
	}
	return nil
}
