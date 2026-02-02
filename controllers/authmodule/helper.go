package authmodule

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"go-cloud-customer/db"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/skip2/go-qrcode"
	"golang.org/x/crypto/pbkdf2"
)

const (
	defaultTryLoginCounterMax  = 5
	defaultLoginFailedLockMins = 10
)

func parseTimePtr(s string) *time.Time {
	s = strings.TrimSpace(s)
	if s == "" || strings.EqualFold(s, "null") {
		return nil
	}
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05.9999999",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05.9999999",
		"2006-01-02T15:04:05",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			u := t.UTC()
			return &u
		}
	}
	return nil
}

func ValidateCredentials(req RequestTotpQrCodeRequest) []ErrorResult {
	var errs []ErrorResult
	if strings.TrimSpace(req.Username) == "" {
		errs = append(errs, ErrBadRequest("Username", "Required"))
	}
	if strings.TrimSpace(req.Password) == "" {
		errs = append(errs, ErrBadRequest("Password", "Required"))
	}
	return DedupErrors(errs)
}

func GetDomainOrDefault(domain string) string {
	domain = strings.TrimSpace(domain)
	if domain != "" {
		return domain
	}
	v := strings.TrimSpace(os.Getenv("DEFAULT_DOMAIN"))
	if v == "" {
		return "example.com"
	}
	return v
}

// ---------- DB (matches AuthModuleDbClient.GetSiteUsersAuthDataAsync) ----------
func GetSiteUsersAuthData(ctx context.Context, username, accountType, domain string) (*SiteUsersAuthData, error) {
	if db.DB == nil {
		return nil, errors.New("DB is nil")
	}

	sp := "v1_PublicRole_AuthModule_GetSiteUsersAuthData"
	q := "EXEC " + sp + " @Username=@Username, @AccountType=@AccountType, @Domain=@Domain"

	rows, err := db.DB.QueryContext(ctx, q,
		sql.Named("Username", username),
		sql.Named("AccountType", accountType),
		sql.Named("Domain", domain),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, nil
	}

	raw := make([]any, len(cols))
	dest := make([]any, len(cols))
	for i := range raw {
		dest[i] = &raw[i]
	}
	if err := rows.Scan(dest...); err != nil {
		return nil, err
	}

	row := map[string]any{}
	for i, c := range cols {
		row[strings.ToLower(strings.TrimSpace(c))] = raw[i]
	}

	getString := func(keys ...string) string {
		for _, k := range keys {
			v := row[strings.ToLower(k)]
			if v == nil {
				continue
			}
			switch t := v.(type) {
			case string:
				return t
			case []byte:
				return string(t)
			default:
				return fmt.Sprint(t)
			}
		}
		return ""
	}

	getInt := func(keys ...string) int {
		for _, k := range keys {
			v := row[strings.ToLower(k)]
			if v == nil {
				continue
			}
			switch t := v.(type) {
			case int:
				return t
			case int32:
				return int(t)
			case int64:
				return int(t)
			case float64:
				return int(t)
			case []byte:
				n, _ := strconv.Atoi(strings.TrimSpace(string(t)))
				return n
			case string:
				n, _ := strconv.Atoi(strings.TrimSpace(t))
				return n
			default:
				n, _ := strconv.Atoi(strings.TrimSpace(fmt.Sprint(t)))
				return n
			}
		}
		return 0
	}

	getBool := func(keys ...string) bool {
		for _, k := range keys {
			v := row[strings.ToLower(k)]
			if v == nil {
				continue
			}
			switch t := v.(type) {
			case bool:
				return t
			case int64:
				return t != 0
			case int32:
				return t != 0
			case int:
				return t != 0
			case []byte:
				s := strings.TrimSpace(string(t))
				return s == "1" || strings.EqualFold(s, "true")
			case string:
				s := strings.TrimSpace(t)
				return s == "1" || strings.EqualFold(s, "true")
			}
		}
		return false
	}

	getTimePtr := func(keys ...string) *time.Time {
		for _, k := range keys {
			v := row[strings.ToLower(k)]
			if v == nil {
				continue
			}
			switch t := v.(type) {
			case time.Time:
				u := t.UTC()
				return &u
			case *time.Time:
				if t == nil {
					return nil
				}
				u := t.UTC()
				return &u
			case []byte:
				return parseTimePtr(string(t))
			case string:
				return parseTimePtr(t)
			default:
				return parseTimePtr(fmt.Sprint(t))
			}
		}
		return nil
	}

	out := &SiteUsersAuthData{
		SiteUsersId:         getInt("SiteUsersId", "siteUsersId", "SiteUserId", "siteUserId", "Id", "ID"),
		SiteName:            getString("SiteName", "siteName"),
		PasswordHash:        getString("PasswordHash", "passwordHash"),
		BSuppressed:         getBool("bSuppressed", "BSuppressed", "suppressed"),
		TryLoginCount:       getInt("TryLoginCount", "tryLoginCount"),
		DateLastFailedLogin: getTimePtr("DateLastFailedLogin", "dateLastFailedLogin", "DateLastFailed", "dateLastFailed"),
	}
	return out, nil
}

