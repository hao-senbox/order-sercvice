package repository

import (
	"context"
	"fmt"
	"math"
	"store/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type OrderRepository interface {
	GetOrders(ctx context.Context, req models.SearchOrderRequest) ([]*models.GroupedOrder, int64, error)
	GetOrderDetail(ctx context.Context, id primitive.ObjectID) (*models.GroupedOrder, error)
	GetOrdersUser(ctx context.Context, TeacherID string) ([]*models.GroupedOrder, error)
	CreateOrder(ctx context.Context, order models.Order) (*models.Order, error)
	UpdateOrder(ctx context.Context, id primitive.ObjectID, status string) error
	DeleteOrder(ctx context.Context, id primitive.ObjectID) error
	UpdateOrderPaymentAndStatus(ctx context.Context, id primitive.ObjectID, payment models.Payment, status string) error
	GetGroupedOrders(ctx context.Context, orders []*models.Order) ([]*models.GroupedOrder, error)
	FindUnPaidOrdersBeforeTime(ctx context.Context, timestamp time.Time, paymentMethod string) ([]*models.Order, error)
	FindOrdersForReminder(ctx context.Context, startTime, endTime time.Time) ([]*models.Order, error)
	MarkReminderSent(ctx context.Context, orderID primitive.ObjectID) error 
}

type orderRepository struct {
	collection *mongo.Collection
}

func NewOrderRepository(collection *mongo.Collection) OrderRepository {
	return &orderRepository{
		collection: collection,
	}
}

func (r *orderRepository) GetOrders(ctx context.Context, req models.SearchOrderRequest) ([]*models.GroupedOrder, int64, error) {
	filter := bson.M{}

	if req.Status != "" {
		filter["status"] = req.Status
	}

	if req.OrderNumber != "" {
		orderNumber := req.OrderNumber
		filter["order_number"] = orderNumber
	}

	totalItems, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	skip := (req.Page - 1) * req.Limit
	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(req.Limit)).
		SetSort(bson.D{{Key: "create_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var orders []*models.Order
	if err := cursor.All(ctx, &orders); err != nil {
		return nil, 0, err
	}

	groupedOrders, err := r.GetGroupedOrders(ctx, orders)
	if err != nil {
		return nil, 0, err
	}

	return groupedOrders, totalItems, nil
}

func (r *orderRepository) GetGroupedOrders(ctx context.Context, orders []*models.Order) ([]*models.GroupedOrder, error) {
	// Create a new slice for our grouped orders
	groupedOrders := make([]*models.GroupedOrder, 0, len(orders))

	// For each order, group the items by student_id
	for _, order := range orders {
		// Create a map to hold items grouped by student_id
		studentItems := make(map[string][]models.OrderItem)

		// Group items by student_id
		for i := range order.Items {
			item := order.Items[i]
			studentItems[item.StudentID] = append(studentItems[item.StudentID], item)
		}

		// Calculate totals per student
		studentTotals := make(map[string]float64)
		for studentID, items := range studentItems {
			total := 0.0
			for _, item := range items {
				total += item.TotalPrice
			}
			studentTotals[studentID] = math.Round(total*100) / 100
		}

		// Create student orders array
		studentOrders := make([]models.StudentOrder, 0, len(studentItems))
		for studentID, orderItems := range studentItems {
			// Convert OrderItems to CartItem
			cartItems := make([]models.CartItem, 0, len(orderItems))
			for _, item := range orderItems {
				cartItem := models.CartItem{
					ProductID:   item.ProductID.Hex(), // Convert ObjectID to string
					ProductName: item.Name,
					Price:       item.Price,
					Quantity:    item.Quantity,
					// Note: ImageURL is missing in OrderItems, so it will be empty
				}
				cartItems = append(cartItems, cartItem)
			}

			studentOrder := models.StudentOrder{
				StudentID:  studentID,
				Items:      cartItems,
				TotalPrice: studentTotals[studentID],
			}
			studentOrders = append(studentOrders, studentOrder)
		}

		// Create the grouped order
		groupedOrder := &models.GroupedOrder{
			ID:              order.ID,
			TeacherID:       order.TeacherID,
			OrderNumber:     order.OrderNumber,
			Email:           order.Email,
			TotalPrice:      math.Round(order.TotalPrice*100) / 100,
			Status:          order.Status,
			StudentOrders:   studentOrders,
			ShippingAddress: order.ShippingAddress,
			Payment:         order.Payment,
			CreatedAt:       order.CreatedAt,
			UpdatedAt:       order.UpdatedAt,
		}

		groupedOrders = append(groupedOrders, groupedOrder)
	}

	return groupedOrders, nil
}

func (r *orderRepository) GetOrderDetail(ctx context.Context, id primitive.ObjectID) (*models.GroupedOrder, error) {

	var order models.Order

	filter := bson.M{"_id": id}

	err := r.collection.FindOne(ctx, filter).Decode(&order)

	if err != nil {
		return nil, err
	}

	orders := []*models.Order{&order}

	orderDetail, err := r.GetGroupedOrders(ctx, orders)
	if err != nil {
		return nil, err
	}

	if len(orderDetail) > 0 {
		return orderDetail[0], nil
	}

	return nil, fmt.Errorf("order not found")

}

func (r *orderRepository) GetOrdersUser(ctx context.Context, TeacherID string) ([]*models.GroupedOrder, error) {

	var orders []*models.Order

	filter := bson.M{"teacher_id": TeacherID}

	cursor, err := r.collection.Find(ctx, filter)

	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &orders); err != nil {
		return nil, err
	}

	orderUser, err := r.GetGroupedOrders(ctx, orders)
	if err != nil {
		return nil, err
	}

	return orderUser, nil
}

func (r *orderRepository) CreateOrder(ctx context.Context, order models.Order) (*models.Order, error) {

	result, err := r.collection.InsertOne(ctx, order)

	if err != nil {
		return nil, err
	}

	order.ID = result.InsertedID.(primitive.ObjectID)

	return &order, nil

}

func (r *orderRepository) UpdateOrder(ctx context.Context, id primitive.ObjectID, status string) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	return nil
}

