package customer

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"go-cloud-customer/constants"
	"go-cloud-customer/db"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/pbkdf2"
)

type ErrorResultFile struct {
	ErrorType   string `json:"errorType"`
	FieldName   string `json:"fieldName"`
	MessageCode string `json:"messageCode"`
}

func EmptyErrors() []ErrorResultFile {
	return []ErrorResultFile{}
}

func OneError(errorType, field, code string) []ErrorResultFile {
	return []ErrorResultFile{{ErrorType: errorType, FieldName: field, MessageCode: code}}
}
func errorsJSON(errs any) string {
	b, _ := json.Marshal(errs)
	return string(b)
}

func ensureHTTP(u string) string {
	u = strings.TrimSpace(u)
	if u == "" {
		return u
	}
	if strings.HasPrefix(u, "https://") || strings.HasPrefix(u, "http://") {
		return u
	}
	return "https://" + u
}

func buildJumioLink(baseURL string, token string) string {
	baseURL = ensureHTTP(baseURL)
	if strings.Contains(baseURL, "?") {
		return baseURL + "&token=" + url.QueryEscape(token)
	}
	return strings.TrimRight(baseURL, "/") + "/idverification?token=" + url.QueryEscape(token)
}

// buildGroupsFromDocumentList: detailsStr me se listData nikalo, phir group banao
func buildGroupsFromDocumentList(detailsStr string) (string, bool) {
	var root any
	if json.Unmarshal([]byte(detailsStr), &root) != nil {
		return "", false
	}

	// Expecting: { "listData": [ ... ] } or { "details": { "listData": [...] } } types
	list := extractListData(root)
	if list == nil {
		return "", false
	}

	type Group struct {
		GroupKey string                   `json:"groupKey"`
		Items    []map[string]interface{} `json:"items"`
	}

	groups := map[string][]map[string]interface{}{}

	for _, it := range list {
		// possible keys (aapke SP columns pe depend)
		key := pickGroupKey(it)
		groups[key] = append(groups[key], it)
	}

	out := make([]Group, 0, len(groups))
	for k, items := range groups {
		out = append(out, Group{GroupKey: k, Items: items})
	}

	b, _ := json.Marshal(map[string]interface{}{
		"groups": out,
	})
	return string(b), true
}

func extractListData(root any) []map[string]interface{} {
	// case1: root["listData"]
	if m, ok := root.(map[string]interface{}); ok {
		if v, ok := m["listData"]; ok {
			return castList(v)
		}
		if d, ok := m["details"]; ok {
			if md, ok := d.(map[string]interface{}); ok {
				if v, ok := md["listData"]; ok {
					return castList(v)
				}
			}
		}
	}
	return nil
}

func castList(v any) []map[string]interface{} {
	arr, ok := v.([]interface{})
	if !ok {
		return nil
	}
	out := make([]map[string]interface{}, 0, len(arr))
	for _, x := range arr {
		if m, ok := x.(map[string]interface{}); ok {
			out = append(out, m)
		}
	}
	return out
}

func pickGroupKey(item map[string]interface{}) string {
	// try common keys
	if v, ok := item["documentGroupName"]; ok {
		if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
			return s
		}
	}
	if v, ok := item["documentGroupId"]; ok {
		return fmt.Sprintf("group_%v", v)
	}
	// fallback
	return "default"
}

func businessKycDocPath(customersId int64) string {
	return fmt.Sprintf("BusinessKYCDocuments/%d/BusinessDocuments/", customersId)
}

// ---- Storage stubs for KYC ----
// Replace with your real implementations.
func StorageSaveOrOverwrite(path, contentType string, fileBytes []byte, container string) (bool, error) {
	// TODO: Implement actual storage logic
	return true, nil
}

func StorageDeleteIfExists(fullPath, container string) error {
	// TODO: Implement actual storage logic
	return nil
}

func AzureBlobSecureRootContainerName() string {
	return "secure-container"
}

func businessKycDocName(customersId, siteUsersId int64, originalFileName string) string {
	return fmt.Sprintf("%d_%d_%s", customersId, siteUsersId, originalFileName)
}