// ---------- Validation rules (matches .NET ValidateUsernameAndPassword) ----------

type LoginStatus string

const (
	LoginStatusSuccess           LoginStatus = "Success"
	LoginStatusUserNotFound      LoginStatus = "UserNotFound"
	LoginStatusAccountSuppressed LoginStatus = "AccountSuppressed"
	LoginStatusAccountLocked     LoginStatus = "AccountLocked"
	LoginStatusPasswordInvalid   LoginStatus = "PasswordInvalid"
)

func ValidateUsernameAndPassword(password string, authData *SiteUsersAuthData) LoginStatus {
	if authData == nil {
		return LoginStatusUserNotFound
	}

	if authData.BSuppressed {
		return LoginStatusAccountSuppressed
	}

	tryMax := GetTryLoginCounterMax()
	lockMinutes := GetLoginFailedLockMinutes()
	if authData.TryLoginCount > tryMax && authData.DateLastFailedLogin != nil {
		if time.Since(authData.DateLastFailedLogin.UTC()).Minutes() <= float64(lockMinutes) {
			return LoginStatusAccountLocked
		}
	}

	if !VerifyPasswordPBKDF2(password, authData.PasswordHash) {
		return LoginStatusPasswordInvalid
	}

	return LoginStatusSuccess
}

func GetTryLoginCounterMax() int {
	v := strings.TrimSpace(os.Getenv("AUTH_TRY_LOGIN_COUNTER_MAX"))
	if v == "" {
		return defaultTryLoginCounterMax
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return defaultTryLoginCounterMax
	}
	return n
}

func GetLoginFailedLockMinutes() int {
	v := strings.TrimSpace(os.Getenv("AUTH_LOGIN_FAILED_LOCK_MINUTES"))
	if v == "" {
		return defaultLoginFailedLockMins
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return defaultLoginFailedLockMins
	}
	return n
}

func GetUsernameAndPasswordValidationFailureErrors(status LoginStatus) []ErrorResult {
	msg := "Username_Or_Password_Incorrect"
	switch status {
	case LoginStatusAccountSuppressed:
		msg = "Account_Suppressed"
	case LoginStatusAccountLocked:
		msg = "Account_Locked"
	}

	return []ErrorResult{
		{ErrorType: "BadRequest", FieldName: "Username", MessageCode: msg},
		{ErrorType: "BadRequest", FieldName: "Password", MessageCode: msg},
	}
}

// ---------- Password verify (matches .NET PasswordHelper.VerifyPassword) ----------

func VerifyPasswordPBKDF2(passwordToVerify, savedPasswordHash string) bool {
	parts := strings.Split(savedPasswordHash, ".")
	if len(parts) != 3 {
		return false
	}

	iterations, err := strconv.Atoi(parts[0])
	if err != nil || iterations <= 0 {
		return false
	}

	salt, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}

	origKeyB64 := parts[2]
	key := pbkdf2.Key([]byte(passwordToVerify), salt, iterations, 32, sha1.New)
	keyB64 := base64.StdEncoding.EncodeToString(key)

	return keyB64 == origKeyB64
}

// ---------- TOTP shared secret + QR (matches .NET GenerateTotpSharedSecretAndQrCode) ----------

func GetTotpLabelSuffix() string {
	v := strings.TrimSpace(os.Getenv("AUTH_TOTP_LABEL"))
	if v == "" {
		return "-(Staging)"
	}
	return v
}

