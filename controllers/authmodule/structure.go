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
}
