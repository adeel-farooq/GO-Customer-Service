package authmodule

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

func (s *AuthModuleService) RequestTotpQrCode(req *RequestTotpQrCodeRequest, res *RequestTotpQrCodeResponse) error {

	if res == nil {
		return fmt.Errorf("RequestTotpQrCode: nil response pointer")
	}
	if req == nil {
		*res = NewBadRequestInvalidJSONResponse()
		return nil
	}

	if errs := ValidateCredentials(*req); len(errs) > 0 {
		*res = NewValidationErrorResponse(errs)
		return nil
	}

	domain := GetDomainOrDefault(req.Domain)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	out := RequestTotpQrCodeService(ctx, *req, domain)
	if strings.TrimSpace(out.Errors) == "" {
		out.Status = "1"
	}
	*res = out
	res.Id = res.ID

	fmt.Println("here is my func", res)
	return nil
}

// EnableTfa enables either SMS or AuthenticatorApp 2FA for the user.
func (s *AuthModuleService) EnableTfa(req *EnableTfaRequest, res *DbResultRPC) error {
	start := time.Now()
	if res == nil {
		return fmt.Errorf("EnableTfa: nil response pointer")
	}
	if req == nil {
		log.Printf("AuthModule EnableTfa: nil request")
		emptyArr, _ := json.Marshal([]ErrorResultPublic{})
		*res = DbResultRPC{ID: 0, Id: 0, Status: "0", Details: "", Errors: string(emptyArr)}
		return nil
	}

	// NOTE: Do not log password, TfaCode, or TotpSharedSecret.
	log.Printf("AuthModule EnableTfa: start username=%s tfaType=%s", strings.TrimSpace(req.Username), strings.TrimSpace(req.TfaType))

	if errs := ValidateEnableTfa(*req); len(errs) > 0 {
		log.Printf("AuthModule EnableTfa: validation failed username=%s tfaType=%s errs=%d", strings.TrimSpace(req.Username), strings.TrimSpace(req.TfaType), len(errs))
		errArr, _ := json.Marshal(errs)
		*res = DbResultRPC{ID: 0, Id: 0, Status: "0", Details: "", Errors: string(errArr)}
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	domain := GetDomainOrDefault(req.Domain)
	log.Printf("AuthModule EnableTfa: resolved domain=%s username=%s", domain, strings.TrimSpace(req.Username))
	authData, err := GetSiteUsersAuthData(ctx, req.Username, "Customer", domain)
	if b, errJson := json.MarshalIndent(authData, "", "  "); errJson == nil {
		log.Printf("AuthModule EnableTfa: authData=%s", string(b))
	} else {
		log.Printf("AuthModule EnableTfa: authData=%+v", authData)
	}
	if err != nil {
		logAuthFailure("GetSiteUsersAuthData", err)
		log.Printf("AuthModule EnableTfa: db error username=%s domain=%s", strings.TrimSpace(req.Username), domain)
		errArr, _ := json.Marshal([]ErrorResultPublic{{ErrorType: "InternalServerError", FieldName: "", MessageCode: "Unexpected_Error"}})
		*res = DbResultRPC{ID: 0, Id: 0, Status: "0", Details: "", Errors: string(errArr)}
		return nil
	}

	loginStatus := ValidateUsernameAndPassword(req.Password, authData)
	if loginStatus != LoginStatusSuccess {
		log.Printf("AuthModule EnableTfa: login failed username=%s status=%s", strings.TrimSpace(req.Username), loginStatus)
		errs := GetUsernameAndPasswordValidationFailureErrors(loginStatus)
		pub := make([]ErrorResultPublic, 0, len(errs))
		for _, e := range errs {
			pub = append(pub, ErrorResultPublic{ErrorType: e.ErrorType, FieldName: e.FieldName, MessageCode: e.MessageCode})
		}
		errArr, _ := json.Marshal(pub)
		*res = DbResultRPC{ID: 0, Id: 0, Status: "0", Details: "", Errors: string(errArr)}
		return nil
	}

	if authData != nil {
		log.Printf("AuthModule EnableTfa: auth ok username=%s siteUsersId=%d", strings.TrimSpace(req.Username), authData.SiteUsersId)
	}

	// already enabled checks
	if req.TfaType == "AuthenticatorApp" && authData != nil && authData.TwoFactorAppAuthEnabled {
		log.Printf("AuthModule EnableTfa: already enabled (app) username=%s siteUsersId=%d", strings.TrimSpace(req.Username), authData.SiteUsersId)
		errArr, _ := json.Marshal([]ErrorResultPublic{
			{ErrorType: "BadRequest", FieldName: "Username", MessageCode: "AuthenticatorAppTfa_Already_Enabled"},
			{ErrorType: "BadRequest", FieldName: "Password", MessageCode: "AuthenticatorAppTfa_Already_Enabled"},
		})
		*res = DbResultRPC{ID: 0, Id: 0, Status: "0", Details: "", Errors: string(errArr)}
		return nil
	}
	if req.TfaType == "SMS" && authData != nil && authData.TwoFactorSMSAuthEnabled {
		log.Printf("AuthModule EnableTfa: already enabled (sms) username=%s siteUsersId=%d", strings.TrimSpace(req.Username), authData.SiteUsersId)
		errArr, _ := json.Marshal([]ErrorResultPublic{
			{ErrorType: "BadRequest", FieldName: "Username", MessageCode: "SmsTfa_Already_Enabled"},
			{ErrorType: "BadRequest", FieldName: "Password", MessageCode: "SmsTfa_Already_Enabled"},
		})
		*res = DbResultRPC{ID: 0, Id: 0, Status: "0", Details: "", Errors: string(errArr)}
		return nil
	}

	// validate TFA code
	st := validateTfaCode(*req, authData)
	if st != tfaOK {
		log.Printf("AuthModule EnableTfa: tfa code invalid username=%s tfaType=%s status=%s", strings.TrimSpace(req.Username), strings.TrimSpace(req.TfaType), st)
		errs := tfaCodeErrors(st)
		if b, err := json.Marshal(errs); err == nil {
			log.Printf("AuthModule EnableTfa: errors=%s", string(b))
		} else {
			log.Printf("AuthModule EnableTfa: errors=%v", errs)
		}
		errArr, _ := json.Marshal(errs)
		*res = DbResultRPC{ID: 0, Id: 0, Status: "0", Details: "", Errors: string(errArr)}
		return nil
	}

	// enable in DB
	if authData == nil {
		log.Printf("AuthModule EnableTfa: authData nil after auth ok? username=%s", strings.TrimSpace(req.Username))
		errArr, _ := json.Marshal([]ErrorResultPublic{{ErrorType: "BadRequest", FieldName: "Username", MessageCode: "Username_Or_Password_Incorrect"}})
		*res = DbResultRPC{ID: 0, Id: 0, Status: "0", Details: "", Errors: string(errArr)}
		return nil
	}

	if req.TfaType == "SMS" {
		sp := "v1_General_Security_EnableSmsTfa"
		q := "EXEC " + sp + " @SiteUsersId=@SiteUsersId, @PhoneNumber=@PhoneNumber"
		log.Printf("AuthModule EnableTfa: running SQL: %s -- @SiteUsersId=%d, @PhoneNumber='%s'", q, authData.SiteUsersId, req.PhoneNumber)
		if err := enableSmsTfa(ctx, authData.SiteUsersId, req.PhoneNumber); err != nil {
			logAuthFailure("EnableSmsTfa", err)
			log.Printf("AuthModule EnableTfa: enable sms failed username=%s siteUsersId=%d", strings.TrimSpace(req.Username), authData.SiteUsersId)
			errArr, _ := json.Marshal([]ErrorResultPublic{{ErrorType: "InternalServerError", FieldName: "", MessageCode: "Unexpected_Error"}})
			*res = DbResultRPC{ID: 0, Id: 0, Status: "0", Details: "", Errors: string(errArr)}
			return nil
		}
		authData.TwoFactorSMSAuthEnabled = true
	} else {
		sp := "v1_General_Security_EnableAppTfa"
		q := "EXEC " + sp + " @SiteUsersId=@SiteUsersId, @TotpSharedSecret=@TotpSharedSecret"
		log.Printf("AuthModule EnableTfa: running SQL: %s -- @SiteUsersId=%d, @TotpSharedSecret='%s'", q, authData.SiteUsersId, req.TotpSharedSecret)
		if err := enableAppTfa(ctx, authData.SiteUsersId, req.TotpSharedSecret); err != nil {
			logAuthFailure("EnableAppTfa", err)
			log.Printf("AuthModule EnableTfa: enable app failed username=%s siteUsersId=%d", strings.TrimSpace(req.Username), authData.SiteUsersId)
			errArr, _ := json.Marshal([]ErrorResultPublic{{ErrorType: "InternalServerError", FieldName: "", MessageCode: "Unexpected_Error"}})
			*res = DbResultRPC{ID: 0, Id: 0, Status: "0", Details: "", Errors: string(errArr)}
			return nil
		}
		authData.TwoFactorAppAuthEnabled = true
		authData.TotpSharedSecret = req.TotpSharedSecret
	}

	// success (Details is RAW JSON string)
	emptyArr, _ := json.Marshal([]ErrorResultPublic{})
	*res = DbResultRPC{
		ID:      authData.SiteUsersId,
		Id:      authData.SiteUsersId,
		Status:  "1",
		Details: enableTfaDetailsJSON(authData),
		Errors:  string(emptyArr),
	}

	// safety: if errors is non-empty, force status 0
	if strings.TrimSpace(res.Errors) != "" {
		res.Status = "0"
	}

	log.Printf("AuthModule EnableTfa: done username=%s siteUsersId=%d status=%s took=%s", strings.TrimSpace(req.Username), res.ID, res.Status, time.Since(start))
	return nil
}
