package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"store/internal/models"
	"store/internal/repository"
	"store/pkg/constants"
	"store/pkg/consul"
	"store/pkg/email"
	"sync"
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
	// CancelUnpaidOrders(ctx context.Context) error
	// SentPaymentReminders(ctx context.Context) error
}

type orderService struct {
	orderRepo    repository.OrderRepository
	cartAPI      *callAPI
	walletAPI    *callAPI
	emailService *email.EmailService
	bankAccount  models.BankAccount
}

type callAPI struct {
	client       consul.ServiceDiscovery
	clientServer *api.CatalogService
}

var (
	getCartUser   = "cart-service"
	getWalletUser = "wallet-service"
	orderMutex    sync.Mutex
	lastOrderDate string
	counter       uint32
)

func NewOrderService(orderRepo repository.OrderRepository, client *api.Client) OrderService {
	cartAPI := NewServiceAPI(client, getCartUser)
	walletAPI := NewServiceAPI(client, getWalletUser)
	emailService := email.NewEmailService()
	bankAccount := models.BankAccount{
		AccountName:   "VOANHHAO",
		AccountNumber: "1020601040",
		BankName:      "Ngân hàng VietComBank",
	}
	return &orderService{
		orderRepo:    orderRepo,
		cartAPI:      cartAPI,
		walletAPI:    walletAPI,
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

	cartData := s.cartAPI.GetCartByUserID(ctx)

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

	var orderItems []models.OrderItem
	var totalStore float64
	var totalService float64

	for _, studentCart := range cartData.([]models.StudentCart) {
		for _, cartItem := range studentCart.Items {
			productID, err := primitive.ObjectIDFromHex(cartItem.ProductID)
			if err != nil {
				fmt.Printf("Invalid product ID for student %s: %v\n", studentCart.StudentID, err)
				continue
			}

			totalPriceStore := float64(cartItem.Quantity) * cartItem.PriceStore
			totalPriceService := float64(cartItem.Quantity) * cartItem.PriceService

			totalStore += totalPriceStore
			totalService += totalPriceService

			orderItems = append(orderItems, models.OrderItem{
				ProductID:         productID,
				Quantity:          cartItem.Quantity,
				PriceStore:        cartItem.PriceStore,
				PriceService:      cartItem.PriceService,
				Name:              cartItem.ProductName,
				TotalPriceStore:   totalPriceStore,
				TotalPriceService: totalPriceService,
				StudentID:         studentCart.StudentID,
			})
		}
	}

	status := models.OrderStatusProcessing
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
		TeacherID:         req.TeacherID,
		OrderNumber:       orderNumber,
		Email:             req.Email,
		TotalPriceStore:   totalStore,
		TotalPriceService: totalService,
		Status:            status,
		Items:             orderItems,
		ShippingAddress: models.Address{
			Street:  req.Street,
			City:    req.City,
			State:   req.State,
			Country: req.Country,
			Phone:   req.Phone,
		},
		Payment:        payment,
		ReminderSent:   false,
		ReminderSentAt: nil,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	response, err := s.walletAPI.DeductBalance(ctx, order.TotalPriceStore, order.TotalPriceService)
	if err != nil {
		return nil, fmt.Errorf("unable to create order: %w", err)
	}

	if respMap, ok := response.(map[string]interface{}); ok {
		if statusCode, exists := respMap["status_code"].(float64); exists && statusCode >= 400 {
			errorMsg := respMap["error"]
			errorCode := respMap["error_code"]
			return nil, fmt.Errorf("API Error: %v, Code: %v, Status: %v", errorMsg, errorCode, statusCode)
		}
	}

	savedOrder, err := s.orderRepo.CreateOrder(ctx, order)
	if err != nil {
		return nil, fmt.Errorf("unable to create order: %w", err)
	}

	// orders := []*models.Order{savedOrder}
	// groupedOrder, err := s.orderRepo.GetGroupedOrders(ctx, orders)
	// if err != nil {
	// 	return nil, fmt.Errorf("unable to create order: %w", err)
	// }

	// Send appropriate emails based on payment method
	// if req.Email != "" {
	// 	switch req.Types {
	// 	case models.PaymentMethodCOD:
	// 		if err := s.emailService.SendOrderConfirmation(req.Email, groupedOrder[0]); err != nil {
	// 			fmt.Printf("failed to send confirmation email: %v\n", err)
	// 		} else {
	// 			fmt.Printf("order confirmation email sent to: %s\n", req.Email)
	// 		}

	// 	case models.PaymentMethodBankTransfer:
	// 		if err := s.emailService.SendOrderConfirmationBank(req.Email, groupedOrder[0], s.bankAccount); err != nil {
	// 			fmt.Printf("Failed to send confirmation email: %v\n", err)
	// 		} else {
	// 			fmt.Printf("order confirmation email sent to: %s\n", req.Email)
	// 		}
	// 	}

	// 	if err := s.emailService.SendEmailNotificationAdmin(os.Getenv("SMTP_USER"), groupedOrder[0]); err != nil {
	// 		fmt.Printf("failed to send notification email: %v\n", err)
	// 	} else {
	// 		log.Printf("order notification email sent to: %s\n", os.Getenv("SMTP_USER"))
	// 	}
	// }

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

func (c *callAPI) GetCartByUserID(ctx context.Context) interface{} {

	endpoint := "/api/v1/cart/items"

	header := map[string]string{
		"Authorization": "Bearer " + ctx.Value(constants.TokenKey).(string),
	}

	res, err := c.client.CallAPI(c.clientServer, endpoint, http.MethodGet, nil, header)
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

	return responseWrapper.Data
}

func (c *callAPI) DeductBalance(ctx context.Context, priceSotre, priceService float64) (interface{}, error) {

	requestBody := map[string]float64{
		"price_store":   priceSotre,
		"price_service": priceService,
	}

	fmt.Printf("Request body: %v\n", requestBody)

	// Chuyển đổi dữ liệu thành JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling JSON: %v", err)
	}

	// Thiết lập headers
	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + ctx.Value(constants.TokenKey).(string),
	}

	// Gọi API sử dụng phương thức POST
	endpoint := "/api/v1/wallet/deduct_balance"
	res, err := c.client.CallAPI(c.clientServer, endpoint, http.MethodPost, jsonData, headers)
	if err != nil {
		return nil, fmt.Errorf("error calling API: %v", err)
	}

	// In ra response để debug
	fmt.Printf("Raw API response: %s\n", res)

	// Xử lý kết quả trả về
	var responseData interface{}
	err = json.Unmarshal([]byte(res), &responseData)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling response: %v", err)
	}

	return responseData, nil
}

