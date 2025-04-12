package models

type UpdateStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=pending processing completed cancelled"`
}

type TeacherRequest struct {
	TeacherID string `json:"teacher_id" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Types	 string `json:"types" validate:"required,oneof=cod bank_transfer"`
}

type SearchOrderRequest struct {
	Page    int    `form:"page" json:"page" validate:"min=1"`
	Limit   int    `form:"limit" json:"limit" validate:"min=1,max=100"`
	Status  string `form:"status" json:"status" validate:"omitempty,oneof=pending processing completed cancelled"`
	OrderId string `form:"order_id" json:"order_id" validate:"omitempty,hexadecimal"`
}