func validateAddKycDocument(req *AddKYCDocumentRequest) []ErrorResult {
	var errs []ErrorResult
	if req == nil {
		return []ErrorResult{{ErrorType: "BadRequest", FieldName: "Json", MessageCode: "Invalid_Json"}}
	}
	if req.DocumentId <= 0 {
		errs = append(errs, ErrorResult{"BadRequest", "DocumentId", "Required"})
	}
	if req.CustomersBusinessDocumentsId == nil {
		if len(req.FileBytes) == 0 {
			errs = append(errs, ErrorResult{"BadRequest", "File", "Required"})
		}
	}
	return errs
}

func validateFileTypeAndExt(contentType, fileName string) *ErrorResult {
	allowedTypes := map[string]bool{
		"application/pdf": true,
		"image/png":       true,
		"image/jpeg":      true,
	}
	if !allowedTypes[strings.TrimSpace(strings.ToLower(contentType))] {
		e := ErrorResult{"BadRequest", "File", "Invalid_Format"}
		return &e
	}

	ext := strings.ToLower(filepath.Ext(fileName))
	allowedExt := map[string]bool{".pdf": true, ".png": true, ".jpeg": true, ".jpg": true}
	if !allowedExt[ext] {
		e := ErrorResult{"BadRequest", "File", "Invalid_Extension"}
		return &e
	}
	return nil
}

func toErrorsJSON(errs []ErrorResult) string {
	b, _ := json.Marshal(errs)
	return string(b)
}

func withTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}

// -------------------- Validation (match .NET attributes) --------------------

// parseLegacyDbErrorString parses legacy DB error strings like:
// "Unauthenticated]_[CustomersID]_[User_Not_Found ]|[ Unauthenticated]_[SiteUsersID]_[User_Not_Found ]|[ ..."
// into []ErrorResultFile objects.
func parseLegacyDbErrorString(s string) []ErrorResultFile {
	var results []ErrorResultFile
	s = strings.TrimSpace(s)
	if s == "" {
		return results
	}
	// Split by |[ or | [
	parts := strings.Split(s, "|[")
	for _, part := range parts {
		part = strings.Trim(part, " []|\t\n\r")
		if part == "" {
			continue
		}
		// Split by ]_[
		fields := strings.Split(part, "]_[")
		// Defensive: pad to 3 fields
		for len(fields) < 3 {
			fields = append(fields, "")
		}
		errType := strings.TrimSpace(fields[0])
		fieldName := strings.TrimSpace(fields[1])
		msgCode := strings.TrimSpace(fields[2])
		results = append(results, ErrorResultFile{
			ErrorType:   errType,
			FieldName:   fieldName,
			MessageCode: msgCode,
		})
	}
	return results
}

var (
	phoneRegexp = regexp.MustCompile(`^(\+?\d+)?\s?(\(\d+\))?\s?(\d+-\d+)?$`)
)

