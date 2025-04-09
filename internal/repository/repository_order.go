package repository

import (
	"context"
	"fmt"
	"store/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrderRepository interface {
	GetOrders(ctx context.Context) ([]*models.GroupedOrder, error)
	GetOrderDetail(ctx context.Context, id primitive.ObjectID) (*models.GroupedOrder, error)
	GetOrdersUser(ctx context.Context, TeacherID string) ([]*models.GroupedOrder, error)
	CreateOrder(ctx context.Context, order models.Order) (*models.Order, error)
	UpdateOrder(ctx context.Context, id primitive.ObjectID, orderRequest models.UpdateStatusRequest) error
	DeleteOrder(ctx context.Context, id primitive.ObjectID) error
	GetGroupedOrders(ctx context.Context, order []*models.Order) ([]*models.GroupedOrder, error)
}

type orderRepository struct {
	collection *mongo.Collection
}

func NewOrderRepository(collection *mongo.Collection) OrderRepository {
	return &orderRepository{
		collection: collection,
	}
}

func (r *orderRepository) GetOrders(ctx context.Context) ([]*models.GroupedOrder, error) {

	var orders []*models.Order

	cursor, err := r.collection.Find(ctx, bson.M{})

	if err != nil {
		return nil, err
	}

	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &orders); err != nil {
		return nil, err
	}

	ordersAll, err := r.GetGroupedOrders(ctx, orders)
	if err != nil {
		return nil, err
	}
	return ordersAll, nil

}

func (r *orderRepository) GetGroupedOrders(ctx context.Context, orders []*models.Order) ([]*models.GroupedOrder, error) {
	// Create a new slice for our grouped orders
	groupedOrders := make([]*models.GroupedOrder, 0, len(orders))

	// For each order, group the items by student_id
	for _, order := range orders {
		// Create a map to hold items grouped by student_id
		studentItems := make(map[string][]models.OrderItems)

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
			studentTotals[studentID] = total
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
			TeacherID:          order.TeacherID,
			Email:           order.Email,
			TotalPrice:      order.TotalPrice,
			Status:          order.Status,
			StudentOrders:   studentOrders,
			ShippingAddress: order.ShippingAddress,
			CreateAt:        order.CreateAt,
			UpdateAt:        order.UpdateAt,
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

func (r *orderRepository) UpdateOrder(ctx context.Context, id primitive.ObjectID, orderRequest models.UpdateStatusRequest) error {

	filter := bson.M{"_id": id}

	update := bson.M{
		"$set": bson.M{
			"status": orderRequest.Status,
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)

	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("no record found with id: %s", id.Hex())
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
