package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/consul/api"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"store/internal/models"
	"store/internal/repository"
	"store/pkg/consul"
	"store/pkg/email"
	"time"
)

type OrderService interface {
	GetGroupedOrders(ctx context.Context, orders []*models.Order) ([]*models.GroupedOrder, error)
	GetOrderDetail(ctx context.Context, id string) (*models.GroupedOrder, error)
	GetOrdersUser(ctx context.Context, userID string) ([]*models.GroupedOrder, error)
	CreateOrder(ctx context.Context, req *models.TeacherIdRequest) (*models.Order, error)
	UpdateOrder(ctx context.Context, id string, orderRequest models.UpdateStatusRequest) error
	DeleteOrder(ctx context.Context, id string) error
}

type orderService struct {
	orderRepo    repository.OrderRepository
	cartAPI      *callAPI
	emailService *email.EmailService
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
	return &orderService{
		orderRepo:    orderRepo,
		cartAPI:      cartAPI,
		emailService: emailService,
	}
}

func (s *orderService) GetGroupedOrders(ctx context.Context, orders []*models.Order) ([]*models.GroupedOrder, error) {
	return s.orderRepo.GetOrders(ctx)
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

func (s *orderService) CreateOrder(ctx context.Context, req *models.TeacherIdRequest) (*models.Order, error) {
	fmt.Printf("Creating order for user: %s\n", req.TeacherID)

	cartData := s.cartAPI.GetCartByUserID(req.TeacherID)
	if cartData == nil {
		return nil, fmt.Errorf("failed to get cart data")
	}

	checkCartStudent := false
	for _, studentCart := range cartData.([]models.StudentCart) {
		if len(studentCart.Items) == 0 {
			continue
		} else {
			checkCartStudent = true
			break
		}
	}

	if !checkCartStudent {
		return nil, fmt.Errorf("cart is empty")
	}

	fmt.Printf("Cart data retrieved: %v\n", cartData.([]models.StudentCart))

	var orderItems []models.OrderItems
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

			orderItems = append(orderItems, models.OrderItems{
				ProductID:  productID,
				Quantity:   cartItem.Quantity,
				Price:      cartItem.Price,
				Name:       cartItem.ProductName,
				TotalPrice: totalPrice,
				StudentID:  studentCart.StudentID,
			})
		}
	}

	order := models.Order{
		TeacherID:  req.TeacherID,
		TotalPrice: total,
		Email:      req.Email,
		Status:     "awaiting_confirmation",
		CreateAt:   time.Now(),
		UpdateAt:   time.Now(),
		Items:      orderItems,
		ShippingAddress: models.Address{
			Type:       "home",
			Street:     "123 Đường Láng",
			City:       "Hà Nội",
			State:      "Hà Nội",
			Country:    "Việt Nam",
			PostalCode: "100000",
			Phone:      "0912345678",
			IsDefault:  true,
		},
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

	if req.Email == "" {
		fmt.Println("Warning: No email address provided for order confirmation")
	} else {
		if err := s.emailService.SendOrderConfirmation(req.Email, groupedOrder[0]); err != nil {
			fmt.Printf("failed to send confirmation email: %v\n", err)
		} else {
			fmt.Printf("order confirmation email sent to: %s\n", req.Email)
		}
	}

	return savedOrder, nil
}

func (s *orderService) UpdateOrder(ctx context.Context, id string, orderRequest models.UpdateStatusRequest) error {

	objectID, err := primitive.ObjectIDFromHex(id)

	if orderRequest.Status == "confirmed" {
		order, err := s.orderRepo.GetOrderDetail(ctx, objectID)
		if err != nil {
			return fmt.Errorf("order not found")
		}
		if order.Status != "awaiting_confirmation" {
			return fmt.Errorf("order status is not awaiting confirmation")
		}
		if err := s.emailService.SendOrderConfirmationUpdate(order.Email, order); err != nil {
			fmt.Printf("failed to send confirmation email: %v\n", err)
		} else {
			fmt.Printf("order confirmation email sent to: %s\n", "hao01638081724@gmail.com")
		}
	}

	if err != nil {
		return fmt.Errorf("invalid product ID")
	}

	return s.orderRepo.UpdateOrder(ctx, objectID, orderRequest)
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
	endpoint := fmt.Sprintf("/api/cart/items/%s", UserID)
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