func GenerateTotpSharedSecretAndQrCode(totpLabel string) (qrCodeBase64Png string, totpSharedSecret string) {
	base32Chars := []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ234567")
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	var sb strings.Builder
	sb.Grow(16)
	for i := 0; i < 16; i++ {
		idx := r.Intn(len(base32Chars) - 1) // match .NET behavior (exclude last)
		sb.WriteByte(base32Chars[idx])
	}
	secret := sb.String()

	qrString := "otpauth://totp/" + totpLabel + "?secret=" + secret

	pngBytes, err := qrcode.Encode(qrString, qrcode.High, 400)
	if err != nil {
		// fallback: no QR, but still return secret
		return "", secret
	}

	return base64.StdEncoding.EncodeToString(pngBytes), secret
}

// ---------- Errors / Formatting ----------

func ErrBadRequest(field, msg string) ErrorResult {
	return ErrorResult{ErrorType: "BadRequest", FieldName: field, MessageCode: msg}
}

func ErrInternal(field, msg string) ErrorResult {
	return ErrorResult{ErrorType: "InternalServerError", FieldName: field, MessageCode: msg}
}

func DedupErrors(in []ErrorResult) []ErrorResult {
	seen := map[string]struct{}{}
	out := make([]ErrorResult, 0, len(in))
	for _, e := range in {
		k := e.ErrorType + "|" + e.FieldName + "|" + e.MessageCode
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, e)
	}
	return out
}

func FormatDbErrors(errs []ErrorResult) string {
	if len(errs) == 0 {
		return ""
	}
	parts := make([]string, 0, len(errs))
	for _, e := range errs {
		parts = append(parts, e.ErrorType+"]_["+e.FieldName+"]_["+e.MessageCode)
	}
	return strings.Join(parts, "]|[ ")
}

func logAuthFailure(action string, err error) {
	if err == nil {
		return
	}
	log.Printf("AuthModule %s failed: %v", action, err)
}

func safeLabel(siteName string) string {
	s := strings.TrimSpace(siteName)
	if s == "" {
		return "MrRetailer"
	}
	// Keep label as-is; avoid URL-encoding to match existing expectation.
	return s
}

func BuildTotpLabel(siteName string) string {
	return safeLabel(siteName) + GetTotpLabelSuffix()
}

func NewUnexpectedErrorResponse() RequestTotpQrCodeResponse {
	return RequestTotpQrCodeResponse{
		ID:      0,
		Id:      0,
		Status:  "0",
		Details: "",
		Errors:  FormatDbErrors([]ErrorResult{ErrInternal("", "Unexpected_Error")}),
	}
}

func NewBadRequestInvalidJSONResponse() RequestTotpQrCodeResponse {
	return RequestTotpQrCodeResponse{
		ID:      0,
		Id:      0,
		Status:  "0",
		Details: "",
		Errors:  FormatDbErrors([]ErrorResult{ErrBadRequest("", "Invalid_JSON")}),
	}
}

func NewValidationErrorResponse(errs []ErrorResult) RequestTotpQrCodeResponse {
	return RequestTotpQrCodeResponse{
		ID:      0,
		Id:      0,
		Status:  "0",
		Details: "",
		Errors:  FormatDbErrors(errs),
	}
}

func RequestTotpQrCodeService(ctx context.Context, req RequestTotpQrCodeRequest, domain string) RequestTotpQrCodeResponse {
	authData, err := GetSiteUsersAuthData(ctx, req.Username, "Customer", domain)
	if err != nil {
		logAuthFailure("GetSiteUsersAuthData", err)
		return NewUnexpectedErrorResponse()
	}

	status := ValidateUsernameAndPassword(req.Password, authData)
	if status == LoginStatusSuccess || status == LoginStatusAccountSuppressed {
		label := BuildTotpLabel("")
		if authData != nil {
			label = BuildTotpLabel(authData.SiteName)
		}

		qrB64, secret := GenerateTotpSharedSecretAndQrCode(label)
		details := TotpQrCodeDetails{QrCode: qrB64, SharedSecret: secret}
		b, err := json.Marshal(details)
		if err != nil {
			logAuthFailure("MarshalTotpQrDetails", err)
			return NewUnexpectedErrorResponse()
		}

		resp := RequestTotpQrCodeResponse{
			ID:      0,
			Id:      0,
			Status:  "1",
			Details: string(b),
			Errors:  "",
		}
		if authData != nil {
			resp.ID = authData.SiteUsersId
			resp.Id = resp.ID
		}
		return resp
	}

	errList := GetUsernameAndPasswordValidationFailureErrors(status)
	return NewValidationErrorResponse(errList)
}
