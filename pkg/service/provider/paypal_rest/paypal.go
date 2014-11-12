// Paypal data types
package paypal_rest

type PayPalPaymentMethod string

const (
	PayPalPaymentMethodPayPal PayPalPaymentMethod = "paypal"
	PayPalPaymentMethodCC                         = "credit_card"
)

type PayPalError struct {
	Name            string `json:"name"`
	Message         string `json:"message"`
	InformationLink string `json:"information_link"`
	Details         string `json:"details"`
}

// PaypalPayer represents the "payer" object as defined by the PayPal REST-API
//
// See https://developer.paypal.com/docs/api/#payer-object
type PaypalPayer struct {
	PaymentMethod PayPalPaymentMethod `json:"payment_method"`
	Status        string              `json:"status,omitempty"`
}

type PayPalPayerInfo struct {
	Email           string                `json:"email,omitempty"`
	FirstName       string                `json:"first_name,omitempty"`
	LastName        string                `json:"last_name,omitempty"`
	PayerID         string                `json:"payer_id,omitempty"`
	Phone           string                `json:"phone,omitempty"`
	ShippingAddress PayPalShippingAddress `json:"shipping_address,omitempty"`
	TaxIDType       string                `json:"tax_id_type,omitempty"`
	TaxID           string                `json:"tax_id,omitempty"`
}

type PayPalShippingAddress struct {
	RecipientName string `json:"recipient_name,omitempty"`
	Type          string `json:"type,omitempty"`
	Line1         string `json:"line1"`
	Line2         string `json:"line2,omitempty"`
	City          string `json:"city"`
	CountryCode   string `json:"country_code"`
	PostalCode    string `json:"postal_code,omitempty"`
	State         string `json:"state,omitempty"`
	Phone         string `json:"phone,omitempty"`
}

// PayPalDetails represents the PayPal "amount" details type
//
// See https://developer.paypal.com/docs/api/#details-object
type PayPalDetails struct {
	Shipping         string `json:"shipping,omitempty"`
	Subtotal         string `json:"subtotal,omitempty"`
	Tax              string `json:"tax,omitempty"`
	Fee              string `json:"fee,omitempty"`
	HandlingFee      string `json:"handling_fee,omitempty"`
	Insurance        string `json:"insurance,omitempty"`
	ShippingDiscount string `json:shipping_discount,omitempty`
}

type PayPalAmount struct {
	Currency string         `json:"currency"`
	Total    string         `json:"total"`
	Details  *PayPalDetails `json:"details,omitempty"`
}

type PayPalTransaction struct {
	Amount         PayPalAmount `json:"amount"`
	Description    string       `json:"description,omitempty"`
	InvoiceNumber  string       `json:"invoice_number,omitempty"`
	Custom         string       `json:"custom,omitempty"`
	SoftDescriptor string       `json:"soft_descriptor,omitempty"`
}

type PayPalRedirectURLs struct {
	ReturnURL string `json:"return_url"`
	CancelURL string `json:"cancel_url"`
}

type PayPalLink struct {
	HRef   string `json:"href"`
	Rel    string `json:"rel"`
	Method string `json:"method"`
}

type PayPalPaymentRequest struct {
	Intent       string              `json:"intent"`
	Payer        PaypalPayer         `json:"payer"`
	Transactions []PayPalTransaction `json:"transactions"`
	RedirectURLs PayPalRedirectURLs  `json:"redirect_urls,omitempty"`
}

type PaypalPayment struct {
	PayPalPaymentRequest

	ID         string       `json:"id"`
	CreateTime string       `json:"create_time"`
	State      string       `json:"state"`
	UpdateTime string       `json:"update_time"`
	Links      []PayPalLink `json:"links"`
}