package stripe

import (
	"log"
	"net/http"
	"os"
	"time"

	"blogr.moe/backend/auth"
	"blogr.moe/backend/database"
	"github.com/labstack/echo/v4"
	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/checkout/session"
	"github.com/stripe/stripe-go/v79/customer"
	"go.mongodb.org/mongo-driver/bson"
)

func CheckoutSuccessHandler(c echo.Context) error {
	stripe.Key = os.Getenv("STRIPE_SECRET")
	checkoutID := c.QueryParam("checkout_id")
	if checkoutID == "" {
		return c.JSON(400, map[string]string{"error": "Invalid checkout ID"})
	}

	sess, err := session.Get(checkoutID, nil)
	if err != nil {
		return c.JSON(500, map[string]string{"error": "Error retrieving session"})
	}

	if sess.PaymentStatus != stripe.CheckoutSessionPaymentStatusPaid {
		return c.JSON(400, map[string]string{"error": "Payment not completed"})
	}

	cust, err := customer.Get(sess.Customer.ID, nil)
	if err != nil {
		return c.JSON(500, map[string]string{"error": "Error fetching customer details"})
	}

	if cust.Email == "" {
		return c.JSON(400, map[string]string{"error": "Invalid email"})
	}

	user, _ := auth.GetUserByEmail(cust.Email)

	user.Premium = true
	user.PremiumExpiry = time.Now().AddDate(0, 1, 0).Format(time.RFC3339)

	filter := bson.M{"uuid": user.UUID}
	update := bson.M{"$set": user}

	_, err = database.DB_Users.Collection("users").UpdateOne(c.Request().Context(), filter, update)
	if err != nil {
		log.Println("Error updating user:", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
	}
	return c.Redirect(301, os.Getenv("FRONTEND_URL")+"/")
}

func GetCheckoutSession(c echo.Context) error {
	stripe.Key = os.Getenv("STRIPE_SECRET")
	user := auth.GetUserFromContext(c)
	if user.Email == "" {
		return c.JSON(400, map[string]string{"error": "Invalid user"})
	}

	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(os.Getenv("STRIPE_SUCCESS_URL")),
		CancelURL:  stripe.String(os.Getenv("STRIPE_CANCEL_URL")),
		Customer:   stripe.String(user.Email),
	}

	sess, err := session.New(params)
	if err != nil {
		return c.JSON(500, map[string]string{"error": "Error creating session"})
	}

	return c.JSON(200, map[string]string{"id": sess.ID})
}
