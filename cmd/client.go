package main

import (
	"coupon-service/internal/config"
	"coupon-service/internal/infrastructure/entity"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/redis/go-redis/v9"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"gorm.io/gorm"

	"coupon-service/api/grpc/service"
	"coupon-service/internal/application"
	"coupon-service/internal/infrastructure/repository"
	"github.com/Sujin1135/coupon-service-interface/protobuf/service/serviceconnect"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := autoMigrate(config.DBClient); err != nil {
		log.Fatalf("failed to migrate this project's database: %v", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	couponRepo := repository.NewCouponRepository(config.DBClient)
	issuedCouponRepo := repository.NewIssuedCouponRepository(config.DBClient)

	couponService := application.NewCouponService(
		redisClient,
		couponRepo,
		issuedCouponRepo,
	)

	grpcService := service.NewGreetServiceHandler(couponService)

	prefix, connectHandler := serviceconnect.NewGreetServiceHandler(grpcService)

	mux := http.NewServeMux()

	mux.Handle(prefix, connectHandler)

	wrappedHandler := addMiddleware(mux)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Starting ConnectRPC server on %s", addr)
	log.Printf("Service available at: %s", prefix)

	if err := http.ListenAndServe(
		addr,
		h2c.NewHandler(wrappedHandler, &http2.Server{}),
	); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func autoMigrate(db *gorm.DB) error {
	couponTableExists := db.Migrator().HasTable(&entity.CouponEntity{})
	issuedCouponTableExists := db.Migrator().HasTable(&entity.IssuedCouponEntity{})

	if couponTableExists && issuedCouponTableExists {
		log.Println("데이터베이스 테이블이 이미 존재합니다. 마이그레이션을 건너뜁니다.")
		return nil
	}

	log.Println("데이터베이스 마이그레이션을 실행합니다...")

	if err := db.AutoMigrate(&entity.CouponEntity{}, &entity.IssuedCouponEntity{}); err != nil {
		return fmt.Errorf("자동 마이그레이션 실패: %w", err)
	}

	log.Println("데이터베이스 마이그레이션이 성공적으로 완료되었습니다.")
	return nil
}

func addMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, Connect-Protocol-Version")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		handler.ServeHTTP(w, r)
	})
}
