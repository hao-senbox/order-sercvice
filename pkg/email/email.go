package email

import (
	"fmt"
	"os"
	"store/internal/models"
	"strconv"
	"gopkg.in/gomail.v2"
)

type EmailService struct {
	dialer *gomail.Dialer
}

func NewEmailService() *EmailService {

	smtpHost := os.Getenv("EMAIL_HOST")

	smtpPortStr := os.Getenv("EMAIL_PORT")
	smtpPort := 587
	if portNum, err := strconv.Atoi(smtpPortStr); err == nil && portNum > 0 {
		smtpPort = portNum
	}

	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASSWORD")

	dialer := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)

	return &EmailService{
		dialer: dialer,
	}
}

func (es *EmailService) SendOrderConfirmation(email string, order *models.GroupedOrder) error {
	// Create email message
	m := gomail.NewMessage()
	m.SetHeader("From", os.Getenv("EMAIL_FROM"))
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Order Confirmation")

	// Create email body with CSS styling
	body := `
        <html>
        <head>
            <style>
                body {
                    font-family: 'Helvetica Neue', Arial, sans-serif;
                    line-height: 1.6;
                    color: #333;
                    margin: 0;
                    padding: 0;
                    background-color: #f4f4f4;
                }
                .container {
                    max-width: 600px;
                    margin: 0 auto;
                    padding: 20px;
                    background-color: #ffffff;
                    border-radius: 8px;
                    box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
                }
                .header {
                    text-align: center;
                    padding: 20px 0;
                    border-bottom: 1px solid #eaeaea;
                }
                .header h2 {
                    color: #2c3e50;
                    margin: 0;
                    font-size: 24px;
                }
                .order-details {
                    padding: 20px 0;
                    border-bottom: 1px solid #eaeaea;
                }
                .order-details p {
                    margin: 8px 0;
                }
                .order-id {
                    font-weight: bold;
                    color: #2c3e50;
                }
                .total-amount {
                    font-size: 18px;
                    font-weight: bold;
                    color: #27ae60;
                }
                .status {
                    display: inline-block;
                    padding: 5px 10px;
                    background-color: #3498db;
                    color: white;
                    border-radius: 4px;
                    font-size: 14px;
                }
                .order-items {
                    padding: 20px 0;
                }
                .order-items h3 {
                    margin-top: 0;
                    color: #2c3e50;
                    font-size: 18px;
                }
                .item-list {
                    list-style-type: none;
                    padding: 0;
                }
                .item {
                    padding: 15px;
                    margin-bottom: 10px;
                    background-color: #f9f9f9;
                    border-radius: 4px;
                    border-left: 4px solid #3498db;
                }
                .item-name {
                    font-weight: bold;
                    color: #2c3e50;
                }
                .item-price {
                    color: #7f8c8d;
                }
                .item-total {
                    font-weight: bold;
                    color: #27ae60;
                }
                .shipping-address {
                    padding: 20px 0;
                    border-top: 1px solid #eaeaea;
                }
                .shipping-address h3 {
                    margin-top: 0;
                    color: #2c3e50;
                    font-size: 18px;
                }
                .address-box {
                    padding: 15px;
                    background-color: #f9f9f9;
                    border-radius: 4px;
                    border-left: 4px solid #e67e22;
                }
                .footer {
                    text-align: center;
                    padding-top: 20px;
                    color: #7f8c8d;
                    font-size: 14px;
                }
            </style>
        </head>
        <body>
            <div class="container">
                <div class="header">
                    <h2>Thank you for your order!</h2>
                </div>
                <div class="order-details">
                    <p class="order-id">Order ID: ` + order.ID.Hex() + `</p>
                    <p class="total-amount">Total Amount: $` + fmt.Sprintf("%.2f", order.TotalPrice) + `</p>
                    <p>Status: <span class="status">Awaiting confirmation</span></p>
                </div>
                <div class="order-items">
                    <h3>Order Items</h3>
                    <ul class="item-list">
    `

	// Add order items to email body
	for _, studentOrder := range order.StudentOrders {
		body += `
				<h4 style="color: #2980b9; margin-bottom: 10px;">Student ID: ` + studentOrder.StudentID + `</h4>
				<ul class="item-list">
			`
		for _, item := range studentOrder.Items {
			body += `
								<li class="item">
									<div class="item-name">` + item.ProductName + `</div>
									<div>Quantity: ` + fmt.Sprintf("%d", item.Quantity) + `</div>
									<div class="item-price">Price: $` + fmt.Sprintf("%.2f", item.Price) + `</div>
									<div class="item-total">Total: $` + fmt.Sprintf("%.2f", float64(item.Quantity)*item.Price) + `</div>
								</li>
				`
		}
	}

	// Complete with shipping address and next steps
	body += `
				</ul>
                    </ul>
                </div>
                <div class="shipping-address">
                    <h3>Shipping Address</h3>
                    <div class="address-box">
                        <p>` + order.ShippingAddress.Street + `<br>
                        ` + order.ShippingAddress.City + `, ` + order.ShippingAddress.State + `, ` + order.ShippingAddress.Country + `<br>
                        ` + order.ShippingAddress.PostalCode + `<br>
                        Phone: ` + order.ShippingAddress.Phone + `</p>
                    </div>
                </div>
                <div class="next-steps">
                    <h3>What's Next?</h3>
                    <p>Your order is now being prepared for shipping. You will receive a shipping notification when your order is on its way.</p>
                </div>
                <div class="footer">
                    <p>Thank you for your order! If you have any questions, please contact our customer support.</p>
                </div>
            </div>
        </body>
        </html>
    `

	m.SetBody("text/html", body)

	// Send email
	if err := es.dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// Add this method to your EmailService
func (es *EmailService) SendOrderConfirmationUpdate(email string, order *models.GroupedOrder) error {
	// Create email message
	m := gomail.NewMessage()
	m.SetHeader("From", os.Getenv("EMAIL_FROM"))
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Your Order Has Been Confirmed")

	// Create email body with CSS styling
	body := `
        <html>
        <head>
            <style>
                body {
                    font-family: 'Helvetica Neue', Arial, sans-serif;
                    line-height: 1.6;
                    color: #333;
                    margin: 0;
                    padding: 0;
                    background-color: #f4f4f4;
                }
                .container {
                    max-width: 600px;
                    margin: 0 auto;
                    padding: 20px;
                    background-color: #ffffff;
                    border-radius: 8px;
                    box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
                }
                .header {
                    text-align: center;
                    padding: 20px 0;
                    border-bottom: 1px solid #eaeaea;
                }
                .header h2 {
                    color: #2c3e50;
                    margin: 0;
                    font-size: 24px;
                }
                .confirmation-message {
                    padding: 20px;
                    text-align: center;
                    background-color: #e8f5e9;
                    border-radius: 4px;
                    margin: 20px 0;
                    color: #2e7d32;
                    font-size: 18px;
                }
                .order-details {
                    padding: 20px 0;
                    border-bottom: 1px solid #eaeaea;
                }
                .order-details p {
                    margin: 8px 0;
                }
                .order-id {
                    font-weight: bold;
                    color: #2c3e50;
                }
                .total-amount {
                    font-size: 18px;
                    font-weight: bold;
                    color: #27ae60;
                }
                .status {
                    display: inline-block;
                    padding: 5px 10px;
                    background-color: #27ae60;
                    color: white;
                    border-radius: 4px;
                    font-size: 14px;
                }
                .order-items {
                    padding: 20px 0;
                }
                .order-items h3 {
                    margin-top: 0;
                    color: #2c3e50;
                    font-size: 18px;
                }
                .item-list {
                    list-style-type: none;
                    padding: 0;
                }
                .item {
                    padding: 15px;
                    margin-bottom: 10px;
                    background-color: #f9f9f9;
                    border-radius: 4px;
                    border-left: 4px solid #3498db;
                }
                .item-name {
                    font-weight: bold;
                    color: #2c3e50;
                }
                .item-price {
                    color: #7f8c8d;
                }
                .item-total {
                    font-weight: bold;
                    color: #27ae60;
                }
                .shipping-address {
                    padding: 20px 0;
                    border-top: 1px solid #eaeaea;
                }
                .shipping-address h3 {
                    margin-top: 0;
                    color: #2c3e50;
                    font-size: 18px;
                }
                .address-box {
                    padding: 15px;
                    background-color: #f9f9f9;
                    border-radius: 4px;
                    border-left: 4px solid #e67e22;
                }
                .next-steps {
                    padding: 20px 0;
                    border-top: 1px solid #eaeaea;
                }
                .footer {
                    text-align: center;
                    padding-top: 20px;
                    color: #7f8c8d;
                    font-size: 14px;
                }
            </style>
        </head>
        <body>
            <div class="container">
                <div class="header">
                    <h2>Order Confirmation</h2>
                </div>
                <div class="confirmation-message">
                    Your order has been confirmed and is being processed!
                </div>
                <div class="order-details">
                    <p class="order-id">Order ID: ` + order.ID.Hex() + `</p>
                    <p class="total-amount">Total Amount: $` + fmt.Sprintf("%.2f", order.TotalPrice) + `</p>
                    <p>Status: <span class="status">Confirmed</span></p>
                </div>
                <div class="order-items">
                    <h3>Order Items</h3>
                    <ul class="item-list">
    `
	
	// Add order items to email body
	for _, studentOrder := range order.StudentOrders {
		body += `
				<h4 style="color: #2980b9; margin-bottom: 10px;">Student ID: ` + studentOrder.StudentID + `</h4>
				<ul class="item-list">
			`
		for _, item := range studentOrder.Items {
			body += `
								<li class="item">
									<div class="item-name">` + item.ProductName + `</div>
									<div>Quantity: ` + fmt.Sprintf("%d", item.Quantity) + `</div>
									<div class="item-price">Price: $` + fmt.Sprintf("%.2f", item.Price) + `</div>
									<div class="item-total">Total: $` + fmt.Sprintf("%.2f", float64(item.Quantity)*item.Price) + `</div>
								</li>
				`
		}
	}

	// Complete with shipping address and next steps
	body += `
				</ul>
                    </ul>
                </div>
                <div class="shipping-address">
                    <h3>Shipping Address</h3>
                    <div class="address-box">
                        <p>` + order.ShippingAddress.Street + `<br>
                        ` + order.ShippingAddress.City + `, ` + order.ShippingAddress.State + `, ` + order.ShippingAddress.Country + `<br>
                        ` + order.ShippingAddress.PostalCode + `<br>
                        Phone: ` + order.ShippingAddress.Phone + `</p>
                    </div>
                </div>
                <div class="next-steps">
                    <h3>What's Next?</h3>
                    <p>Your order is now being prepared for shipping. You will receive a shipping notification when your order is on its way.</p>
                </div>
                <div class="footer">
                    <p>Thank you for your order! If you have any questions, please contact our customer support.</p>
                </div>
            </div>
        </body>
        </html>
    `

	m.SetBody("text/html", body)

	// Send email
	if err := es.dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send confirmation email: %w", err)
	}

	return nil
}
