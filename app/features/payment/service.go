package payment

import (
	"context"
	"errors"
	"fmt"

	"carf-chorf.com/app/app/domaine"
	"carf-chorf.com/app/app/services/discord_service"
	"carf-chorf.com/app/app/services/mail_service"
	"carf-chorf.com/app/app/services/paypal_service"
	"carf-chorf.com/app/app/services/report_service"
	"carf-chorf.com/app/app/services/transaction_service"
	"carf-chorf.com/app/pkg/money"
	"carf-chorf.com/app/pkg/simple_logger"
	"github.com/google/uuid"
)

var (
	ErrTransactionNotFound  = errors.New("transaction not found")
	ErrPaymentNotStarted    = errors.New("payment has not been started")
	ErrPaymentNotCompleted  = errors.New("payment was not completed")
	ErrPaymentAmountInvalid = errors.New("payment amount does not match the transaction")
)

// CreateOrder creates a PayPal order and stores its id on the transaction.
func CreatePayPalOrder(ctx context.Context, token string) (string, error) {
	transaction, err := transaction_service.FindByToken(token)
	if err != nil {
		return "", err
	}
	if transaction == nil {
		return "", ErrTransactionNotFound
	}
	if transaction.Status == domaine.TransactionPaid {
		return "", transaction_service.ErrTransactionNotPending
	}

	pid, err := paypal_service.CreateOrder(ctx, transaction)
	if err != nil {
		simple_logger.Errorf("payment: create paypal order: %v", err)
		return "", err
	}
	if err := transaction_service.SetPid(transaction.Token, pid); err != nil {
		return "", err
	}
	return pid, nil
}

// Capture captures the PayPal order, completes the purchase, and returns the
// page the browser should visit next.
func CapturePayPalOrder(ctx context.Context, token string) (string, error) {
	transaction, err := transaction_service.FindByToken(token)
	if err != nil {
		return "", err
	}
	if transaction == nil {
		return "", ErrTransactionNotFound
	}

	reportAccessToken := ""
	if transaction.Status != domaine.TransactionPaid {
		if transaction.Pid == "" {
			return "", ErrPaymentNotStarted
		}

		if transaction.CacheRecord != nil {
			reportAccessToken = uuid.NewString()
			reportHTML, err := report_service.Prepare(ctx, transaction.CacheRecord.Vin)
			if err != nil {
				simple_logger.Errorf("payment: prepare report: %v", err)
				message := fmt.Sprintf(
					"🆘 **Report preparation failed**\nVIN: `%s`\nTransaction: `%s`\nUser: `%s`\nError: %v",
					transaction.CacheRecord.Vin,
					transaction.Token,
					transaction.User.Email,
					err,
				)
				go func() {
					if notifyErr := discord_service.SendSOSNotification(message); notifyErr != nil {
						simple_logger.Errorf("payment: send report SOS: %v", notifyErr)
					}
				}()
				return "", fmt.Errorf("could not prepare the vehicle report")
			}
			if err := report_service.Save(reportAccessToken, reportHTML); err != nil {
				_ = report_service.Delete(reportAccessToken)
				simple_logger.Errorf("payment: save report before capture: %v", err)
				return "", fmt.Errorf("could not save the vehicle report")
			}
			reportHTML = nil
		}

		paypalOrder, err := paypal_service.CaptureOrder(ctx, transaction.Pid)
		if err != nil {
			// could not capture maybe has been captured before
			paypalOrder, err = paypal_service.GetOrder(ctx, transaction.Pid)
			if err != nil || !paypalOrder.IsPaid() {
				cleanupReport(reportAccessToken)
				return "", fmt.Errorf("could not complete payment")
			}
		}
		if err := verifyCapturedOrder(paypalOrder, transaction); err != nil {
			cleanupReport(reportAccessToken)
			return "", err
		}
	}

	transaction, order, orderCreated, err := transaction_service.CompletePurchase(transaction.Token, transaction.Pid, reportAccessToken)
	if err != nil {
		cleanupReport(reportAccessToken)
		simple_logger.Errorf("payment: complete purchase: %v", err)
		return "", err
	}

	if transaction.CacheRecord != nil {

		if order == nil || order.AccessToken == "" {
			cleanupReport(reportAccessToken)
			return "", errors.New("report order was not created")
		}
		if reportAccessToken != "" && order.AccessToken != reportAccessToken {
			cleanupReport(reportAccessToken)
		}

		if !report_service.Has(order.AccessToken) {
			if _, err := report_service.Ensure(ctx, order); err != nil {
				simple_logger.Errorf("payment: ensure report after purchase: %v", err)
			}
		}

		if orderCreated {
			sendReportEmail(transaction.User.Email, order.AccessToken, transaction.CacheRecord.Vin, transaction.CacheRecord.Vehicle)
		}
		return "/reports/view?access_token=" + order.AccessToken, nil
	}

	return "/payment/success?token=" + transaction.Token, nil
}

func cleanupReport(accessToken string) {
	if accessToken == "" {
		return
	}
	if err := report_service.Delete(accessToken); err != nil {
		simple_logger.Errorf("payment: cleanup report: %v", err)
	}
}

func verifyCapturedOrder(order *paypal_service.Order, transaction *domaine.Transaction) error {
	if !order.IsPaid() {
		return ErrPaymentNotCompleted
	}
	_, value, ok := order.CapturedAmount()
	if !ok || value != money.Decimal(transaction.Amount) {
		return ErrPaymentAmountInvalid
	}
	return nil
}

func sendReportEmail(to, accessToken, vin, vehicle string) {
	go func() {
		if err := mail_service.SendReport(to, report_service.URL(accessToken), vin, vehicle); err != nil {
			simple_logger.Errorf("payment: send report email: %v", err)
		}
	}()
}
