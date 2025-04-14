package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"store/internal/models"
	"store/internal/repository"
	"store/pkg/consul"
	"store/pkg/email"
	"time"

	"github.com/hashicorp/consul/api"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderService interface {
	GetGroupedOrders(ctx context.Context, req models.SearchOrderRequest) ([]*models.GroupedOrder, int64, error)
	GetOrderDetail(ctx context.Context, id string) (*models.GroupedOrder, error)
	GetOrdersUser(ctx context.Context, userID string) ([]*models.GroupedOrder, error)
	CreateOrder(ctx context.Context, req *models.CreateOrderRequest) (*models.Order, error)
	DeleteOrder(ctx context.Context, id string) error
	VerifyPayment(ctx context.Context, orderID string) error
	CancelUnpaidOrder(ctx context.Context, orderID string) error
	CancelUnpaidOrders(ctx context.Context) error
}

type orderService struct {
	orderRepo    repository.OrderRepository
	cartAPI      *callAPI
	emailService *email.EmailService
	bankAccount  models.BankAccount
}

type callAPI struct {
	client       consul.ServiceDiscovery
	clientServer *api.CatalogService
}

var (
	getCartUser = "cart-service"
)

func NewOrderService(orderRepo repository.OrderRepository, client *api.Client) OrderService {
	cartAPI := NewServiceAPI(client, getCartUser)
	emailService := email.NewEmailService()
	bankAccount := models.BankAccount{
		AccountName:   "VOANHHAO",
		AccountNumber: "1020601040",
		BankName:      "Ngân hàng VietComBank",
	}
	return &orderService{
		orderRepo:    orderRepo,
		cartAPI:      cartAPI,
		emailService: emailService,
		bankAccount:  bankAccount,
	}
}

func (s *orderService) GetGroupedOrders(ctx context.Context, req models.SearchOrderRequest) ([]*models.GroupedOrder, int64, error) {
	return s.orderRepo.GetOrders(ctx, req)
}

func (s *orderService) GetOrderDetail(ctx context.Context, id string) (*models.GroupedOrder, error) {

	objectID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return nil, fmt.Errorf("invalid product ID")
	}

	return s.orderRepo.GetOrderDetail(ctx, objectID)
}

func (s *orderService) GetOrdersUser(ctx context.Context, TeacherID string) ([]*models.GroupedOrder, error) {
	return s.orderRepo.GetOrdersUser(ctx, TeacherID)
}

func (s *orderService) CreateOrder(ctx context.Context, req *models.CreateOrderRequest) (*models.Order, error) {
	
	cartData := s.cartAPI.GetCartByUserID(req.TeacherID)
	if cartData == nil {
		return nil, fmt.Errorf("failed to get cart data")
	}

	checkCartStudent := false
	for _, studentCart := range cartData.([]models.StudentCart) {
		if len(studentCart.Items) == 0 {
			continue
		} else if len(studentCart.Items) > 0 {
			checkCartStudent = true
			break
		}
	}

	if !checkCartStudent {
		return nil, fmt.Errorf("cart is empty")
	}

	fmt.Printf("Cart data retrieved: %v\n", cartData.([]models.StudentCart))

	var orderItems []models.OrderItem
	var total float64

	for _, studentCart := range cartData.([]models.StudentCart) {
		for _, cartItem := range studentCart.Items {
			productID, err := primitive.ObjectIDFromHex(cartItem.ProductID)
			if err != nil {
				fmt.Printf("Invalid product ID for student %s: %v\n", studentCart.StudentID, err)
				continue
			}

			totalPrice := float64(cartItem.Quantity) * cartItem.Price
			total += totalPrice

			orderItems = append(orderItems, models.OrderItem{
				ProductID:  productID,
				Quantity:   cartItem.Quantity,
				Price:      cartItem.Price,
				Name:       cartItem.ProductName,
				TotalPrice: totalPrice,
				StudentID:  studentCart.StudentID,
			})
		}
	}

	status := models.OrderStatusPending
	orderNumber := generateOrderNumber()

	payment := models.Payment{
		Method: req.Types,
		Paid:   false,
	}

	if req.Types == models.PaymentMethodBankTransfer {
		content := fmt.Sprintf("%s - %s", orderNumber, "BANK_TRANSFER")
		payment.TransferContent = &content
	}

	order := models.Order{
		TeacherID:   req.TeacherID,
		OrderNumber: orderNumber,
		Email:       req.Email,
		TotalPrice:  total,
		Status:      status,
		Items:       orderItems,
		ShippingAddress: models.Address{
			Street:     req.Street,
			City:       req.City,
			State:      req.State,
			Country:    req.Country,
			Phone:      req.Phone,
		},
		Payment:   payment,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	savedOrder, err := s.orderRepo.CreateOrder(ctx, order)
	if err != nil {
		return nil, fmt.Errorf("unable to create order: %w", err)
	}

	orders := []*models.Order{savedOrder}
	groupedOrder, err := s.orderRepo.GetGroupedOrders(ctx, orders)
	if err != nil {
		return nil, fmt.Errorf("unable to create order: %w", err)
	}

	// Send appropriate emails based on payment method
	if req.Email != "" {
		if req.Types == models.PaymentMethodCOD {
			if err := s.emailService.SendOrderConfirmation(req.Email, groupedOrder[0]); err != nil {
				fmt.Printf("failed to send confirmation email: %v\n", err)
			} else {
				fmt.Printf("order confirmation email sent to: %s\n", req.Email)
			}

			if err := s.emailService.SendEmailNotificationAdmin(os.Getenv("SMTP_USER")); err != nil {
				fmt.Printf("failed to send notification email: %v\n", err)
			} else {
				log.Printf("order notification email sent to: %s\n", os.Getenv("SMTP_USER"))
			}
		} else if req.Types == models.PaymentMethodBankTransfer {
			if err := s.emailService.SendOrderConfirmationBank(req.Email, groupedOrder[0], s.bankAccount); err != nil {
				fmt.Printf("Failed to send confirmation email: %v\n", err)
			} else {
				fmt.Printf("order confirmation email sent to: %s\n", req.Email)
			}

			if err := s.emailService.SendEmailNotificationAdmin(os.Getenv("SMTP_USER")); err != nil {
				fmt.Printf("failed to send notification email: %v\n", err)
			} else {
				log.Printf("order notification email sent to: %s\n", os.Getenv("SMTP_USER"))
			}
		}
	}

	return savedOrder, nil
}

func (s *orderService) DeleteOrder(ctx context.Context, id string) error {

	objectID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return fmt.Errorf("invalid product ID")
	}

	return s.orderRepo.DeleteOrder(ctx, objectID)

}

