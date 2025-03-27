package main

import (
	"coupon-service/internal/config"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/redis/go-redis/v9"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

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