func generateOrderNumber() string {

	orderMutex.Lock()
	defer orderMutex.Unlock()

	now := time.Now()
	today := now.Format("060102")

	if today != lastOrderDate {
		lastOrderDate = today
		counter = 0
	}

	counter++

	return fmt.Sprintf("DH%s%04d", today, counter)
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

	now := time.Now()
	payment := models.Payment{
		Method:          order.Payment.Method,
		Paid:            true,
		PaidAt:          &now,
		TransferContent: order.Payment.TransferContent,
	}

	statusUpdate := models.OrderStatusProcessing

	if err := s.orderRepo.UpdateOrderPaymentAndStatus(ctx, objectID, payment, statusUpdate); err != nil {
		return fmt.Errorf("failed to update order payment and status: %w", err)
	}

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

	statusUpdate := models.OrderStatusCanceled

	if err := s.orderRepo.UpdateOrder(ctx, objectID, statusUpdate); err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	if order.Email != "" {
		if err := s.emailService.SendOrderCancellation(order.Email, order); err != nil {
			fmt.Printf("failed to send cancellation email: %v\n", err)
		}
	}

	return nil
}

// func (s *orderService) CancelUnpaidOrders(ctx context.Context) error {
// 	threshold := time.Now().Add(-1 * 24 * time.Hour)

// 	unpaidOrders, err := s.orderRepo.FindUnPaidOrdersBeforeTime(ctx, threshold, models.OrderStatusPending)
// 	if err != nil {
// 		log.Printf("Error finding unpaid orders: %v", err)
// 		return err
// 	}

// 	log.Printf("Found %d unpaid orders to cancel", len(unpaidOrders))

// 	for _, order := range unpaidOrders {
// 		if err := s.CancelUnpaidOrder(ctx, order.ID.Hex()); err != nil {
// 			log.Printf("Failed to cancel order %s: %v", order.ID.Hex(), err)
// 			continue
// 		}
// 		log.Printf("Successfully cancelled unpaid order: %s", order.ID.Hex())
// 	}
// 	return nil
// }

// func (s *orderService) SentPaymentReminder(ctx context.Context, orderID primitive.ObjectID, hoursLeft int) error {

// 	order, err := s.orderRepo.GetOrderDetail(ctx, orderID)
// 	if err != nil {
// 		return fmt.Errorf("order not found")
// 	}

// 	if order.Email != "" {
// 		if err := s.emailService.SendReminderOrder(order.Email, order, hoursLeft, s.bankAccount); err != nil {
// 			fmt.Printf("failed to send cancellation email: %v\n", err)
// 		}
// 	}

// 	return nil

// }

// func (s *orderService) SentPaymentReminders(ctx context.Context) error {

// 	reminderTime := time.Now().Add(-20 * time.Hour)
// 	endTime := time.Now().Add(-22 * time.Hour)

// 	unpaidOrders, err := s.orderRepo.FindOrdersForReminder(ctx, reminderTime, endTime)
// 	if err != nil {
// 		log.Printf("Error finding orders for payment reminder: %v", err)
//         return err
// 	}

// 	log.Printf("Found %d unpaid orders needing payment reminder", len(unpaidOrders))

// 	for _, order := range unpaidOrders {
// 		hoursLeft := int(24 - time.Since(order.CreatedAt).Hours())

// 		if err := s.SentPaymentReminder(ctx, order.ID, hoursLeft); err != nil {
// 			log.Printf("Failed to send reminder for order %s: %v", order.ID.Hex(), err)
//             continue
// 		}

// 		if err := s.orderRepo.MarkReminderSent(ctx, order.ID); err != nil {
// 			log.Printf("Failed to mark reminder sent for order %s: %v", order.ID.Hex(), err)
// 		}

// 		log.Printf("Successfully sent payment reminder for order: %s", order.ID.Hex())
// 	}

// 	return nil
// }
