package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type APIResponse struct {
	StatusCode int         `json:"status_code"`
	Message    string      `json:"message,omitempty"`
	Data       interface{} `json:"data,omitempty"`
	Error      string      `json:"error,omitempty"`
	ErrorCode  string      `json:"error_code,omitempty"`
}

type CartResponse struct {
	StatusCode int           `json:"status_code"`
	Message    string        `json:"message"`
	Data       []StudentCart `json:"data"`
}

type StudentCart struct {
	StudentID  string     `json:"_id"`
	Items      []CartItem `json:"items"`
	TotalPrice float64    `json:"total_price"`
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
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
}

type PaginatedResponse struct {
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalPages int         `json:"total_pages"`
	TotalItems int64       `json:"total_items"`
	Data       interface{} `json:"data"`
}

const (
	ErrInvalidOperation = "ERR_INVALID_OPERATION"
	ErrInvalidRequest   = "ERR_INVALID_REQUEST"
)