func NewServiceAPI(client *api.Client, serviceName string) *callAPI {
	sd, err := consul.NewServiceDiscovery(client, serviceName)
	if err != nil {
		fmt.Printf("Error creating service discovery: %v\n", err)
		return nil
	}

	service, err := sd.DiscoverService()
	if err != nil {
		fmt.Printf("Error discovering service: %v\n", err)
		return nil
	}

	return &callAPI{
		client:       sd,
		clientServer: service,
	}
}

func (c *callAPI) GetCartByUserID(UserID string) interface{} {
	endpoint := fmt.Sprintf("/api/v1/cart/items/%s", UserID)
	fmt.Printf("Calling cart service with endpoint: %s\n", endpoint)

	res, err := c.client.CallAPI(c.clientServer, endpoint, http.MethodGet, nil, nil)
	if err != nil {
		fmt.Printf("Error calling cart service: %v\n", err)
		return nil
	}

	fmt.Printf("Response from cart service: %s\n", res)
	var responseWrapper struct {
		StatusCode int                  `json:"status_code"`
		Message    string               `json:"message"`
		Data       []models.StudentCart `json:"data"`
	}

	err = json.Unmarshal([]byte(res), &responseWrapper)
	if err != nil {
		fmt.Printf("Error unmarshaling cart data: %v\n", err)
		return nil
	}
	// Return the actual cart data from inside the wrapper
	return responseWrapper.Data
}

func generateOrderNumber() string {
	timestamp := time.Now().Format("20060102150405") // YYYYMMDDhhmmss
	randomNum := 100 + rand.Intn(900)
	return fmt.Sprintf("DH%s%d", timestamp, randomNum)
}

func (r *orderService) GetBankAccountInfor() models.BankAccount {
	return r.bankAccount
}

func (s *orderService) VerifyPayment(ctx context.Context, orderID string) error {
	objectID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return fmt.Errorf("invalid order ID")
	}

	order, err := s.orderRepo.GetOrderDetail(ctx, objectID)
	if err != nil {
		return fmt.Errorf("order not found")
	}

	if order.Payment.Paid {
		return fmt.Errorf("payment has already been verified")
	}

	// Update payment status
	now := time.Now()
	payment := models.Payment{
		Method:          order.Payment.Method,
		Paid:            true,
		PaidAt:          &now,
		TransferContent: order.Payment.TransferContent,
	}

	// Update order status
	statusUpdate := models.OrderStatusProcessing

	// Update both payment and status in the database
	if err := s.orderRepo.UpdateOrderPaymentAndStatus(ctx, objectID, payment, statusUpdate); err != nil {
		return fmt.Errorf("failed to update order payment and status: %w", err)
	}

	// Send confirmation email
	if order.Email != "" {
		if err := s.emailService.SendOrderConfirmationUpdate(order.Email, order); err != nil {
			fmt.Printf("failed to send payment confirmation email: %v\n", err)
		}
	}

	return nil
}

func (s *orderService) CancelUnpaidOrder(ctx context.Context, orderID string) error {
	objectID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return fmt.Errorf("invalid order ID")
	}

	order, err := s.orderRepo.GetOrderDetail(ctx, objectID)
	if err != nil {
		return fmt.Errorf("order not found")
	}

	if order.Payment.Paid {
		return fmt.Errorf("order has already been paid")
	}

	// Update order status to cancelled
	statusUpdate := models.OrderStatusCanceled

	if err := s.orderRepo.UpdateOrder(ctx, objectID, statusUpdate); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	// Send cancellation email
	if order.Email != "" {
		if err := s.emailService.SendOrderCancellation(order.Email, order); err != nil {
			fmt.Printf("failed to send cancellation email: %v\n", err)
		}
	}

	return nil
}

func (s *orderService) CancelUnpaidOrders(ctx context.Context) error {
	threshold := time.Now().Add(-1 * 24 * time.Hour)

	unpaidOrders, err := s.orderRepo.FindUnPaidOrdersBeforeTime(ctx, threshold, models.OrderStatusPending)
	if err != nil {
		log.Printf("Error finding unpaid orders: %v", err)
		return err
	}

	log.Printf("Found %d unpaid orders to cancel", len(unpaidOrders))

	for _, order := range unpaidOrders {
		if err := s.CancelUnpaidOrder(ctx, order.ID.Hex()); err != nil {
			log.Printf("Failed to cancel order %s: %v", order.ID.Hex(), err)
			continue
		}
		log.Printf("Successfully cancelled unpaid order: %s", order.ID.Hex())
	}
	return nil 
}
