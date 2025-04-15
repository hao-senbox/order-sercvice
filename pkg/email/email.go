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
	m.SetHeader("From", os.Getenv("SMTP_USER"))
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Thank you for your order")

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
				.shipping-info {
					margin-top: 20px;
					padding: 15px;
					background-color: #f9f9f9;
					border-radius: 4px;
				}
				.note-cod {
					background-color: #fff3cd;
					padding: 20px;
					border-left: 5px solid #ffc107;
					margin: 20px 0;
					border-radius: 4px;
				}
				.footer {
					text-align: center;
					margin-top: 30px;
					color: #666;
					font-size: 14px;
				}
			</style>
		</head>
		<body>
			<div class="header">
				Order Created
			</div>
			<div class="container">
				<div class="confirmation-message">
					Your order has been successfully created!
				</div>
				<div class="order-info">
					<span class="order-number">Order #` + order.OrderNumber + `</span>
					<p>Thank you for placing your order. Your order has been received and is currently being processed.</p>
				</div>
				<table>
					<tr>
						<th>Product</th>
						<th>Quantity</th>
						<th>Price</th>
					</tr>` + generateOrderItemsTable(order) + `
				</table>
				<div class="total">
					Total: $` + fmt.Sprintf("%.2f", order.TotalPrice) + `
				</div>
				<div class="note-cod">
					<h3 style="margin-top: 0; color: #856404;">Cash on Delivery (COD) Information</h3>
					<p>💵 Please prepare the exact amount in cash to pay the delivery personnel upon receiving your order.</p>
				</div>
				<div class="shipping-info">
					<h3>Delivery Information</h3>
					<p>` + formatAddress(order.ShippingAddress) + `</p>
				</div>
				<div class="footer">
					<p>We will notify you once your order has been confirmed and is being processed.</p>
					<p>© 2024 Your Store. All rights reserved.</p>
				</div>
			</div>
		</body>
		</html>`

	m.SetBody("text/html", body)

	if err := es.dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send confirmation email: %w", err)
	}

	return nil
}

func (es *EmailService) SendOrderConfirmationUpdate(email string, order *models.GroupedOrder) error {
	m := gomail.NewMessage()
	m.SetHeader("From", os.Getenv("SMTP_USER"))
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Your order has been confirmed.")

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
					<span class="order-number">Order #` + order.OrderNumber + `</span>
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
	state := ""
	if order.ShippingAddress.State != nil {
		state = *order.ShippingAddress.State
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
						%s %s %s<br>
						Phone: %s
					</p>
				</div>
				<div class="next-steps">
					<h3>Next steps</h3>
					<p>1. Your order is being prepared</p>
					<p>2. Please keep an eye on your phone to receive the goods from the delivery person.</p>
					<p>3. Track your order through your account</p>
				</div>
			</div>
		</body>
		</html>`,
		order.TotalPrice,
		order.ShippingAddress.Street,
		order.ShippingAddress.City,
		state,
		order.ShippingAddress.Country,
		order.ShippingAddress.Phone)

	m.SetBody("text/html", body)

	if err := es.dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send confirmation email: %w", err)
	}

	return nil
}

