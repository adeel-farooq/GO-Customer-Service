package profile

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	customer "go-cloud-customer/controllers/customer"
)

func withTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}

func errorsJSON(errs any) string {
	b, _ := json.Marshal(errs)
	return string(b)
}

func parseLegacyDbErrorString(s string) []ErrorResultDto {
	var results []ErrorResultDto
	s = strings.TrimSpace(s)
	if s == "" {
		return results
	}
	parts := strings.Split(s, "|[")
	for _, part := range parts {
		part = strings.Trim(part, " []|\t\n\r")
		if part == "" {
			continue
		}
		fields := strings.Split(part, "]_[")
		for len(fields) < 3 {
			fields = append(fields, "")
		}
		results = append(results, ErrorResultDto{
			ErrorType:   strings.TrimSpace(fields[0]),
			FieldName:   strings.TrimSpace(fields[1]),
			MessageCode: strings.TrimSpace(fields[2]),
		})
	}
	return results
}

// returns: id, status, errorsStr, err ; detailsOut is JSON string output column
func dbGetSPResult(ctx context.Context, sp string, detailsOut *string, params map[string]any) (int64, string, string, error) {
	if strings.TrimSpace(sp) == "" {
		return 0, "0", "spName is empty", fmt.Errorf("spName is empty")
	}

	// Only allow known SPs for this package.
	allowedSPs := map[string]bool{
		"v2_CustomerRole_ProfileModule_GetUserInfo": true,
	}
	if !allowedSPs[sp] {
		return 0, "0", "SP not implemented or does not exist", fmt.Errorf("SP not implemented: %s", sp)
	}

	normalized := map[string]any{}
	for k, v := range params {
		kTrim := strings.TrimPrefix(k, "@")
		normalized[kTrim] = v
	}

	dbRes, err := customer.ExecSPDbResult(ctx, sp, normalized)
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

	errorsStr := ""
	if s, ok := dbRes.Errors.(string); ok {
		errorsStr = s
	}

	id := int64(dbRes.ID)
	status := dbRes.Status
	return id, status, errorsStr, nil
}
