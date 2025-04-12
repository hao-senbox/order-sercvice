package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	OrderStatusPending     = "pending"
	OrderStatusProcessing  = "processing"
	OrderStatusCompleted   = "completed"
	OrderStatusCanceled    = "canceled"
)

type Order struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	OrderNumber     string             `bson:"order_number" json:"order_number"`
	TeacherID       string             `bson:"teacher_id" json:"teacher_id"`
	Email           string             `bson:"email" json:"email"`
	TotalPrice      float64            `bson:"total_price" json:"total_price"`
	Status          string        `bson:"status" json:"status"`
	Items           []OrderItem        `bson:"items" json:"items"`
	ShippingAddress Address            `bson:"shipping_address" json:"shipping_address"`
	Payment         Payment            `bson:"payment" json:"payment"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
}

type OrderItem struct {
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

type Payment struct {
	Method          string     `bson:"method" json:"method"`
	Paid            bool       `bson:"paid" json:"paid"`
	TransferContent *string    `bson:"transfer_content" json:"transfer_content"`
	PaidAt          *time.Time `bson:"paid_at" json:"paid_at"`
}

type CartItem struct {
	ImageURL    string  `json:"image_url"`
	Price       float64 `json:"price"`
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
}

type BankAccount struct {
	AccountName   string `bson:"account_name" json:"account_name"`
	AccountNumber string `bson:"account_number" json:"account_number"`
	BankName      string `bson:"bank_name" json:"bank_name"`
}

const (
	PaymentMethodBankTransfer = "bank_transfer"
	PaymentMethodCOD          = "cod"
)