func (r *orderRepository) DeleteOrder(ctx context.Context, id primitive.ObjectID) error {

	filter := bson.M{"_id": id}

	result, err := r.collection.DeleteOne(ctx, filter)

	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("no record found with id: %s", id.Hex())
	}

	return nil

}

func (r *orderRepository) UpdateOrderPaymentAndStatus(ctx context.Context, id primitive.ObjectID, payment models.Payment, status string) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"payment":    payment,
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update order payment and status: %w", err)
	}

	return nil
}

func (r *orderRepository) FindUnPaidOrdersBeforeTime(ctx context.Context, timestamp time.Time, paymentMethod string) ([]*models.Order, error) {
	filter := bson.M{
		"payment.paid":   false,
		"created_at": bson.M{"$lt": timestamp},
		"status":         models.OrderStatusPending,
		"payment.method": models.PaymentMethodBankTransfer,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("error finding unpaid orders %v", err)
	}
	defer cursor.Close(ctx)

	var orders []*models.Order

	if err := cursor.All(ctx, &orders); err != nil {
		return nil, fmt.Errorf("error decoding orders: %w", err)
	}

	return orders, nil
}

func (r *orderRepository) FindOrdersForReminder(ctx context.Context, startTime, endTime time.Time) ([]*models.Order, error) {
	
	filter := bson.M{
		"payment.paid": false,
		"status": models.OrderStatusPending,
		"payment.method": models.PaymentMethodBankTransfer,
		"reminder_sent": bson.M{"$ne": true},
		"created_at": bson.M{"$gte": endTime, "$lte": startTime},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("error finding orders for reminder: %v", err)
	}
	defer cursor.Close(ctx)

	var orders []*models.Order
	if err := cursor.All(ctx, &orders); err != nil {
		return nil, fmt.Errorf("error decoding orders: %w", err)
	}

	return orders, nil
}

func (r *orderRepository) MarkReminderSent(ctx context.Context, orderID primitive.ObjectID) error {
	
	update := bson.M{
		"$set": bson.M{
			"reminder_sent": true,
			"reminder_sent_at": time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": orderID},
		update,
	)
	if err != nil {
		return fmt.Errorf("error updating reminder status: %w", err)
	}

	return nil
}

