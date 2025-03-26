package main

import (
	"context"
	"coupon-service/internal/application"
	"coupon-service/internal/infrastructure/cache"
	"fmt"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cacheClient := cache.NewCacheClient[int]()
	err := cacheClient.Set(ctx, "coupon:coupon-1:remaining", 500)
	if err != nil {
		fmt.Println("failed to set coupon:coupon-1:remaining")
	}

	couponService := application.NewCouponService()
	result, err := couponService.IssueCoupon(ctx, "coupon-1", "user-1")
	if err != nil {
		fmt.Println("occurred an error when issuing a coupon")
	}
	if result == false {
		fmt.Println("failed to issue a coupon")
	} else {
		fmt.Println("coupon issued successfully")
	}
}
