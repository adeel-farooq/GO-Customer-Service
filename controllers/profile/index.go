package profile

import (
	"encoding/json"
	"fmt"
	"strings"
)

type ProfileService struct{}

// RPC name: "ProfileService.GetUserInfo"
func (s *ProfileService) GetUserInfo(req *GetUserInfoRequest, res *GetUserInfoResult) error {
	if res == nil {
		return fmt.Errorf("GetUserInfo: nil response pointer")
	}

	// nil req -> standard bad request
	if req == nil {
		bad := []ErrorResultDto{{ErrorType: "BadRequest", FieldName: "Json", MessageCode: "Invalid_Json"}}
		*res = GetUserInfoResult{
			ID:      0,
			Status:  "0",
			Details: GetUserInfo{},
			Errors:  []string{string(errorsJSON(bad))}, // if errorsJSON returns []byte or string
		}
		return nil
	}

	ctx, cancel := withTimeout()
	defer cancel()

	sp := "v2_CustomerRole_ProfileModule_GetUserInfo"

	var detailsStr string
	id, status, errorsStr, err := dbGetSPResult(ctx, sp, &detailsStr, map[string]any{
		"@User_SiteUsersID": req.SiteUsersId,
		"@CustomersID":      req.CustomersId,
	})
	if err != nil {
		*res = GetUserInfoResult{ID: 0, Status: "0", Details: GetUserInfo{}, Errors: []string{err.Error()}}
		return nil
	}

	// ✅ Parse details JSON -> struct (important for gob + frontend)
	var details GetUserInfo
	if strings.TrimSpace(detailsStr) != "" {
		if e := json.Unmarshal([]byte(detailsStr), &details); e != nil {
			*res = GetUserInfoResult{
				ID:      id,
				Status:  "0",
				Details: GetUserInfo{},
				Errors:  []string{"Invalid_Details_JSON"},
			}
			return nil
		}
	}

	// ✅ Errors normalize -> []string
	errs := normalizeErrorsToSlice(errorsStr)

	// ✅ Status rule: errors empty -> 1 else 0 (tumhari existing logic)
	statusOut := strings.TrimSpace(status)
	if len(errs) == 0 {
		statusOut = "1"
	} else {
		statusOut = "0"
	}

	*res = GetUserInfoResult{
		ID:      id,
		Status:  statusOut,
		Details: details,
		Errors:  errs,
	}
	return nil
}
