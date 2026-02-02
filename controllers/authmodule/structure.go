package authmodule

import "time"

// AuthModuleService is the RPC service.
// Example RPC method string: "AuthModuleService.RequestTotpQrCode".
type AuthModuleService struct{}

type RequestTotpQrCodeRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	// Optional: gateway can pass domain; otherwise DEFAULT_DOMAIN is used.
	Domain string `json:"domain,omitempty"`
}

// EnableTfaRequest enables 2FA for a customer user.
// RPC method string: "AuthModuleService.EnableTfa".
type EnableTfaRequest struct {
	Username         string `json:"username"`
	Password         string `json:"password"`
	TfaType          string `json:"tfaType"` // "SMS" | "AuthenticatorApp"
	TfaCode          string `json:"tfaCode"`
	PhoneNumber      string `json:"phoneNumber,omitempty"`      // required if SMS
	TotpSharedSecret string `json:"totpSharedSecret,omitempty"` // required if AuthenticatorApp

	// Gateway will pass this (like .NET helpers).
	Domain string `json:"domain"`
}

// DbResultRPC is a gob-safe result wrapper (Details/Errors are raw JSON strings).
// Some clients decode into `ID`, others into `Id`, so we send both.
type DbResultRPC struct {
	ID     int    `json:"id"`
	Id     int    `json:"Id"`
	Status string `json:"status"`

	Details string `json:"details"` // raw JSON string
	Errors  string `json:"errors"`  // raw JSON string or empty
}

// ErrorResultPublic is used when we encode validation errors as JSON strings.
type ErrorResultPublic struct {
	ErrorType   string `json:"errorType"`
	FieldName   string `json:"fieldName"`
	MessageCode string `json:"messageCode"`
}

type TotpQrCodeDetails struct {
	QrCode       string `json:"qrCode"`
	SharedSecret string `json:"sharedSecret"`
}

type RequestTotpQrCodeResponse struct {
	ID      int    `json:"id"`
	Id      int    `json:"Id"` // legacy (.NET style)
	Status  string `json:"status"`
	Details string `json:"details"` // JSON string for gob compatibility
	// NOTE: keep as string for gob compatibility (no interface{}).
	Errors string `json:"errors"`
}

type ErrorResult struct {
	ErrorType   string `json:"-"`
	FieldName   string `json:"fieldName"`
	MessageCode string `json:"messageCode"`
}

type SiteUsersAuthData struct {
	SiteUsersId  int
	SiteName     string
	PasswordHash string

	BSuppressed         bool
	TryLoginCount       int
	DateLastFailedLogin *time.Time

	// TFA state
	TwoFactorSMSAuthEnabled bool
	TwoFactorAppAuthEnabled bool

	// SMS code info (if SMS flow is used)
	TfaCode       string
	TfaCodeExpiry *time.Time

	// stored secret (some systems store it after enable)
	TotpSharedSecret string

	// Email verified flag etc (for login flow)
	BEmailVerified bool
}
