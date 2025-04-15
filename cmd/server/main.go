package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"store/config"
	"store/internal/api"
	"store/internal/repository"
	"store/internal/service"
	"store/pkg/consul"
	"store/pkg/zap"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	cronV3 "github.com/robfig/cron/v3" // Sử dụng phiên bản v3
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func main() {
	// Tải biến môi trường từ tệp .env
	if err := godotenv.Load(); err != nil {
		log.Println("Không tìm thấy tệp .env, sử dụng biến môi trường hệ thống")
	}

	// Khởi tạo cấu hình
	cfg := config.LoadConfig()

	logger, err := zap.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Initialize Consul connection
	consulConn := consul.NewConsulConn(logger, cfg)
	consulClient := consulConn.Connect()
	defer consulConn.Deregister()

	// Thiết lập kết nối MongoDB
	mongoClient, err := connectToMongoDB(cfg.MongoURI)
	if err != nil {
		log.Fatalf("Không thể kết nối đến MongoDB: %v", err)
	}
	defer func() {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			log.Fatal(err)
		}
	}()

	// Khởi tạo repository và service
	orderCollection := mongoClient.Database(cfg.MongoDB).Collection("orders")
	orderRepo := repository.NewOrderRepository(orderCollection)
	orderService := service.NewOrderService(orderRepo, consulClient)
	
	// Thiết lập router với Gin
	router := gin.Default()

	// Khởi tạo cron với cấu hình mới
	c := cronV3.New(cronV3.WithSeconds(), cronV3.WithLogger(cronV3.DefaultLogger))
	
	// Thêm job hủy đơn hàng mỗi giờ
	_, err = c.AddFunc("0 0 * * * *", func() { // Chạy mỗi giờ tại phút thứ 0, giây thứ 0
		
		log.Println("Bắt đầu chạy job...")
		ctx := context.Background()

		if err := orderService.CancelUnpaidOrders(ctx); err != nil {
			log.Printf("Lỗi khi hủy đơn hàng chưa thanh toán: %v", err)
		} else {
			log.Println("Job hủy đơn hàng chưa thanh toán hoàn thành")
		}

		if err := orderService.SentPaymentReminders(ctx); err != nil {
			log.Printf("Lỗi khi gửi nhắc nhở thanh toán: %v", err)
		} else {
			log.Println("Job nhắc nhở thanh toán hoàn thành")
		}
	})

	
	if err != nil {
		log.Printf("Lỗi khi thiết lập cron job: %v", err)
	} else {
		c.Start()
		log.Println("Cron job đã được khởi động")
	}
	
	// Đăng ký handlers
	api.RegisterHandlers(router, orderService)

	// Khởi tạo server HTTP
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Chạy server trong goroutine riêng biệt
	go func() {
		log.Printf("Server đang chạy trên cổng %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Lỗi khi khởi chạy server: %v", err)
		}
	}()

	// Thiết lập xử lý tắt server một cách nhẹ nhàng
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Đang tắt server...")

	// Dừng cron job
	log.Println("Đang dừng cron job...")
	cronCtx := c.Stop()
	<-cronCtx.Done()
	log.Println("Cron job đã dừng")

	// Tắt server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Lỗi khi tắt server: %v", err)
	}
	log.Println("Server đã tắt")
}

func connectToMongoDB(uri string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	
	// Kiểm tra kết nối
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}
	
	log.Println("Kết nối đến MongoDB thành công")
	return client, nil
}