package domain

import (
	"github.com/google/uuid"
	"math/rand"
	"time"
)

type IssuedCoupon struct {
	ID         string    `json:"id"`
	CouponID   string    `json:"coupon_id"`
	Code       string    `json:"code"`
	CreatedAt  time.Time `json:"created_at"`
	ModifiedAt time.Time `json:"modified_at"`
}

func NewIssuedCoupon(couponId string, createdAt time.Time) *IssuedCoupon {
	id := uuid.New()

	return &IssuedCoupon{
		ID:         id.String(),
		CouponID:   couponId,
		Code:       generateUniqueCode(),
		CreatedAt:  createdAt,
		ModifiedAt: createdAt,
	}
}

func generateUniqueCode() string {
	// 랜덤 시드 초기화
	rand.Seed(time.Now().UnixNano())

	// 코드 길이를 1-10 사이에서 랜덤하게 결정
	length := rand.Intn(10) + 1

	// 결과 문자열을 위한 rune 슬라이스
	result := make([]rune, length)

	// 한글 유니코드 범위 (가-힣)
	// 초성 범위: AC00-D7A3
	const (
		hangulStart = 0xAC00
		hangulEnd   = 0xD7A3
		hangulCount = hangulEnd - hangulStart + 1
	)

	// 각 위치에 한글 또는 숫자 랜덤 배치
	for i := 0; i < length; i++ {
		// 50% 확률로 한글 또는 숫자 선택
		if rand.Intn(2) == 0 {
			// 한글 문자 (가-힣)
			result[i] = rune(hangulStart + rand.Intn(hangulCount))
		} else {
			// 숫자 (0-9)
			result[i] = rune('0' + rand.Intn(10))
		}
	}

	return string(result)
}