func (es *EmailService) SendOrderConfirmationBank(email string, order *models.GroupedOrder, bank models.BankAccount) error {
	m := gomail.NewMessage()
	m.SetHeader("From", os.Getenv("SMTP_USER"))
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Thank you for your order")

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
				.bank-info {
					background-color: #fff3cd;
					padding: 20px;
					border-left: 5px solid #ffc107;
					margin: 20px 0;
					border-radius: 4px;
				}
				.shipping-info {
					margin-top: 20px;
					padding: 15px;
					background-color: #f9f9f9;
					border-radius: 4px;
				}
				.footer {
					text-align: center;
					margin-top: 30px;
					color: #666;
					font-size: 14px;
				}
			</style>
		</head>
			<body>
				<div class="header">
					Order Created
				</div>
				<div class="container">
					<div class="confirmation-message">
						Your order has been successfully created!
					</div>
					<div class="order-info">
						<span class="order-number">Order #` + order.OrderNumber + `</span>
						<p>Thank you for placing your order. We have received your order and it is currently pending confirmation.</p>
						<div>
							<div class="bank-info">
								<h3 style="margin-top: 0; color: #856404;">Bank Transfer Information</h3>
								<p><strong>Bank Name:</strong> ` + bank.BankName + `</p>
								<p><strong>Account Holder:</strong> ` + bank.AccountName + `</p>
								<p><strong>Account Number:</strong> ` + bank.AccountNumber + `</p>
								<p><strong>Total Price:</strong> $` + fmt.Sprintf("%.2f", order.TotalPrice) + `</p>
								<p><strong>Transfer Note:</strong> <span style="color: #c0392b;">` + *order.Payment.TransferContent + `</span></p>
								<p style="margin-top: 10px;">⚠️ Please make sure to include the correct transfer note to help us verify your payment quickly.</p>
							</div>
						<div>
					</div>
					<table>
						<tr>
							<th>Product</th>
							<th>Quantity</th>
							<th>Price</th>
						</tr>` + generateOrderItemsTable(order) + `
					</table>
					<div class="total">
						Total: $` + fmt.Sprintf("%.2f", order.TotalPrice) + `
					</div>
					<div class="shipping-info">
						<h3>Delivery Information</h3>
						<p>` + formatAddress(order.ShippingAddress) + `</p>
					</div>
				</div>
			</body>

		</html>`

	m.SetBody("text/html", body)

	if err := es.dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send confirmation email: %w", err)
	}

	return nil
}

// Helper function to generate order items table
func generateOrderItemsTable(order *models.GroupedOrder) string {
	var tableRows string
	for _, studentOrder := range order.StudentOrders {
		for _, item := range studentOrder.Items {
			tableRows += fmt.Sprintf(`
					<tr>
						<td>%s</td>
						<td>%d</td>
						<td>$%.2f</td>
					</tr>`, item.ProductName, item.Quantity, item.Price)
		}
	}
	return tableRows
}

// Helper function to format address
func formatAddress(address models.Address) string {
	state := ""
	if address.State != nil {
		state = *address.State
	}
	return fmt.Sprintf(`%s<br>
		%s, %s<br>
		%s<br>
		Phone: %s`,
		address.Street,
		address.City,
		state,
		address.Country,
		address.Phone)
}

func (es *EmailService) SendEmailNotificationAdmin(email string, order *models.GroupedOrder) error {
	m := gomail.NewMessage()
	m.SetHeader("From", os.Getenv("SMTP_USER"))
	m.SetHeader("To", email)
	m.SetHeader("Subject", "New Order Notification")

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
					background-color: #3498db;
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
				.notification {
					background-color: #e3f2fd;
					border-radius: 4px;
					padding: 20px;
					margin: 20px 0;
					text-align: center;
					color: #0d47a1;
					font-size: 18px;
				}
				.order-info {
					margin: 20px 0;
					padding: 15px;
					background-color: #f9f9f9;
					border-radius: 4px;
				}
				.order-number {
					color: #3498db;
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
					border-bottom: 2px solid #3498db;
				}
				td {
					padding: 12px;
					border-bottom: 1px solid #ddd;
				}
				.total {
					font-size: 18px;
					color: #3498db;
					font-weight: bold;
					text-align: right;
					padding: 15px 0;
				}
				.shipping-info {
					margin-top: 20px;
					padding: 15px;
					background-color: #f9f9f9;
					border-radius: 4px;
				}
				.action-button {
					display: inline-block;
					background-color: #3498db;
					color: white;
					padding: 10px 20px;
					text-decoration: none;
					border-radius: 4px;
					margin: 20px 0;
				}
				.action-button:hover {
					background-color: #2980b9;
				}
				.payment-info {
					background-color: #e3f2fd;
					padding: 20px;
					border-left: 5px solid #3498db;
					margin: 20px 0;
					border-radius: 4px;
				}
			</style>
		</head>
		<body>
			<div class="header">
				New Order Received
			</div>
			<div class="container">
				<div class="notification">
					A new order has been placed and requires your attention.
				</div>
				<div class="order-info">
					<span class="order-number">Order #` + order.OrderNumber + `</span>
					<p>Order Date: ` + order.CreatedAt.Format("January 2, 2006 15:04:05") + `</p>
					<p>Status: ` + order.Status + `</p>
				</div>
				<table>
					<tr>
						<th>Product</th>
						<th>Quantity</th>
						<th>Price</th>
					</tr>` + generateOrderItemsTable(order) + `
				</table>
				<div class="total">
					Total: $` + fmt.Sprintf("%.2f", order.TotalPrice) + `
				</div>
				<div class="payment-info">
					<h3 style="margin-top: 0; color: #0d47a1;">Payment Information</h3>
					<p><strong>Payment Method:</strong> ` + order.Payment.Method + `</p>
					<p><strong>Payment Status:</strong> ` + order.Status + `</p>
					<p><strong>Transfer Note:</strong> ` + *order.Payment.TransferContent + `</p>
				</div>
				<div class="shipping-info">
					<h3>Delivery Information</h3>
					<p>` + formatAddress(order.ShippingAddress) + `</p>
				</div>
				<div style="text-align: center;">
					<a href="https://admin.yourstore.com/orders" class="action-button">View Order Details</a>
				</div>
			</div>
		</body>
		</html>`

	m.SetBody("text/html", body)

	if err := es.dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send admin notification email: %w", err)
	}

	return nil
}