func ValidateRegisterCustomerUser(req CustomerUserRegistrationRequest) []ErrorResult {
	var errs []ErrorResult

	if strings.TrimSpace(req.FirstName) == "" {
		errs = append(errs, ErrBadRequest("FirstName", "Required"))
	}
	if strings.TrimSpace(req.LastName) == "" {
		errs = append(errs, ErrBadRequest("LastName", "Required"))
	}
	if strings.TrimSpace(req.CurrencyCode) == "" {
		errs = append(errs, ErrBadRequest("CurrencyCode", "Required"))
	}
	if strings.TrimSpace(req.EmailAddress) == "" {
		errs = append(errs, ErrBadRequest("EmailAddress", "Required"))
	} else if !IsValidEmail(req.EmailAddress) {
		errs = append(errs, ErrBadRequest("EmailAddress", "Invalid"))
	}

	if strings.TrimSpace(req.Password) == "" {
		errs = append(errs, ErrBadRequest("Password", "Required"))
	} else {
		if len(req.Password) < 8 || len(req.Password) > 20 {
			errs = append(errs, ErrBadRequest("Password", "Invalid_Password_Length"))
		}
		if !PasswordPolicyMatch(req.Password) {
			errs = append(errs, ErrBadRequest("Password", "Password_Policy_No_Match"))
		}
	}

	if strings.TrimSpace(req.ConfirmPassword) == "" {
		errs = append(errs, ErrBadRequest("ConfirmPassword", "Required"))
	}

	if strings.TrimSpace(req.Password) != "" && strings.TrimSpace(req.ConfirmPassword) != "" {
		if req.Password != req.ConfirmPassword {
			errs = append(errs, ErrBadRequest("Password", "Password_No_Match"))
		}
	}

	if strings.TrimSpace(req.RegisteredCountryCode) == "" {
		errs = append(errs, ErrBadRequest("RegisteredCountryCode", "Required"))
	}

	if strings.TrimSpace(req.PhoneNumber) == "" {
		errs = append(errs, ErrBadRequest("PhoneNumber", "Required"))
	} else {
		if len(req.PhoneNumber) < 6 || len(req.PhoneNumber) > 50 || !phoneRegexp.MatchString(req.PhoneNumber) {
			errs = append(errs, ErrBadRequest("PhoneNumber", "Invalid_PhoneNumber"))
		}
	}

	// bConductsThirdPartyPayments comes as a bool from the gateway; false is a valid value.

	if req.RequestedProducts == nil || len(req.RequestedProducts) == 0 {
		errs = append(errs, ErrBadRequest("RequestedProducts", "Required"))
	}

	return DedupErrors(errs)
}

func PasswordPolicyMatch(pw string) bool {
	if pw == "" || len(pw) < 8 {
		return false
	}
	hasLetter := false
	hasNumber := false
	for _, c := range pw {
		if c >= '0' && c <= '9' {
			hasNumber = true
		} else if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') {
			hasLetter = true
		}
	}
	return hasLetter && hasNumber
}

func IsValidEmail(s string) bool {
	if len(s) < 3 || !strings.Contains(s, "@") {
		return false
	}
	parts := strings.Split(s, "@")
	return len(parts) == 2 && parts[0] != "" && strings.Contains(parts[1], ".")
}

// -------------------- Password Hashing (EXACT .NET PBKDF2 format) --------------------

// .NET:
// SALT_SIZE=16, KEY_SIZE=32, ITERATIONS=10000
// returns "{ITERATIONS}.{saltB64}.{keyB64}"
func HashPasswordPBKDF2(password string) (string, error) {
	const (
		saltSize   = 16
		keySize    = 32
		iterations = 10000
	)

	salt := make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	key := pbkdf2.Key([]byte(password), salt, iterations, keySize, sha1.New)

	saltB64 := base64.StdEncoding.EncodeToString(salt)
	keyB64 := base64.StdEncoding.EncodeToString(key)

	return strconv.Itoa(iterations) + "." + saltB64 + "." + keyB64, nil
}

// .NET: Next(0,9) => digits 0..8
func GenerateEmailVerificationCode6Digits_0to8() string {
	b := make([]byte, 6)
	for i := 0; i < 6; i++ {
		n := make([]byte, 1)
		_, _ = rand.Read(n)
		d := int(n[0]) % 9
		b[i] = byte('0' + d)
	}
	return string(b)
}

func GetEmailVerificationExpiryMinutes() int {
	if n := constants.Cfg.ApplicationSettings.EmailVerificationCodeExpiryTimeMinutes; n > 0 {
		return n
	}

	v := strings.TrimSpace(os.Getenv("EMAIL_VERIFICATION_CODE_EXPIRY_MINUTES"))
	if v == "" {
		return 1440
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return 1440
	}
	return n
}

func GetDomainOrDefault(domain string) string {
	domain = NormalizeDomain(domain)
	if domain != "" {
		if isLocalhostDomain(domain) {
			// If gateway passes localhost during local dev, force configured domain so DB logic matches env.
			domain = ""
		} else {
			return domain
		}
	}
	if v := strings.TrimSpace(constants.Cfg.ApplicationSettings.DefaultDomain); v != "" {
		return v
	}
	defaultDomain := strings.TrimSpace(os.Getenv("DEFAULT_DOMAIN"))
	if defaultDomain == "" {
		defaultDomain = "example.com"
	}
	return defaultDomain
}

