package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/thxnxt11/payment_test/config"
	"github.com/thxnxt11/payment_test/controllers"
	"github.com/thxnxt11/payment_test/routes"
	"github.com/thxnxt11/payment_test/services"
)

func main() {
	cfg := config.Load()

	ctx := context.Background()
	minioClient, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: cfg.MinioUseSSL,
	})
	if err != nil {
		log.Fatalf("minio client init failed: %v", err)
	}

	if err := services.EnsureBucket(ctx, minioClient, cfg.MinioBucket); err != nil {
		log.Fatalf("minio bucket init failed: %v", err)
	}

	qrService := services.NewQRService(minioClient, cfg.MinioBucket, time.Duration(cfg.MinioURLExpiryMins)*time.Minute)
	qrController := controllers.NewQRController(qrService)

	handler := routes.NewRouter(cfg.CORSOrigin, qrController)
	server := &http.Server{
		Addr:              cfg.ServerAddr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("backend listening on %s", cfg.ServerAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}