func (es *EmailService) SendOrderCancellation(email string, order *models.GroupedOrder) error {
	m := gomail.NewMessage()
	m.SetHeader("From", os.Getenv("SMTP_USER"))
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Order Cancellation Notice")

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
					background-color: #e74c3c;
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
				.message {
					background-color: #fde8e7;
					border-radius: 4px;
					padding: 20px;
					margin: 20px 0;
					text-align: center;
					color: #c0392b;
					font-size: 18px;
				}
				.order-info {
					margin: 20px 0;
					padding: 15px;
					background-color: #f9f9f9;
					border-radius: 4px;
				}
				.order-number {
					color: #e74c3c;
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
					border-bottom: 2px solid #e74c3c;
				}
				td {
					padding: 12px;
					border-bottom: 1px solid #ddd;
				}
				.total {
					font-size: 18px;
					color: #e74c3c;
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
			</style>
		</head>
		<body>
			<div class="header">
				Order Cancelled
			</div>
			<div class="container">
				<div class="message">
					Your order has been cancelled due to non-payment
				</div>
				<div class="order-info">
					<span class="order-number">Order #` + order.OrderNumber + `</span>
					<p>We regret to inform you that your order has been cancelled as we did not receive payment within the required timeframe.</p>
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
				<div class="next-steps">
					<h3>What to do next?</h3>
					<p>1. You can place a new order if you still wish to purchase these items</p>
					<p>2. Contact our customer support if you have any questions</p>
				</div>
			</div>
		</body>
		</html>`, order.TotalPrice)

	m.SetBody("text/html", body)

	if err := es.dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send cancellation email: %w", err)
	}

	return nil
}

func (es *EmailService) SendReminderOrder(email string, order *models.GroupedOrder, hoursLeft int, bank models.BankAccount) error {
	m := gomail.NewMessage()
	m.SetHeader("From", os.Getenv("SMTP_USER"))
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Payment Reminder for Your Order")

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
					background-color: #f39c12;
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
				.message {
					background-color: #fff3cd;
					border-radius: 4px;
					padding: 20px;
					margin: 20px 0;
					text-align: center;
					color: #85640a;
					font-size: 18px;
				}
				.order-info {
					margin: 20px 0;
					padding: 15px;
					background-color: #f9f9f9;
					border-radius: 4px;
				}
				.order-number {
					color: #f39c12;
					font-weight: bold;
				}
				.reminder {
					color: #d35400;
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
					border-bottom: 2px solid #f39c12;
				}
				td {
					padding: 12px;
					border-bottom: 1px solid #ddd;
				}
				.total {
					font-size: 18px;
					color: #f39c12;
					font-weight: bold;
					text-align: right;
					padding: 15px 0;
				}
				.action-button {
					display: inline-block;
					padding: 10px 20px;
					background-color: #27ae60;
					color: white;
					text-decoration: none;
					border-radius: 5px;
					margin-top: 15px;
				}
				.action-button:hover {
					background-color: #219653;
				}
				.bank-info {
					background-color: #fff3cd;
					padding: 20px;
					border-left: 5px solid #ffc107;
					margin: 20px 0;
					border-radius: 4px;
				}
			</style>
		</head>
		<body>
			<div class="header">
				Payment Reminder
			</div>
			<div class="container">
				<div class="message">
					Don't forget to complete your order!
				</div>
				<div class="order-info">
					<span class="order-number">Order #` + order.OrderNumber + `</span>
					<p>This is a friendly reminder that your order is awaiting payment. You have <span class="reminder">` + fmt.Sprintf("%d hours", hoursLeft) + `</span> left to complete your purchase before it is automatically cancelled.</p>
					<div>
						<div class="bank-info">
							<h3 style="margin-top: 0; color: #856404;">Bank Transfer Information</h3>
							<p><strong>Bank Name:</strong> ` + bank.BankName + `</p>
							<p><strong>Account Holder:</strong> ` + bank.AccountName + `</p>
							<p><strong>Account Number:</strong> ` + bank.AccountNumber + `</p>
							<p><strong>Total Price:</strong> $` + fmt.Sprintf("%.2f", order.TotalPrice) + `</p>
							<p><strong>Transfer Note:</strong> <span style="color: #c0392b;">` + *order.Payment.TransferContent + `</span></p>
							<p style="margin-top: 10px;">⚠️ Please make sure to include the correct transfer note to help us verify your payment quickly.</p>
						</div>
					<div>
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
				<p>If you have already made the payment, please disregard this email.</p>
				<p>If you have any questions, please contact our customer support.</p>
			</div>
		</body>
		</html>`, order.TotalPrice)

	m.SetBody("text/html", body)

	if err := es.dialer.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send payment reminder email: %w", err)
	}

	return nil
}
