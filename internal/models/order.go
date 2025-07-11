package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	OrderStatusPending    = "pending"
	OrderStatusProcessing = "processing"
	OrderStatusCompleted  = "completed"
	OrderStatusCanceled   = "canceled"
)

type Order struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	OrderNumber       string             `bson:"order_number" json:"order_number"`
	TeacherID         string             `bson:"teacher_id" json:"teacher_id"`
	Email             string             `bson:"email" json:"email"`
	TotalPriceStore   float64            `bson:"total_price_store" json:"total_price_store"`
	TotalPriceService float64            `bson:"total_price_service" json:"total_price_service"`
	Status            string             `bson:"status" json:"status"`
	Items             []OrderItem        `bson:"items" json:"items"`
	ShippingAddress   Address            `bson:"shipping_address" json:"shipping_address"`
	Payment           Payment            `bson:"payment" json:"payment"`
	ReminderSent      bool               `bson:"reminder_sent,omitempty"`
	ReminderSentAt    *time.Time         `bson:"reminder_sent_at,omitempty"`
	CreatedAt         time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Time          `bson:"updated_at" json:"updated_at"`
}

type OrderItem struct {
	ProductID         primitive.ObjectID `bson:"product_id" json:"product_id"`
	Quantity          int                `bson:"quantity" json:"quantity"`
	PriceStore        float64            `bson:"price_store" json:"price_store"`
	PriceService      float64            `bson:"price_service" json:"price_service"`
	Name              string             `bson:"name" json:"name"`
	TotalPriceStore   float64            `bson:"total_price_store" json:"total_price_store"`
	TotalPriceService float64            `bson:"total_price_service" json:"total_price_service"`
	StudentID         string             `bson:"student_id" json:"student_id"`
}

type Address struct {
	Street  string  `bson:"street" json:"street"`
	City    string  `bson:"city" json:"city"`
	State   *string `bson:"state,omitempty" json:"state"`
	Country string  `bson:"country" json:"country"`
	Phone   string  `bson:"phone" json:"phone"`
}

type Payment struct {
	Method          string     `bson:"method" json:"method"`
	Paid            bool       `bson:"paid" json:"paid"`
	TransferContent *string    `bson:"transfer_content" json:"transfer_content"`
	PaidAt          *time.Time `bson:"paid_at" json:"paid_at"`
}

type CartItem struct {
	ImageURL     string  `json:"image_url"`
	PriceStore   float64 `json:"price_store"`
	PriceService float64 `json:"price_service"`
	ProductID    string  `json:"product_id"`
	ProductName  string  `json:"product_name"`
	Quantity     int     `json:"quantity"`
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
