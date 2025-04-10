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
	m := gomail.NewMessage()
	m.SetHeader("From", os.Getenv("EMAIL_FROM"))
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Thank you for your order")

	// Create email body with CSS styling
	body := `
		<html>
		<head>
			<style>
				body {
					font-family: Arial, sans-serif;
					line-height: 1.6;
					color: #333;
					margin: 0;
					padding: 0;
				}
				.header {
					background-color: #8b4b8b;
					color: white;
					padding: 20px;
					text-align: left;
					font-size: 24px;
				}
				.container {
					max-width: 600px;
					margin: 0 auto;
					padding: 20px;
				}
				.message {
					margin: 20px 0;
					line-height: 1.6;
				}
				.order-info {
					margin: 20px 0;
					color: #8b4b8b;
				}
				table {
					width: 100%;
					border-collapse: collapse;
					margin: 20px 0;
				}
				th {
					background-color: #f5f5f5;
					padding: 10px;
					text-align: left;
					border-bottom: 1px solid #ddd;
				}
				td {
					padding: 10px;
					border-bottom: 1px solid #ddd;
				}
				.subtotal {
					font-weight: bold;
				}
				.shipping {
					padding: 10px 0;
				}
				.shipping-title {
					font-weight: bold;
				}
				.payment {
					padding: 10px 0;
					border-bottom: 1px solid #ddd;
				}
				.total {
					font-weight: bold;
					padding: 10px 0;
				}
				.billing-address {
					margin-top: 20px;
					color: #8b4b8b;
				}
				.address-details {
					margin-top: 10px;
					line-height: 1.4;
				}
			</style>
		</head>
		<body>
			<div class="header">
				Thank you for your order
			</div>
			<div class="container">
				<div class="message">
					Your order has been received and is now being processed. Your order details are shown below for your reference:
				</div>
				<div class="order-info">
					Order #` + order.ID.Hex() + ` (` + order.CreatedAt.Format("January 2, 2006") + `)
				</div>
				<table>
					<tr>
						<th>Product</th>
						<th>Quantity</th>
						<th>Price</th>
					</tr>`

	// Add order items
	var subtotal float64
	for _, studentOrder := range order.StudentOrders {
		for _, item := range studentOrder.Items {
			itemTotal := float64(item.Quantity) * item.Price
			subtotal += itemTotal
			body += fmt.Sprintf(`
					<tr>
						<td>%s</td>
						<td>%d</td>
						<td>$%.2f</td>
					</tr>`, item.ProductName, item.Quantity, item.Price)
		}
	}

	// Add subtotal, shipping, and total
	body += fmt.Sprintf(`
				</table>
				<div class="subtotal">
					Subtotal: $%.2f
				</div>
				<div class="shipping">
					<span class="shipping-title">Shipping:</span> $0.00 via Local pickup
				</div>
				<div class="shipping">
					<span class="shipping-title">Shipping address:</span><br>
					%s<br>
					<a href="#">View on Google Maps</a>
				</div>
				<div class="payment">
					<span class="shipping-title">Payment method:</span> Direct bank transfer
				</div>
				<div class="total">
					Total: $%.2f (includes $%.2f Tax)
				</div>
				<div class="billing-address">
					<h3>Billing address</h3>
					<div class="address-details">
						%s<br>
						%s<br>
						%s<br>
						%s %s<br>
						Phone: %s
					</div>
				</div>
			</div>
		</body>
		</html>`,
		subtotal,
		order.ShippingAddress.Street,
		order.TotalPrice,
		order.TotalPrice*0.1,
		order.ShippingAddress.Street,
		order.ShippingAddress.City,
		order.ShippingAddress.State,
		order.ShippingAddress.PostalCode,
		order.ShippingAddress.Country,
		order.ShippingAddress.Phone)

	m.SetBody("text/html", body)

	if err := es.dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (es *EmailService) SendOrderConfirmationUpdate(email string, order *models.GroupedOrder) error {
	m := gomail.NewMessage()
	m.SetHeader("From", os.Getenv("EMAIL_FROM"))
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Đơn hàng của bạn đã được xác nhận")

	body := `
		<html>
		<head>
			<style>
				body {
					font-family: Arial, sans-serif;
					line-height: 1.6;
					color: #333;
					margin: 0;
					padding: 0;
				}
				.header {
					background-color: #27ae60;
					color: white;
					padding: 20px;
					text-align: center;
					font-size: 24px;
				}
				.container {
					max-width: 600px;
					margin: 0 auto;
					padding: 20px;
				}
				.confirmation-message {
					background-color: #e8f5e9;
					border-radius: 4px;
					padding: 20px;
					margin: 20px 0;
					text-align: center;
					color: #2e7d32;
					font-size: 18px;
				}
				.order-info {
					margin: 20px 0;
					padding: 15px;
					background-color: #f9f9f9;
					border-radius: 4px;
				}
				.order-number {
					color: #27ae60;
					font-weight: bold;
				}
				table {
					width: 100%;
					border-collapse: collapse;
					margin: 20px 0;
				}
				th {
					background-color: #f5f5f5;
					padding: 12px;
					text-align: left;
					border-bottom: 2px solid #27ae60;
				}
				td {
					padding: 12px;
					border-bottom: 1px solid #ddd;
				}
				.total {
					font-size: 18px;
					color: #27ae60;
					font-weight: bold;
					text-align: right;
					padding: 15px 0;
				}
				.next-steps {
					background-color: #f9f9f9;
					padding: 20px;
					border-radius: 4px;
					margin: 20px 0;
				}
				.next-steps h3 {
					color: #2c3e50;
					margin-top: 0;
				}
				.shipping-info {
					margin-top: 20px;
					padding: 15px;
					background-color: #f9f9f9;
					border-radius: 4px;
				}
			</style>
		</head>
		<body>
			<div class="header">
				Order confirmed
			</div>
			<div class="container">
				<div class="confirmation-message">
					Your order has been confirmed and is being prepared!
				</div>
				<div class="order-info">
					<span class="order-number">Order #` + order.ID.Hex() + `</span>
					<p>Thank you for placing your order. Your order has been confirmed and is being processed.</p>
				</div>
				<table>
					<tr>
						<th>Product</th>
						<th>Quantity</th>
						<th>Price</th>
					</tr>`

	var subtotal float64
	for _, studentOrder := range order.StudentOrders {
		for _, item := range studentOrder.Items {
			itemTotal := float64(item.Quantity) * item.Price
			subtotal += itemTotal
			body += fmt.Sprintf(`
					<tr>
						<td>%s</td>
						<td>%d</td>
						<td>$%.2f</td>
					</tr>`, item.ProductName, item.Quantity, item.Price)
		}
	}

	body += fmt.Sprintf(`
				</table>
				<div class="total">
					Total: $%.2f
				</div>
				<div class="shipping-info">
					<h3>Delivery information</h3>
					<p>
						%s<br>
						%s, %s<br>
						%s %s<br>
						Phone: %s
					</p>
				</div>
				<div class="next-steps">
					<h3>Next steps</h3>
					<p>1. Your order is being prepared</p>
					<p>2. You will receive an email when your order is delivered to the shipping carrier.</p>
					<p>3. Track your order through your account</p>
				</div>
			</div>
		</body>
		</html>`,
		order.TotalPrice,
		order.ShippingAddress.Street,
		order.ShippingAddress.City,
		order.ShippingAddress.State,
		order.ShippingAddress.PostalCode,
		order.ShippingAddress.Country,
		order.ShippingAddress.Phone)

	m.SetBody("text/html", body)

	if err := es.dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send confirmation email: %w", err)
	}

	return nil
}
