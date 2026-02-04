package profile

import (
	"encoding/json"
	"fmt"
	"strings"
)

type ProfileService struct{}

// RPC name: "ProfileService.GetUserInfo"
func (s *ProfileService) GetUserInfo(req *GetUserInfoRequest, res *DbResultRPC) error {
	if res == nil {
		return fmt.Errorf("GetUserInfo: nil response pointer")
	}
	if req == nil {
		bad := []ErrorResultDto{{ErrorType: "BadRequest", FieldName: "Json", MessageCode: "Invalid_Json"}}
		*res = DbResultRPC{ID: 0, Id: 0, Status: "0", Details: "", Errors: errorsJSON(bad)}
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
		*res = DbResultRPC{ID: 0, Id: 0, Status: "0", Details: "", Errors: err.Error()}
		return nil
	}

	finalErrors := errorsStr
	statusOut := status
	if strings.TrimSpace(errorsStr) == "" {
		statusOut = "1"
	} else {
		statusOut = "0"
		trimmed := strings.TrimSpace(errorsStr)
		if !strings.HasPrefix(trimmed, "[") {
			parsed := parseLegacyDbErrorString(errorsStr)
			b, _ := json.Marshal(parsed)
			finalErrors = string(b)
		}
	}

	*res = DbResultRPC{ID: id, Id: id, Status: statusOut, Details: detailsStr, Errors: finalErrors}
	return nil
}