func NormalizeCultureInfo(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "en-US"
	}
	// Accept-Language like: "en-US,en;q=0.9" => "en-US"
	first := strings.Split(s, ",")[0]
	first = strings.TrimSpace(strings.Split(first, ";")[0])
	if first == "" {
		return "en-US"
	}
	return first
}

func NormalizeDomain(in string) string {
	s := strings.TrimSpace(in)
	if s == "" {
		return ""
	}
	// If scheme present, parse as URL.
	if strings.Contains(s, "://") {
		u, err := url.Parse(s)
		if err == nil && u.Host != "" {
			s = u.Host
		}
	}
	// Strip any path part.
	if i := strings.IndexByte(s, '/'); i >= 0 {
		s = s[:i]
	}
	// Strip port when present.
	if strings.HasPrefix(s, "[") {
		if h, _, err := net.SplitHostPort(s); err == nil {
			return strings.Trim(h, "[]")
		}
		return strings.Trim(strings.Trim(s, "[]"), " ")
	}
	if h, _, err := net.SplitHostPort(s); err == nil {
		return h
	}
	// net.SplitHostPort fails for host:port when missing port or ambiguous; best-effort strip single-colon ports.
	if strings.Count(s, ":") == 1 {
		parts := strings.SplitN(s, ":", 2)
		return parts[0]
	}
	return s
}

func isLocalhostDomain(host string) bool {
	host = strings.TrimSpace(strings.ToLower(host))
	switch host {
	case "localhost", "127.0.0.1", "::1":
		return true
	default:
		return false
	}
}

// -------------------- RequestedProducts CDL --------------------

func ConvertIntsToCdl(arr []int) string {
	if arr == nil {
		return ""
	}
	parts := make([]string, 0, len(arr))
	for _, v := range arr {
		parts = append(parts, strconv.Itoa(v))
	}
	return strings.Trim(strings.Join(parts, ","), ",")
}

// -------------------- DB Layer (SP execution + DbResult mapping) --------------------

func ExecSPDbResult(ctx context.Context, spName string, params map[string]any) (DbResult, error) {
	if db.DB == nil {
		return DbResult{}, errors.New("DB is nil")
	}
	if strings.TrimSpace(spName) == "" {
		return DbResult{}, errors.New("spName is empty")
	}

	logStoredProcedureCall(spName, params)

	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	assignments := make([]string, 0, len(params))
	args := make([]any, 0, len(params))
	for _, k := range keys {
		assignments = append(assignments, "@"+k+"=@"+k)
		args = append(args, sql.Named(k, params[k]))
	}

	q := "EXEC " + spName
	if len(assignments) > 0 {
		q += " " + strings.Join(assignments, ", ")
	}

	row := db.DB.QueryRowContext(ctx, q, args...)
	var id sql.NullInt64
	var status sql.NullString
	var details sql.NullString
	var errs sql.NullString
	if err := row.Scan(&id, &status, &details, &errs); err != nil {
		return DbResult{}, fmt.Errorf("sp scan failed: %w", err)
	}

	out := DbResult{}
	if id.Valid {
		out.ID = int(id.Int64)
		out.Id = out.ID
	}
	if status.Valid {
		out.Status = status.String
	}
	if details.Valid {
		out.Details = details.String
	}
	if errs.Valid {
		out.Errors = errs.String
	}

	return out, nil
}

func logStoredProcedureCall(spName string, params map[string]any) {
	includeSecrets := envBool("SP_LOG_INCLUDE_SECRETS", false)
	safe := sanitizeSPParams(params, includeSecrets)

	// JSON (quick view)
	b, err := json.Marshal(safe)
	if err != nil {
		fmt.Printf("SP call: %s params=<unmarshalable>\n", spName)
	} else {
		fmt.Printf("SP call: %s params=%s\n", spName, string(b))
	}

	// SQL Server runnable script (copy/paste into SSMS)
	fmt.Printf("SP call (SQL):\n%s\n", formatSQLServerExecScript(spName, safe))
}

