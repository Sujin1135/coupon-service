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
	couponRepository       *repository.CouponRepository
	issuedCouponRepository *repository.IssuedCouponRepository
}

func NewCouponService(
	cacheClient *redis.Client,
	couponRepository *repository.CouponRepository,
	issuedCouponRepository *repository.IssuedCouponRepository,
) *CouponService {
	return &CouponService{
		cache:                  cache.NewCacheClient(cacheClient),
		couponRepository:       couponRepository,
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

const (
	FailedSaveCouponError        = CreateCouponError("failed to save coupon")
	CouponDataRecoveryError      = CreateCouponError("failed to recover coupon")
	CouponCacheDataRecoveryError = CreateCouponError("failed to recover coupon caching data")
	CouponCacheError             = CreateCouponError("failed to cache coupon data")
)

type IssueCouponError string

func (e IssueCouponError) Error() string { return string(e) }

type CreateCouponError string

func (e CreateCouponError) Error() string { return string(e) }

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

func (c *CouponService) CreateCoupon(
	ctx context.Context,
	name string,
	amount int64,
	issuedAt time.Time,
	expiresAt time.Time,
) (*domain.Coupon, error) {
	coupon := domain.NewCoupon(name, amount, issuedAt, expiresAt)
	err := c.couponRepository.Save(coupon)
	if err != nil {
		fmt.Println(err.Error())
		return nil, FailedSaveCouponError
	}
	if err2 := c.cacheCouponData(ctx, coupon); err2 != nil {
		return nil, err2
	}
	if err3 := c.cacheCouponCount(ctx, coupon); err3 != nil {
		return nil, err3
	}

	return coupon, nil
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

func (c *CouponService) cacheCouponData(ctx context.Context, coupon *domain.Coupon) error {
	err := c.cache.Set(ctx, genCouponDataKey(coupon.ID), coupon)
	if err != nil {
		fmt.Println(err.Error())

		if err3 := c.couponRepository.Delete(coupon.ID); err3 != nil {
			log.Println(err3.Error())
			return CouponDataRecoveryError
		}
		return CouponCacheError
	}
	return nil
}

func (c *CouponService) cacheCouponCount(ctx context.Context, coupon *domain.Coupon) error {
	err := c.cache.Set(ctx, genCouponAmountKey(coupon.ID), coupon.IssueAmount)
	if err != nil {
		fmt.Println(err.Error())

		if err2 := c.couponRepository.Delete(coupon.ID); err2 != nil {
			log.Println(err2.Error())
			return CouponDataRecoveryError
		}

		if err3 := c.cache.Del(ctx, genCouponDataKey(coupon.ID)); err3 != nil {
			log.Println(err3.Error())
			return CouponCacheDataRecoveryError
		}

		return CouponCacheError
	}
	return nil
}

func genCouponDataKey(couponID string) string {
	return fmt.Sprintf("coupon:%s:data", couponID)
}

func genCouponAmountKey(couponID string) string {
	return fmt.Sprintf("coupon:%s:remaining", couponID)
}
