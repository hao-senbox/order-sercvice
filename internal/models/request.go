package models

type UpdateStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=pending processing completed cancelled"`
}

type CreateOrderRequest struct {
	TeacherID string  `json:"teacher_id" validate:"required"`
	Email     string  `json:"email" validate:"required,email"`
	Types     string  `json:"types" validate:"required,oneof=cod bank_transfer"`
	Street    string  `json:"street" validate:"required"`
	City      string  `json:"city" validate:"required"`
	State     *string `json:"state"`
	Country   string  `json:"country" validate:"required"`
	Phone     string  `json:"phone" validate:"required"`
}

type SearchOrderRequest struct {
	Page        int    `form:"page" json:"page" validate:"min=1"`
	Limit       int    `form:"limit" json:"limit" validate:"min=1,max=100"`
	Status      string `form:"status" json:"status" validate:"omitempty,oneof=pending processing completed cancelled"`
	OrderNumber string `form:"order_number" json:"order_number" validate:"omitempty"`
}