func sanitizeSPParams(params map[string]any, includeSecrets bool) map[string]any {
	if params == nil {
		return map[string]any{}
	}

	redact := func(key string) bool {
		if includeSecrets {
			return false
		}
		k := strings.ToLower(strings.TrimSpace(key))
		switch k {
		case "passwordhash", "password", "confirmpassword", "emailverificationcode":
			return true
		default:
			return false
		}
	}

	out := make(map[string]any, len(params))
	for k, v := range params {
		if redact(k) {
			out[k] = "***REDACTED***"
			continue
		}
		out[k] = v
	}
	return out
}

func envBool(key string, def bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	return strings.EqualFold(v, "true") || v == "1" || strings.EqualFold(v, "yes")
}

func formatSQLServerExecScript(spName string, params map[string]any) string {
	if params == nil {
		params = map[string]any{}
	}

	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	b.WriteString("-- Copy/paste into SSMS\n")
	b.WriteString("-- NOTE: set SP_LOG_INCLUDE_SECRETS=1 to log real secrets (default redacts)\n")

	b.WriteString("EXEC dbo.")
	b.WriteString(spName)
	if len(keys) > 0 {
		b.WriteString("\n")
	}
	for i, k := range keys {
		b.WriteString("    @")
		b.WriteString(k)
		b.WriteString(" = ")
		b.WriteString(sqlServerLiteral(params[k]))
		if i < len(keys)-1 {
			b.WriteString(",\n")
		} else {
			b.WriteString(";\n")
		}
	}

	return b.String()
}

func sqlServerTypesForSP(spName string) map[string]string {
	// Types based on:
	// ALTER PROCEDURE [dbo].[v2_PublicRole_RegistrationModule_RegisterCustomerUser] (...)
	if strings.EqualFold(spName, "v2_PublicRole_RegistrationModule_RegisterCustomerUser") {
		return map[string]string{
			"FirstName":                   "NVARCHAR(100)",
			"MiddleNames":                 "NVARCHAR(200)",
			"LastName":                    "NVARCHAR(100)",
			"CurrencyCode":                "NVARCHAR(10)",
			"EmailAddress":                "NVARCHAR(200)",
			"PasswordHash":                "NVARCHAR(200)",
			"Domain":                      "NVARCHAR(200)",
			"CultureInfo":                 "NVARCHAR(50)",
			"EmailVerificationCode":       "NVARCHAR(20)",
			"EmailVerificationCodeExpiry": "DATETIME",
			"PhoneNumber":                 "NVARCHAR(50)",
			"ConsentIpAddress":            "NVARCHAR(100)",
			"ConsentCountryIso2Code":      "NVARCHAR(10)",
			"ConsentStateProvinceIsoCode": "NVARCHAR(10)",
			"CountryISOCode":              "NVARCHAR(10)",
			"RequestedProducts":           "NVARCHAR(500)",
			"bConductsThirdPartyPayments": "bit",
		}
	}
	return map[string]string{}
}

func sqlServerLiteral(v any) string {
	if v == nil {
		return "NULL"
	}

	switch t := v.(type) {
	case string:
		return "N'" + escapeSQLString(t) + "'"
	case bool:
		if t {
			return "1"
		}
		return "0"
	case time.Time:
		// SSMS-friendly implicit cast to DATETIME
		ts := t.UTC().Format("2006-01-02T15:04:05.999")
		return "'" + ts + "'"
	case *time.Time:
		if t == nil {
			return "NULL"
		}
		ts := t.UTC().Format("2006-01-02T15:04:05.999")
		return "'" + ts + "'"
	case int:
		return strconv.Itoa(t)
	case int64:
		return strconv.FormatInt(t, 10)
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	default:
		// Best-effort JSON representation for complex types.
		b, err := json.Marshal(t)
		if err != nil {
			return "N'" + escapeSQLString(fmt.Sprintf("%v", t)) + "'"
		}
		return "N'" + escapeSQLString(string(b)) + "'"
	}
}

