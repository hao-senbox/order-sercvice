package api

import (
	"fmt"
	"math"
	"net/http"
	"store/internal/models"
	"store/internal/service"

	"github.com/gin-gonic/gin"
)

type OrderHandlers struct {
	orderService service.OrderService
}

func NewOrderHandlers(orderService service.OrderService) *OrderHandlers {
	return &OrderHandlers{
		orderService: orderService,
	}
}

func RegisterHandlers(router *gin.Engine, orderService service.OrderService) {

	handlers := NewOrderHandlers(orderService)

	adminOrderGroup := router.Group("/api/v1/admin/orders")
	{
		adminOrderGroup.GET("", handlers.GetOrders)
		adminOrderGroup.POST("/:id/verify-payment", handlers.VerifyPayment)
		adminOrderGroup.POST("/:id/cancel", handlers.CancelUnpaidOrder)
	}

	orderGroup := router.Group("/api/v1/orders")
	{
		orderGroup.GET("/:user", handlers.GetOrdersUser)
		orderGroup.POST("/items", handlers.CreateOrder)
		orderGroup.GET("/items/:id", handlers.GetOrderDetail)
		orderGroup.DELETE("/items/:id", handlers.DeleteOrder)
	}
}

func (h *OrderHandlers) CreateOrder(c *gin.Context) {

	var req models.CreateOrderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		SendError(c, http.StatusBadRequest, err, models.ErrInvalidRequest)
		return
	}

	if req.TeacherID == "" {
		SendError(c, http.StatusBadRequest, fmt.Errorf("teacher ID cannot be empty"), models.ErrInvalidRequest)
		return
	}

	if req.Email == "" {
		SendError(c, http.StatusBadRequest, fmt.Errorf("email cannot be empty"), models.ErrInvalidRequest)
		return
	}

	if req.Types == "" {
		SendError(c, http.StatusBadRequest, fmt.Errorf("types cannot be empty"), models.ErrInvalidRequest)
		return
	}

	order, err := h.orderService.CreateOrder(c.Request.Context(), &req)

	if err != nil {
		SendError(c, http.StatusBadRequest, err, models.ErrInvalidOperation)
		return
	}

	SendSuccess(c, http.StatusOK, "Order created successfully", order)

}

func (h *OrderHandlers) GetOrders(c *gin.Context) {

	var req models.SearchOrderRequest

	if err := c.ShouldBindQuery(&req); err != nil {
		SendError(c, http.StatusBadRequest, err, models.ErrInvalidRequest)
		return
	}

	if req.Page == 0 {
		req.Page = 1
	}

	if req.Limit == 0 {
		req.Limit = 10
	}

	results, totalItems, err := h.orderService.GetGroupedOrders(c.Request.Context(), req)

	if err != nil {
		SendError(c, http.StatusBadRequest, err, models.ErrInvalidRequest)
		return
	}

	totalPages := int(math.Ceil(float64(totalItems) / float64(req.Limit)))

	response := models.PaginatedResponse{
		Page:       req.Page,
		Limit:      req.Limit,
		TotalPages: totalPages,
		TotalItems: totalItems,
		Data:       results,
	}

	SendSuccess(c, http.StatusOK, "Order data retrieved successfully", response)

}

func (h *OrderHandlers) GetOrdersUser(c *gin.Context) {
	TeacherID := c.Param("user")

	if TeacherID == "" {
		SendError(c, http.StatusBadRequest, fmt.Errorf("teacher ID cannot be empty"), models.ErrInvalidRequest)
	}

	orderUser, err := h.orderService.GetOrdersUser(c.Request.Context(), TeacherID)

	if err != nil {
		SendError(c, http.StatusBadRequest, err, models.ErrInvalidRequest)
		return
	}

	SendSuccess(c, http.StatusOK, "Order data of teacher retrieved successfully", orderUser)

}

func (h *OrderHandlers) GetOrderDetail(c *gin.Context) {

	if c.Param("id") == "" {
		SendError(c, http.StatusBadRequest, fmt.Errorf("ID order cannot be empty"), models.ErrInvalidRequest)
	}

	order, err := h.orderService.GetOrderDetail(c.Request.Context(), c.Param("id"))

	if err != nil {
		SendError(c, http.StatusInternalServerError, err, models.ErrInvalidOperation)
		return
	}

	SendSuccess(c, http.StatusOK, "Order detail data retrieved successfully", order)

}

func (h *OrderHandlers) DeleteOrder(c *gin.Context) {

	if c.Param("id") == "" {
		SendError(c, http.StatusBadRequest, fmt.Errorf("ID cannot be empty"), models.ErrInvalidRequest)
	}

	err := h.orderService.DeleteOrder(c.Request.Context(), c.Param("id"))

	if err != nil {
		SendError(c, http.StatusInternalServerError, err, models.ErrInvalidOperation)
		return
	}

	SendSuccess(c, http.StatusOK, "Deleted successfully", nil)
}

func (h *OrderHandlers) VerifyPayment(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		SendError(c, http.StatusBadRequest, fmt.Errorf("order ID cannot be empty"), models.ErrInvalidRequest)
		return
	}

	err := h.orderService.VerifyPayment(c.Request.Context(), orderID)
	if err != nil {
		SendError(c, http.StatusBadRequest, err, models.ErrInvalidOperation)
		return
	}

	SendSuccess(c, http.StatusOK, "Payment verified successfully", nil)
}

func (h *OrderHandlers) CancelUnpaidOrder(c *gin.Context) {
	
	orderID := c.Param("id")
	if orderID == "" {
		SendError(c, http.StatusBadRequest, fmt.Errorf("order ID cannot be empty"), models.ErrInvalidRequest)
		return
	}

	err := h.orderService.CancelUnpaidOrder(c.Request.Context(), orderID)
	if err != nil {
		SendError(c, http.StatusBadRequest, err, models.ErrInvalidOperation)
		return
	}

	SendSuccess(c, http.StatusOK, "Order cancelled successfully", nil)
}
