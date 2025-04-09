package api

import (
	"fmt"
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

	adminOrderGroup := router.Group("/api/admin/orders")
	{
		adminOrderGroup.GET("", handlers.GetOrders)
	}

	orderGroup := router.Group("/api/orders")
	{
		orderGroup.GET("/:user", handlers.GetOrdersUser)
		orderGroup.POST("/items", handlers.CreateOrder)
		orderGroup.GET("/items/:id", handlers.GetOrderDetail)
		orderGroup.PUT("/items/:id", handlers.UpdateOrder)
		orderGroup.DELETE("/items/:id", handlers.DeleteOrder)
	}
}

func (h *OrderHandlers) CreateOrder(c *gin.Context) {
	
	var req models.TeacherIdRequest

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

	order, err := h.orderService.CreateOrder(c.Request.Context(), &req)

	if err != nil {
		SendError(c, http.StatusBadRequest, err, models.ErrInvalidOperation)
		return 
	}

	SendSuccess(c, http.StatusOK, "Order created successfully", order)

}

func (h *OrderHandlers) GetOrders(c *gin.Context) {

	var orders []*models.Order
	ordersAll, err := h.orderService.GetGroupedOrders(c.Request.Context(), orders)

	if err != nil {
		SendError(c, http.StatusBadRequest, err, models.ErrInvalidRequest)
		return
	}

	SendSuccess(c, http.StatusOK, "Order data retrieved successfully", ordersAll)

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

func (h *OrderHandlers) UpdateOrder(c *gin.Context) {

	var OrderRequest models.UpdateStatusRequest

	if c.Param("id") == "" {
		SendError(c, http.StatusBadRequest, fmt.Errorf("ID cannot be empty"), models.ErrInvalidRequest)
	}

	if err := c.ShouldBindJSON(&OrderRequest); err != nil {
		SendError(c, http.StatusBadRequest, err, models.ErrInvalidRequest)
		return
	}

	if OrderRequest.Status == "" {
		SendError(c, http.StatusBadRequest, fmt.Errorf("status cannot be empty"), models.ErrInvalidRequest)
	}

	order := h.orderService.UpdateOrder(c.Request.Context(), c.Param("id"), OrderRequest)

	if order != nil {	
		SendError(c, http.StatusInternalServerError, order, models.ErrInvalidOperation)
		return
	}

	SendSuccess(c, http.StatusOK, "Update successfully", nil)
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