func escapeSQLString(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

func TransformDbResultToResult(dbRes DbResult) Result {
	if dbRes.Status == "1" {
		return NewSuccessResult(int64(dbRes.Id))
	}
	errList := ParseDbErrors(coerceDbErrorsToString(dbRes.Errors))
	return NewFailureResult(errList...)
}

func coerceDbErrorsToString(v any) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	case fmt.Stringer:
		return t.String()
	default:
		// Best-effort: some callers/DB drivers may return structured data.
		b, err := json.Marshal(t)
		if err == nil {
			return string(b)
		}
		return fmt.Sprintf("%v", t)
	}
}

func ParseDbErrors(dbErrorString string) []ErrorResult {
	dbErrorString = strings.TrimSpace(dbErrorString)
	if dbErrorString == "" {
		return []ErrorResult{ErrInternal("", "Unexpected_Error")}
	}

	const (
		errorDelimiter    = "]|[ "
		fieldMsgDelimiter = "]_["
	)

	tokens := strings.Split(dbErrorString, errorDelimiter)
	var errs []ErrorResult
	for _, tok := range tokens {
		t := strings.TrimSpace(tok)
		if t == "" {
			continue
		}
		parts := strings.Split(t, fieldMsgDelimiter)

		et := ""
		fn := ""
		mc := ""
		if len(parts) > 0 {
			et = strings.TrimSpace(parts[0])
		}
		if len(parts) > 1 {
			fn = strings.TrimSpace(parts[1])
		}
		if len(parts) > 2 {
			mc = strings.TrimSpace(parts[2])
		}

		errs = append(errs, ErrorResult{ErrorType: et, FieldName: fn, MessageCode: mc})
	}

	if len(errs) == 0 {
		return []ErrorResult{ErrInternal("", "Unexpected_Error")}
	}
	return DedupErrors(errs)
}

// FormatDbErrors encodes errors in the same string format DB/.NET typically uses:
// ERROR_DELIMITER = "]|[ "
// FIELD_MESSAGE_CODE_DELIMITER = "]_["
// Output token: "{ErrorType}]_[{FieldName}]_[{MessageCode}"
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

// -------------------- Notification (best-effort placeholder) --------------------

func SendCustomerRegistrationEmailBestEffort(domain, email, firstName, code string, year int, cultureInfo string) error {
	_ = domain
	_ = email
	_ = firstName
	_ = code
	_ = year
	_ = cultureInfo
	return nil
}

// -------------------- Error helpers --------------------

func ErrBadRequest(field, msg string) ErrorResult {
	return ErrorResult{ErrorType: "BadRequest", FieldName: field, MessageCode: msg}
}

func ErrInternal(field, msg string) ErrorResult {
	return ErrorResult{ErrorType: "InternalServerError", FieldName: field, MessageCode: msg}
}

func DedupErrors(in []ErrorResult) []ErrorResult {
	seen := make(map[string]struct{}, len(in))
	out := make([]ErrorResult, 0, len(in))
	for _, e := range in {
		key := e.ErrorType + "|" + e.FieldName + "|" + e.MessageCode
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, e)
	}
	return out
}

func ValidateUploadDocument(req *UploadDocumentRequest) []ErrorResult {
	var errs []ErrorResult

	if req == nil {
		return []ErrorResult{{ErrorType: "BadRequest", FieldName: "", MessageCode: "Invalid_JSON"}}
	}
	if len(req.FileBytes) == 0 {
		errs = append(errs, ErrorResult{"BadRequest", "file", "Required"})
	}
	if strings.TrimSpace(req.FileName) == "" {
		errs = append(errs, ErrorResult{"BadRequest", "fileName", "Required"})
	}
	// if req.BusinessId <= 0 {
	// 	errs = append(errs, ErrorResult{"BadRequest", "businessId", "Required"})
	// }
	// if req.DocumentTypeId <= 0 {
	// 	errs = append(errs, ErrorResult{"BadRequest", "documentTypeId", "Required"})
	// }

	// size limit example (10MB). Adjust to your rules
	const maxSize = 10 * 1024 * 1024
	if len(req.FileBytes) > maxSize {
		errs = append(errs, ErrorResult{"BadRequest", "file", "File_Too_Large"})
	}

	return DedupErrorsFile(errs)
}

