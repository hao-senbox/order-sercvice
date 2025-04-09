package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Order struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	TeacherID       string             `bson:"teacher_id" json:"teacher_id"`
	Email           string             `bson:"email" json:"email"`
	TotalPrice      float64            `bson:"total_amount" json:"total_amount"`
	Status          string             `bson:"status" json:"status"`
	Items           []OrderItems       `bson:"items" json:"items"`
	ShippingAddress Address            `bson:"shipping_address" json:"shipping_address"`
	CreateAt        time.Time          `bson:"create_at" json:"create_at"`
	UpdateAt        time.Time          `bson:"update_at" json:"update_at"`
}

type OrderItems struct {
	ProductID  primitive.ObjectID `bson:"product_id" json:"product_id"`
	Quantity   int                `bson:"quantity" json:"quantity"`
	Price      float64            `bson:"price" json:"price"`
	Name       string             `bson:"name" json:"name"`
	TotalPrice float64            `bson:"total_price" json:"total_price"`
	StudentID  string             `bson:"student_id" json:"student_id"`
}

type Address struct {
	Type       string `bson:"type" json:"type"`
	Street     string `bson:"street" json:"street"`
	City       string `bson:"city" json:"city"`
	State      string `bson:"state" json:"state"`
	Country    string `bson:"country" json:"country"`
	PostalCode string `bson:"postal_code" json:"postal_code"`
	Phone      string `bson:"phone" json:"phone"`
	IsDefault  bool   `bson:"is_default" json:"is_default"`
}

type CartItem struct {
	ImageURL    string  `json:"image_url"`
	Price       float64 `json:"price"`
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
}

type StudentOrder struct {
	StudentID  string     `json:"student_id"`
	Items      []CartItem `json:"items"`
	TotalPrice float64    `json:"total_price"`
}

type GroupedOrder struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	TeacherID       string             `bson:"teacher_id" json:"teacher_id"`
	Email           string             `bson:"email" json:"email"`
	TotalPrice      float64            `bson:"total_price" json:"total_price"`
	Status          string             `bson:"status" json:"status"`
	StudentOrders   []StudentOrder     `bson:"student_orders" json:"student_orders"`
	ShippingAddress Address            `bson:"shipping_address" json:"shipping_address"`
	CreateAt        time.Time          `bson:"create_at" json:"create_at"`
	UpdateAt        time.Time          `bson:"update_at" json:"update_at"`
}

type StudentCart struct {
	StudentID  string     `json:"_id"`
	Items      []CartItem `json:"items"`
	TotalPrice float64    `json:"total_price"`
}

type CartAPIResponse struct {
	StatusCode int           `json:"status_code"`
	Message    string        `json:"message"`
	Data       []StudentCart `json:"data"`
}

type UpdateStatusRequest struct {
	Status string `json:"status"`
}

type TeacherIdRequest struct {
	TeacherID string `json:"teacher_id"`
	Email     string `json:"email"`
}
