package restricted

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// RPC method: "RestrictedService.GetRegistrationDetails"
// Returns raw JSON string in Details and raw error string in Errors.
func (s *RestrictedService) GetRegistrationDetails(_ *struct{}, res *DbResultRPC) error {
	if res == nil {
		return fmt.Errorf("GetRegistrationDetails: nil response pointer")
	}

	ctx, cancel := withTimeout()
	defer cancel()

	sp := "v2_PublicRole_RegistrationModule_GetRegistrationDetails"

	var detailsStr string
	id, status, errorsStr, err := dbGetSPResult(ctx, sp, &detailsStr)
	if err != nil {
		*res = DbResultRPC{ID: 0, Id: 0, Status: "0", Details: "", Errors: err.Error()}
		return nil
	}

	*res = DbResultRPC{
		ID:      id,
		Id:      id,
		Status:  status,
		Details: detailsStr,
		Errors:  errorsStr,
	}

	if strings.TrimSpace(res.Errors) == "" {
		res.Status = "1"
	}

	return nil
}

func (s *RestrictedService) GetErrorMessages(req *GetErrorMessagesRequest, res *DbResult) error {
	// start := time.Now()
	if res == nil {
		return fmt.Errorf("GetErrorMessages: nil response pointer")
	}

	if req == nil {
		*res = DbResult{
			ID:      0,
			Id:      0,
			Status:  "0",
			Details: nil,
			Errors:  []ErrorResult{{ErrorType: "BadRequest", FieldName: "", MessageCode: "Invalid_JSON"}},
		}
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var detailsStr string

	id, status, errorsStr, err := dbGetSPResult(ctx, "v1_PublicRole_CountryRestrictedActionsModule_GetErrorMessages", &detailsStr)
	if err != nil {
		*res = DbResult{ID: 0, Id: 0, Status: "0", Details: nil, Errors: err.Error()}
		return nil
	}

	// ✅ Microservice -> gateway safe: send raw JSON string
	*res = DbResult{
		ID:      id,
		Id:      id,
		Status:  status,
		Details: detailsStr, // <-- RAW JSON STRING
		Errors:  errorsStr,
	}

	if strings.TrimSpace(errorsStr) == "" {
		res.Status = "1"
	}

	return nil
}
func (s *RestrictedService) GetCountryANDProduct(req *GetErrorMessagesRequest, res *DbResultRPC) error {
	start := time.Now()

	if res == nil {
		return fmt.Errorf("GetCountryANDProduct: nil response pointer")
	}

	if req == nil {
		*res = DbResultRPC{
			ID: 0, Id: 0, Status: "0",
			Details: "",
			Errors:  "Invalid_JSON",
		}
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// SP returns JSON string in Details column
	var detailsStr string

	id, status, errorsStr, err := dbGetSPResult(ctx, "v2_PublicRole_RegistrationModule_GetRegistrationDetails", &detailsStr)
	if err != nil {
		logDbFailure("v2_PublicRole_RegistrationModule_GetRegistrationDetails", err, time.Since(start))
		*res = DbResultRPC{ID: 0, Id: 0, Status: "0", Details: "", Errors: err.Error()}
		return nil
	}

	*res = DbResultRPC{
		ID:      id,
		Id:      id,
		Status:  status,
		Details: detailsStr, // ✅ raw JSON string
		Errors:  errorsStr,  // ✅ raw error string
	}

	if strings.TrimSpace(errorsStr) == "" {
		res.Status = "1"
	}

	return nil
}