func ValidateSaveVerificationForm(req *SaveVerificationFormRequest) []ErrorResultFile {
	var errs []ErrorResultFile

	if req == nil {
		return []ErrorResultFile{{ErrorType: "BadRequest", FieldName: "Json", MessageCode: "Invalid_Json"}}
	}
	cmd := strings.TrimSpace(string(req.Command))
	if cmd == "" || cmd == "null" {
		errs = append(errs, ErrorResultFile{ErrorType: "BadRequest", FieldName: "Json", MessageCode: "Invalid_Json"})
	}

	return DedupErrorsResultFile(errs)
}

func DedupErrorsResultFile(in []ErrorResultFile) []ErrorResultFile {
	seen := map[string]struct{}{}
	out := make([]ErrorResultFile, 0, len(in))
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

func marshalErrorResultFiles(errs []ErrorResultFile) string {
	b, _ := json.Marshal(errs)
	return string(b)
}

// -------------------- Business form helpers (SaveVerificationForm) --------------------

// Azure storage helper (plug your implementation)
type StorageHelper interface {
	CheckFileExists(path string, container string) (bool, error)
}

var Storage StorageHelper

// .NET: private string GetBusinessFormDocumentPath(int customersId)
// => $"BusinessFormDocuments/{customersId}/BusinessDocuments/"
func businessFormDocumentPath(customersId int64) string {
	return fmt.Sprintf("BusinessFormDocuments/%d/BusinessDocuments/", customersId)
}

func buildOwnershipGraph(root OwnershipTreeNode) OwnershipGraph {
	vertices := make([]OwnershipVertex, 0, 64)
	edges := make([]OwnershipEdge, 0, 128)

	vertices = append(vertices, OwnershipVertex{
		OwnersGuid: root.OwnersGuid,
		VertexData: OwnershipVertexData{BIsBusiness: root.BIsBusiness},
	})

	if len(root.Children) > 0 {
		for _, sub := range root.Children {
			hasChildren := len(sub.Children) > 0
			edges = append(edges, OwnershipEdge{
				SourceVertexGuid: root.OwnersGuid,
				TargetVertexGuid: sub.OwnersGuid,
				EdgeData: OwnershipEdgeData{
					PercentageOwned:   sub.PercentageOwned,
					PositionAtCompany: sub.PositionAtCompany,
					BControllingParty: sub.BControllingParty,
				},
				BInSpanningTree: hasChildren,
			})
		}

		for _, sub := range root.Children {
			subGraph := buildOwnershipGraph(sub)
			vertices = mergeVertices(vertices, subGraph.Vertices)
			edges = mergeEdges(edges, subGraph.Edges)
		}
	}

	return OwnershipGraph{Vertices: vertices, Edges: edges}
}

func mergeVertices(base []OwnershipVertex, add []OwnershipVertex) []OwnershipVertex {
	seen := map[string]bool{}
	for _, v := range base {
		seen[v.OwnersGuid] = true
	}
	for _, v := range add {
		if !seen[v.OwnersGuid] {
			seen[v.OwnersGuid] = true
			base = append(base, v)
		}
	}
	return base
}

func edgeKey(e OwnershipEdge) string {
	return e.SourceVertexGuid + "->" + e.TargetVertexGuid
}

func mergeEdges(base []OwnershipEdge, add []OwnershipEdge) []OwnershipEdge {
	seen := map[string]bool{}
	for _, e := range base {
		seen[edgeKey(e)] = true
	}
	for _, e := range add {
		k := edgeKey(e)
		if !seen[k] {
			seen[k] = true
			base = append(base, e)
		}
	}
	return base
}

func boolPtr(v bool) *bool { return &v }

const (
	innerErrorDelimiter = "]|[ "
	innerFieldDelimiter = "]_["
)

func ParseInnerErrors(inner string) []ErrorResultDto {
	inner = strings.TrimSpace(inner)
	if inner == "" {
		return []ErrorResultDto{}
	}

	tokens := strings.Split(inner, innerErrorDelimiter)
	out := make([]ErrorResultDto, 0, len(tokens))
	for _, t := range tokens {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		parts := strings.Split(t, innerFieldDelimiter)
		out = append(out, ErrorResultDto{
			ErrorType:   strings.TrimSpace(getStr(parts, 0)),
			FieldName:   strings.TrimSpace(getStr(parts, 1)),
			MessageCode: strings.TrimSpace(getStr(parts, 2)),
		})
	}
	return out
}

func getStr(a []string, i int) string {
	if i >= 0 && i < len(a) {
		return a[i]
	}
	return ""
}

func DedupErrorsFile(in []ErrorResult) []ErrorResult {
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

// Storage upload stub: replace with your S3 client code
func UploadToStorage(fileName string, contentType string, fileBytes []byte) (storageKey string, publicURL string, err error) {
	sum := sha256.Sum256(fileBytes)
	hash := hex.EncodeToString(sum[:])

	// Example key:
	key := fmt.Sprintf("business-docs/%s/%s", hash, fileName)

	// TODO: call S3 PutObject here
	// publicURL := "https://<bucket>.s3.amazonaws.com/" + key

	return key, "", nil
}

func dbGetSPResult(ctx context.Context, sp string, detailsOut *string, namedParams ...any) (id int64, status string, errorsStr string, err error) {
	// Only allow known SPs
	allowedSPs := map[string]bool{
		"v1_CustomerRole_BusinessModule_GetKYCDocumentList":            true,
		"v1_CustomerRole_BusinessModule_ValidateKYCDocumentUpload":     true,
		"v1_CustomerRole_BusinessModule_CompleteKYCDocumentUpload":     true,
		"v1_CustomerRole_BusinessModule_GetKYCDocumentDetails":         true,
		"v1_CustomerRole_BusinessModule_DeleteKYCDocument":             true,
		"V3_CustomerRole_BusinessModule_GetVerificationForm":           true,
		"v2_CustomerRole_BusinessModule_GetBusinessVerificationStatus": true,
	}
	if !allowedSPs[sp] {
		return 0, "0", "SP not implemented or does not exist", fmt.Errorf("SP not implemented: %s", sp)
	}

	// Convert namedParams (variadic) to map[string]any
	params := map[string]any{}
	if len(namedParams) == 1 {
		// If a single map is passed
		if m, ok := namedParams[0].(map[string]interface{}); ok {
			params = m
		} else if m, ok := namedParams[0].(map[string]any); ok {
			params = m
		}
	} else if len(namedParams)%2 == 0 {
		for i := 0; i < len(namedParams); i += 2 {
			k, _ := namedParams[i].(string)
			k = strings.TrimPrefix(k, "@")
			params[k] = namedParams[i+1]
		}
	}

	// Normalize keys (some legacy code passes "@CustomersId"; ExecSPDbResult adds '@' itself)
	normalized := map[string]any{}
	for k, v := range params {
		kTrim := strings.TrimPrefix(k, "@")
		// If both forms exist, prefer the non-@ key
		if strings.HasPrefix(k, "@") {
			if _, ok := params[kTrim]; ok {
				continue
			}
		}
		normalized[kTrim] = v
	}
	params = normalized

	dbRes, err := ExecSPDbResult(ctx, sp, params)
	if err != nil {
		return 0, "0", err.Error(), err
	}
	if detailsOut != nil {
		if s, ok := dbRes.Details.(string); ok {
			*detailsOut = s
		} else {
			*detailsOut = ""
		}
	}
	if s, ok := dbRes.Errors.(string); ok {
		errorsStr = s
	} else {
		errorsStr = ""
	}
	return int64(dbRes.ID), dbRes.Status, errorsStr, nil
}
