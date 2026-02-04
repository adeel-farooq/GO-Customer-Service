package restricted

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"go-cloud-customer/db"
	"log"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

func withTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}

const (
	errorDelimiter = "]|[ "
	fieldDelimiter = "]_["
)

func ParseDbErrors(dbErrors string) []ErrorResult {
	dbErrors = strings.TrimSpace(dbErrors)
	if dbErrors == "" {
		return []ErrorResult{{ErrorType: "InternalServerError", FieldName: "", MessageCode: "Unexpected_Error"}}
	}
	tokens := strings.Split(dbErrors, errorDelimiter)
	out := make([]ErrorResult, 0, len(tokens))
	for _, t := range tokens {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		parts := strings.Split(t, fieldDelimiter)
		er := ErrorResult{
			ErrorType:   strings.TrimSpace(get(parts, 0)),
			FieldName:   strings.TrimSpace(get(parts, 1)),
			MessageCode: strings.TrimSpace(get(parts, 2)),
		}
		out = append(out, er)
	}
	return out
}

func get(arr []string, i int) string {
	if i >= 0 && i < len(arr) {
		return arr[i]
	}
	return ""
}

const (
	errorMessagesCacheKey = "Error_Messages"
	errorMessagesCacheTTL = time.Hour
)

type cacheItem struct {
	value     DbResult
	expiresAt time.Time
}

var (
	cacheMu sync.RWMutex
	cache   = make(map[string]cacheItem)
)

/* ===================== CACHE ===================== */

func cacheGet(key string) (DbResult, bool) {
	cacheMu.RLock()
	it, ok := cache[key]
	cacheMu.RUnlock()

	if !ok {
		return DbResult{}, false
	}

	if time.Now().UTC().After(it.expiresAt) {
		cacheMu.Lock()
		delete(cache, key)
		cacheMu.Unlock()
		return DbResult{}, false
	}

	return it.value, true
}

func cacheSet(key string, v DbResult, ttl time.Duration) {
	cacheMu.Lock()
	cache[key] = cacheItem{
		value:     v,
		expiresAt: time.Now().UTC().Add(ttl),
	}
	cacheMu.Unlock()
}

/* ===================== DB ===================== */

func dbGetSPResult(
	ctx context.Context,
	spName string,
	detailsOut any,
) (id int, status string, errMsg string, err error) {

	if db.DB == nil {
		return 0, "0", "DB is nil", errors.New("DB is nil")
	}

	start := time.Now()
	rows, err := db.DB.QueryContext(ctx, "EXEC "+spName)
	if err != nil {
		logDbFailure(spName, err, time.Since(start))
		return 0, "0", err.Error(), err
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return 0, "0", err.Error(), err
		}
		return 0, "0", "no rows returned", errors.New("no rows returned")
	}

	cols, err := rows.Columns()
	if err != nil {
		return 0, "0", err.Error(), err
	}

	raw := make([]sql.RawBytes, len(cols))
	dest := make([]any, len(cols))
	for i := range raw {
		dest[i] = &raw[i]
	}

	if err := rows.Scan(dest...); err != nil {
		return 0, "0", err.Error(), err
	}

	get := func(name string) string {
		for i, c := range cols {
			if strings.EqualFold(strings.TrimSpace(c), name) {
				return strings.TrimSpace(string(raw[i]))
			}
		}
		return ""
	}

	if n, err2 := strconv.Atoi(get("Id")); err2 == nil {
		id = n
	}

	status = get("Status")
	errMsg = get("Errors")

	// ---- DETAILS JSON ----
	detailsJSON := get("Details")
	if detailsOut != nil && detailsJSON != "" {
		switch v := detailsOut.(type) {
		case *string:
			*v = detailsJSON
		case *interface{}:
			parsed, err := ParseAnyJSON([]byte(detailsJSON))
			if err == nil {
				*v = parsed
			} else {
				*v = nil
				log.Printf("SP %s: ParseAnyJSON failed: %v", spName, err)
			}
		default:
			if err := json.Unmarshal([]byte(detailsJSON), detailsOut); err != nil {
				log.Printf("SP %s: Details JSON unmarshal failed: %v", spName, err)
				resetPointer(detailsOut)
			}
		}
	}

	log.Printf("SP %s executed successfully in %s", spName, time.Since(start))
	return id, status, errMsg, nil
}

/* ===================== HELPERS ===================== */

func resetPointer(v any) {
	rv := reflect.ValueOf(v)
	if rv.IsValid() && rv.Kind() == reflect.Ptr && !rv.IsNil() {
		rv.Elem().Set(reflect.Zero(rv.Elem().Type()))
	}
}

func errBadRequest(field, msg string) ErrorResult {
	return ErrorResult{
		ErrorType:   "BadRequest",
		FieldName:   field,
		MessageCode: msg,
	}
}

func errInternal(field, msg string) ErrorResult {
	return ErrorResult{
		ErrorType:   "InternalServerError",
		FieldName:   field,
		MessageCode: msg,
	}
}

func formatDbErrors(errs []ErrorResult) string {
	if len(errs) == 0 {
		return ""
	}

	out := make([]string, 0, len(errs))
	for _, e := range errs {
		out = append(out, e.ErrorType+"]_["+e.FieldName+"]_["+e.MessageCode)
	}
	return strings.Join(out, "]|[ ")
}

func logDbFailure(sp string, err error, elapsed time.Duration) {
	log.Printf("DB/SP failed sp=%q err=%v elapsed=%s", sp, err, elapsed)
}

// ParseAnyJSON parses arbitrary JSON into interface{} using json.Number for numbers
func ParseAnyJSON(b []byte) (interface{}, error) {
	var out interface{}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	err := dec.Decode(&out)
	return out, err
}

func parseDetailsJSON(detailsStr string) any {
	detailsStr = strings.TrimSpace(detailsStr)
	if detailsStr == "" || detailsStr == "null" {
		return nil
	}

	var out any
	if err := json.Unmarshal([]byte(detailsStr), &out); err != nil {
		// JSON invalid => return raw string
		return detailsStr
	}
	return out
}
func toCamelCase(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}

	// Handle snake_case, PascalCase, etc
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})

	for i := range parts {
		if i == 0 {
			parts[i] = strings.ToLower(parts[i][:1]) + parts[i][1:]
		} else {
			parts[i] = strings.ToUpper(parts[i][:1]) + strings.ToLower(parts[i][1:])
		}
	}

	return strings.Join(parts, "")
}
func mapKeysToCamelCase(data interface{}) interface{} {
	switch v := data.(type) {

	case map[string]interface{}:
		mapped := make(map[string]interface{}, len(v))
		for key, value := range v {
			camelKey := toCamelCase(key)
			mapped[camelKey] = mapKeysToCamelCase(value)
		}
		return mapped

	case []interface{}:
		for i := range v {
			v[i] = mapKeysToCamelCase(v[i])
		}
		return v

	default:
		// string, number, bool, nil
		return v
	}
}
