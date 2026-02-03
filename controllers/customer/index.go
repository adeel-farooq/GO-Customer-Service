package customer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

func (s *CustomerService) RegisterCustomerUser(req *CustomerUserRegistrationRequest, res *DbResult) error {
	start := time.Now()

	if res == nil {
		return fmt.Errorf("RegisterCustomerUser: nil response pointer")
	}
	if req == nil {

		*res = DbResult{ID: 0, Id: 0, Status: "0", Details: "", Errors: FormatDbErrors([]ErrorResult{ErrBadRequest("", "Invalid_JSON")})}
		return nil
	}

	if errs := ValidateRegisterCustomerUser(*req); len(errs) > 0 {
		for _, e := range errs {
			log.Printf("RegisterCustomerUser: validation error field=%q code=%q", e.FieldName, e.MessageCode)
		}
		*res = DbResult{ID: 0, Id: 0, Status: "0", Details: "", Errors: FormatDbErrors(errs)}
		return nil
	}

	passwordHash, err := HashPasswordPBKDF2(req.Password)
	if err != nil {
		log.Printf("RegisterCustomerUser: password hash failed err=%v elapsed=%s", err, time.Since(start))
		*res = DbResult{ID: 0, Id: 0, Status: "0", Details: "", Errors: FormatDbErrors([]ErrorResult{ErrInternal("", "Unexpected_Error")})}
		return nil
	}

	emailVerificationCode := GenerateEmailVerificationCode6Digits_0to8()

	cultureInfo := NormalizeCultureInfo(req.CultureInfo)

	expiryMins := GetEmailVerificationExpiryMinutes()
	emailVerificationExpiry := time.Now().UTC().Add(time.Duration(expiryMins) * time.Minute)

	domain := GetDomainOrDefault(req.Domain)

	requestedProductsCdl := ConvertIntsToCdl(req.RequestedProducts)

	sp := "v2_PublicRole_RegistrationModule_RegisterCustomerUser"
	params := map[string]any{
		"FirstName":                   req.FirstName,
		"MiddleNames":                 req.MiddleNames,
		"LastName":                    req.LastName,
		"CurrencyCode":                req.CurrencyCode,
		"EmailAddress":                req.EmailAddress,
		"PasswordHash":                passwordHash,
		"Domain":                      domain,
		"CultureInfo":                 cultureInfo,
		"EmailVerificationCode":       emailVerificationCode,
		"EmailVerificationCodeExpiry": emailVerificationExpiry,
		"PhoneNumber":                 req.PhoneNumber,
		"ConsentIpAddress":            req.ConsentIpAddress,
		"ConsentCountryIso2Code":      req.ConsentCountryIso2Code,
		"ConsentStateProvinceIsoCode": req.ConsentStateProvinceIsoCode,
		"CountryISOCode":              req.RegisteredCountryCode,
		"RequestedProducts":           requestedProductsCdl,
		"bConductsThirdPartyPayments": req.BConductsThirdPartyPayments,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dbRes, err := ExecSPDbResult(ctx, sp, params)
	if err != nil {
		log.Printf("RegisterCustomerUser: DB/SP failed sp=%q err=%v elapsed=%s", sp, err, time.Since(start))
		*res = DbResult{ID: 0, Id: 0, Status: "0", Details: "", Errors: err.Error()}
		return nil
	}

	if dbRes.Status == "1" {
		_ = SendCustomerRegistrationEmailBestEffort(
			domain,
			req.EmailAddress,
			req.FirstName,
			emailVerificationCode,
			time.Now().UTC().Year(),
			cultureInfo,
		)
	}

	*res = dbRes
	res.Id = res.ID
	return nil
}

// RPC method: "BusinessService.UploadDocument"
func (s *BusinessService) UploadDocument(req *UploadDocumentRequest, res *DbResultFile) error {
	if res == nil {
		return fmt.Errorf("UploadDocument: nil response pointer")
	}

	if req == nil {
		errJSON, _ := json.Marshal([]ErrorResultFile{{ErrorType: "BadRequest", FieldName: "", MessageCode: "Invalid_JSON"}})
		*res = DbResultFile{
			ID: 0, Id: 0, Status: "0",
			Details: "",
			Errors:  string(errJSON),
		}
		return nil
	}

	// Validate like .NET style
	if errs := ValidateUploadDocument(req); len(errs) > 0 {
		errJSON, _ := json.Marshal(errs)
		*res = DbResultFile{ID: 0, Id: 0, Status: "0", Details: "", Errors: string(errJSON)}
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1) Upload to storage
	key, url, err := UploadToStorage(req.FileName, req.ContentType, req.FileBytes)
	if err != nil {
		*res = DbResultFile{ID: 0, Id: 0, Status: "0", Details: "", Errors: err.Error()}
		return nil
	}

	// 2) Save metadata in DB
	// IMPORTANT: SP name is a placeholder. Replace with your actual SP.
	sp := "v3_CustomerRole_BusinessModule_UploadDocument"

	var detailsStr string
	id, status, errorsStr, dbErr := dbGetSPResult(
		ctx, sp, &detailsStr,
		"SiteUsersId", req.SiteUsersId,
		"BusinessId", req.BusinessId,
		"DocumentTypeId", req.DocumentTypeId,
		"DocumentName", req.DocumentName,
		"Description", req.Description,
		"FileName", req.FileName,
		"ContentType", req.ContentType,
		"StorageKey", key,
		"DocumentUrl", url,
	)

	if dbErr != nil {
		*res = DbResultFile{ID: 0, Id: 0, Status: "0", Details: "", Errors: dbErr.Error()}
		return nil
	}

	// 3) Microservice -> gateway safe (RAW JSON string)
	*res = DbResultFile{
		ID:      id,
		Id:      id,
		Status:  status,
		Details: detailsStr, // RAW JSON STRING
		Errors:  errorsStr,  // string
	}

	// If errors empty -> status 1 (same pattern as your sample)
	if strings.TrimSpace(res.Errors) == "" {
		res.Status = "1"
	}

	return nil
}

// RPC method: "BusinessService.SaveVerificationForm"
func (s *BusinessService) SaveVerificationForm(req *SaveVerificationFormRequest, res *DbResultFile) error {
	if res == nil {
		return fmt.Errorf("SaveVerificationForm: nil response pointer")
	}

	// controller: if command == null => Invalid_Json
	if req == nil || strings.TrimSpace(string(req.Command)) == "" || strings.TrimSpace(string(req.Command)) == "null" {
		*res = DbResultFile{
			ID: 0, Id: 0, Status: "0",
			Details: "",
			Errors:  marshalErrorResultFiles([]ErrorResultFile{{ErrorType: "BadRequest", FieldName: "Json", MessageCode: "Invalid_Json"}}),
		}
		return nil
	}

	if errs := ValidateSaveVerificationForm(req); len(errs) > 0 {
		*res = DbResultFile{ID: 0, Id: 0, Status: "0", Details: "", Errors: marshalErrorResultFiles(errs)}
		return nil
	}

	// Parse minimal fields from command
	var cmd VerificationForm
	if err := json.Unmarshal(req.Command, &cmd); err != nil {
		*res = DbResultFile{
			ID: 0, Id: 0, Status: "0",
			Details: "",
			Errors:  marshalErrorResultFiles([]ErrorResultFile{{ErrorType: "BadRequest", FieldName: "Json", MessageCode: "Invalid_Json"}}),
		}
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	sp := "v3_CustomerRole_BusinessModule_SaveVerificationForm"

	// ownerValidationResult / registrationValidationResult (placeholder until rules are ported)
	ownerValidation := ResultDto{Id: 0, Errors: []ErrorResultFile{}}
	regValidation := ResultDto{Id: 0, Errors: []ErrorResultFile{}}
	jsonOwnerValidation, _ := json.Marshal(ownerValidation)
	jsonRegValidation, _ := json.Marshal(regValidation)

	// OwnershipGraph + OwnershipTree JSON
	jsonOwnershipGraph := "{}"
	jsonOwnershipTree := "null"
	if cmd.OwnerInformation != nil && len(cmd.OwnerInformation.BeneficialOwnersStructure) > 0 {
		bos := strings.TrimSpace(string(cmd.OwnerInformation.BeneficialOwnersStructure))
		if bos != "" && bos != "null" {
			jsonOwnershipTree = bos
			var children []OwnershipTreeNode
			_ = json.Unmarshal(cmd.OwnerInformation.BeneficialOwnersStructure, &children)
			root := OwnershipTreeNode{
				BIsBusiness:       boolPtr(true),
				OwnersGuid:        "",
				PercentageOwned:   0,
				BControllingParty: boolPtr(false),
				PositionAtCompany: "",
				Children:          children,
			}
			graph := buildOwnershipGraph(root)
			b, _ := json.Marshal(graph)
			jsonOwnershipGraph = string(b)
		}
	}

	// Missing filenames check (best-effort; Storage can be nil)
	missing := make([]string, 0, 16)
	if Storage != nil {
		basePath := businessFormDocumentPath(req.CustomersId)
		container := "" // TODO: read AzureBlobSecureRootContainerName from config

		if cmd.OwnerInformation != nil {
			for _, bo := range cmd.OwnerInformation.IndividualBeneficialOwners {
				fn := strings.TrimSpace(bo.ProofOfAddressFilename)
				if fn == "" {
					continue
				}
				blobName := fmt.Sprintf("%d_%d_%s", req.CustomersId, req.SiteUsersId, fn)
				exists, _ := Storage.CheckFileExists(basePath+blobName, container)
				if !exists {
					missing = append(missing, fn)
				}
			}
		}

		if cmd.OperationsInformation != nil {
			fn := strings.TrimSpace(cmd.OperationsInformation.FinancialInstitutionFormFileName)
			if fn != "" {
				blobName := fmt.Sprintf("%d_%d_%s", req.CustomersId, req.SiteUsersId, fn)
				exists, _ := Storage.CheckFileExists(basePath+blobName, container)
				if !exists {
					missing = append(missing, fn)
				}
			}
		}
	}

	jsonData := string(req.Command)

	params := map[string]any{
		"SiteUsersID":                      req.SiteUsersId,
		"CustomersID":                      req.CustomersId,
		"BusinessVerificationStep":         cmd.BusinessVerificationStep,
		"JsonData":                         jsonData,
		"JsonOwnerValidationResult":        string(jsonOwnerValidation),
		"JsonRegistrationValidationResult": string(jsonRegValidation),
		"JsonOwnershipGraph":               jsonOwnershipGraph,
		"JsonOwnershipTree":                jsonOwnershipTree,
		"MissingFilenames":                 strings.Join(missing, ","),
		"AddedBy":                          req.AddedBy,
	}

	dbRes, err := ExecSPDbResult(ctx, sp, params)
	if err != nil {
		*res = DbResultFile{ID: 0, Id: 0, Status: "0", Details: "", Errors: err.Error()}
		return nil
	}

	var detailsStr string
	if s, ok := dbRes.Details.(string); ok {
		detailsStr = s
	}
	var errorsStr string
	if s, ok := dbRes.Errors.(string); ok {
		errorsStr = s
	}

	id := int64(dbRes.ID)
	status := dbRes.Status
	if strings.TrimSpace(errorsStr) == "" {
		status = "1"
	} else {
		status = "0"
	}

	// Always marshal errors as JSON array of objects (string), even if legacy string from DB
	finalErrors := errorsStr
	if strings.TrimSpace(errorsStr) != "" {
		trimmed := strings.TrimSpace(errorsStr)
		if !strings.HasPrefix(trimmed, "[") { // legacy string, not JSON array
			parsed := parseLegacyDbErrorString(errorsStr)
			b, _ := json.Marshal(parsed)
			finalErrors = string(b)
		}
	}

	*res = DbResultFile{
		ID:      id,
		Id:      id,
		Status:  status,
		Details: detailsStr,
		Errors:  finalErrors,
	}
	return nil
}
