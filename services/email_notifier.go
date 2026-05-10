package services

import (
	"context"
	"fmt"
	"html/template"
	"os"
	"strings"

	"github.com/resend/resend-go/v3"
)

type PackPurchaseEmailPayload struct {
	ToEmail    string
	TenantID   int64
	OrderID    string
	CatalogID  string
	PackName   string
	ReceiverID string
	ICCID      string
	QRCode     string
}

type EmailNotifier interface {
	SendPackPurchaseQRCode(ctx context.Context, payload PackPurchaseEmailPayload) error
}

type noopEmailNotifier struct{}

func (n *noopEmailNotifier) SendPackPurchaseQRCode(_ context.Context, _ PackPurchaseEmailPayload) error {
	return nil
}

type resendEmailNotifier struct {
	client    *resend.Client
	fromEmail string
}

func NewEmailNotifierFromEnv() EmailNotifier {
	apiKey := strings.TrimSpace(os.Getenv("RESEND_API_KEY"))
	if apiKey == "" {
		return &noopEmailNotifier{}
	}

	fromEmail := strings.TrimSpace(os.Getenv("RESEND_FROM_EMAIL"))
	if fromEmail == "" {
		fromEmail = "onboarding@resend.dev"
	}

	return &resendEmailNotifier{
		client:    resend.NewClient(apiKey),
		fromEmail: fromEmail,
	}
}

func (n *resendEmailNotifier) SendPackPurchaseQRCode(_ context.Context, payload PackPurchaseEmailPayload) error {
	toEmail := strings.TrimSpace(payload.ToEmail)
	if toEmail == "" {
		return nil
	}
	if strings.TrimSpace(payload.QRCode) == "" {
		return nil
	}

	subject := fmt.Sprintf("Your eSIM QR Code for %s", firstNonEmptyValue(payload.PackName, payload.CatalogID, "Purchased Pack"))
	htmlBody, err := renderPackPurchaseEmailTemplate(payload)
	if err != nil {
		return err
	}

	params := &resend.SendEmailRequest{
		From:    n.fromEmail,
		To:      []string{toEmail},
		Subject: subject,
		Html:    htmlBody,
	}

	_, err = n.client.Emails.Send(params)
	return err
}

func renderPackPurchaseEmailTemplate(payload PackPurchaseEmailPayload) (string, error) {
	const tmpl = `
<!doctype html>
<html>
  <body style="font-family: Arial, sans-serif; color: #111827; line-height: 1.5;">
    <h2 style="margin: 0 0 16px;">Your Pack Has Been Purchased Successfully</h2>
    <p>Hello {{.ReceiverID}},</p>
    <p>Your pack purchase is successful. Use the QR code below to install your eSIM.</p>
    <table cellpadding="6" cellspacing="0" style="margin: 12px 0; border-collapse: collapse;">
      <tr><td><strong>Order ID:</strong></td><td>{{.OrderID}}</td></tr>
      <tr><td><strong>Tenant ID:</strong></td><td>{{.TenantID}}</td></tr>
      <tr><td><strong>Catalog ID:</strong></td><td>{{.CatalogID}}</td></tr>
      <tr><td><strong>Pack Name:</strong></td><td>{{.PackName}}</td></tr>
      <tr><td><strong>ICCID:</strong></td><td>{{.ICCID}}</td></tr>
    </table>
    <p><strong>eSIM QR Code</strong></p>
    <p><img src="{{.QRCode}}" alt="eSIM QR Code" style="max-width: 260px; border: 1px solid #e5e7eb; padding: 8px;" /></p>
    <p>If the image does not render in your email client, copy this QR URL:</p>
    <p><a href="{{.QRCode}}">{{.QRCode}}</a></p>
    <p>Regards,<br/>Circles B2B Team</p>
  </body>
</html>`

	t, err := template.New("pack_purchase_email").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	if err := t.Execute(&builder, payload); err != nil {
		return "", err
	}
	return builder.String(), nil
}

func firstNonEmptyValue(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
