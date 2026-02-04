package customer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

// SubmitVerificationForm RPC method for business verification
func (s *BusinessService) SubmitVerificationForm(req *SubmitVerificationFormRequest, res *DbResultRPC) error {
	if res == nil {
		return fmt.Errorf("SubmitVerificationForm: nil response pointer")
	}
	if req == nil || len(req.Command) == 0 || string(req.Command) == "null" {
		*res = DbResultRPC{
			ID: 0, Id: 0, Status: "0",
			Details: "",
			Errors:  `[{"errorType":"BadRequest","fieldName":"Json","messageCode":"Invalid_Json"}]`,
		}
		return nil
	}

	// Parse minimal needed fields (reuse VerificationForm struct from saveverificationform)
	var cmd VerificationForm
	if err := json.Unmarshal(req.Command, &cmd); err != nil {
		*res = DbResultRPC{
			ID: 0, Id: 0, Status: "0",
			Details: "",
			Errors:  `[{"errorType":"BadRequest","fieldName":"Json","messageCode":"Invalid_Json"}]`,
		}
		return nil
	}

	// REUSE: owner/registration validation (if you already have, plug it)
	ownerValidation := ResultDto{Id: 0, Errors: EmptyErrors()}
	regValidation := ResultDto{Id: 0, Errors: EmptyErrors()}

	// REUSE: ownership graph + tree (same as saveverificationform)
	jsonOwnershipGraph := "{}"
	jsonOwnershipTree := "null"
	if cmd.OwnerInformation != nil && len(cmd.OwnerInformation.BeneficialOwnersStructure) > 0 && string(cmd.OwnerInformation.BeneficialOwnersStructure) != "null" {
		jsonOwnershipTree = string(cmd.OwnerInformation.BeneficialOwnersStructure)
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

	// REUSE: missing filenames check (same as saveverificationform)
	missing := make([]string, 0, 16)
	container := AzureBlobSecureRootContainerName()
	basePath := businessFormDocumentPath(req.CustomersId)
	if Storage != nil && cmd.OwnerInformation != nil {
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
	if Storage != nil && cmd.OperationsInformation != nil {
		fn := strings.TrimSpace(cmd.OperationsInformation.FinancialInstitutionFormFileName)
		if fn != "" {
			blobName := fmt.Sprintf("%d_%d_%s", req.CustomersId, req.SiteUsersId, fn)
			exists, _ := Storage.CheckFileExists(basePath+blobName, container)
			if !exists {
				missing = append(missing, fn)
			}
		}
	}

	// Prepare JSON strings like .NET DbClient
	jsonData := string(req.Command)
	jsonOwnerValidation, _ := json.Marshal(ownerValidation)
	jsonRegValidation, _ := json.Marshal(regValidation)

	ctx, cancel := withTimeout()
	defer cancel()

	sp := "v3_CustomerRole_BusinessModule_SubmitVerificationForm"
	var detailsStr string
	id, status, errorsStr, err := dbGetSPResult(ctx, sp, &detailsStr, map[string]interface{}{
		"SiteUsersID":                      req.SiteUsersId,
		"CustomersID":                      req.CustomersId,
		"JsonData":                         jsonData,
		"JsonOwnerValidationResult":        string(jsonOwnerValidation),
		"JsonRegistrationValidationResult": string(jsonRegValidation),
		"JsonOwnershipGraph":               jsonOwnershipGraph,
		"JsonOwnershipTree":                jsonOwnershipTree,
		"MissingFilenames":                 strings.Join(missing, ","),
		"AddedBy":                          req.AddedBy,
	})
	if err != nil {
		*res = DbResultRPC{ID: 0, Id: 0, Status: "0", Details: "", Errors: err.Error()}
		return nil
	}

	*res = DbResultRPC{ID: id, Id: id, Status: status, Details: detailsStr, Errors: errorsStr}
	if strings.TrimSpace(res.Errors) == "" {
		res.Status = "1"
	}

	// Success: send significant party emails (if needed)
	if res.Status == "1" && strings.TrimSpace(res.Details) != "" {
		var details SaveFormResponse
		if json.Unmarshal([]byte(res.Details), &details) == nil {
			if details.ListInnerErrors != nil && len(details.ListInnerErrors) == 0 {
				for _, sparty := range details.SignificantParties {
					if strings.TrimSpace(sparty.EmailAddress) == "" || TokenHelper == nil || NotificationsClient == nil {
						continue
					}
					token, tokErr := TokenHelper.GenerateSignificantPartiesJumioToken(req.CustomersId, sparty.Id)
					if tokErr != nil {
						continue
					}
					portalURL := "" // TODO: call v1_CustomerRole_General_GetLicenseesBrandsForSiteUser if needed
					base := portalURL
					if strings.TrimSpace(base) == "" {
						base = req.Domain
					}
					jumioLink := buildJumioLink(base, token)
					_ = NotificationsClient.SendSignificantPartyJumioEmail(sparty.EmailAddress, details.CompanyName, jumioLink, req.Domain)
				}
			}
		}
	}

	return nil
}

// DocumentGroupList RPC method
func (s *BusinessService) DocumentGroupList(req *GetDocumentGroupListRequest, res *BusinessDocumentsResult) error {
	if res == nil {
		return fmt.Errorf("DocumentGroupList: nil response pointer")
	}
	if req == nil {
		*res = BusinessDocumentsResult{
			ID:      0,
			Status:  "0",
			Details: []BusinessDocumentGroup{},
			Errors:  []ErrorItem{{FieldName: "Json", MessageCode: "Invalid_Json"}},
		}
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	sp := "v1_CustomerRole_BusinessModule_GetKYCDocumentList"

	var detailsStr string
	id, _, errorsStr, err := dbGetSPResult(ctx, sp, &detailsStr, map[string]any{
		"CustomersId": req.CustomersId,
		"SiteUsersId": req.SiteUsersId,
	})
	if err != nil {
		*res = BusinessDocumentsResult{
			ID:      0,
			Status:  "0",
			Details: []BusinessDocumentGroup{},
			Errors:  []ErrorItem{{FieldName: "Db", MessageCode: "Unexpected_Error"}},
		}
		return nil
	}

	// ✅ DB errors -> structured errors array
	if strings.TrimSpace(errorsStr) != "" {
		errorsArr := parseLegacyDbErrorString(errorsStr)
		errItems := errorResultFilesToErrorItems(errorsArr)
		if len(errItems) == 0 {
			errItems = []ErrorItem{{FieldName: "Db", MessageCode: strings.TrimSpace(errorsStr)}}
		}
		*res = BusinessDocumentsResult{
			ID:      int(id),
			Status:  "0",
			Details: []BusinessDocumentGroup{},
			Errors:  errItems,
		}
		return nil
	}

	// ✅ Parse detailsStr as FLAT docs list
	flatDocs, ok := parseFlatBusinessDocuments(detailsStr)
	if !ok {
		*res = BusinessDocumentsResult{
			ID:      int(id),
			Status:  "0",
			Details: []BusinessDocumentGroup{},
			Errors:  []ErrorItem{{FieldName: "Details", MessageCode: "Invalid_Json"}},
		}
		return nil
	}

	// ✅ Group by category (helper returns []BusinessDocumentGroup)
	grouped := GroupDocumentsByCategory(flatDocs)

	*res = BusinessDocumentsResult{
		ID:      int(id),
		Status:  "1",
		Details: grouped,
		Errors:  []ErrorItem{}, // ALWAYS []
	}
	return nil
}

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
	_, _, err := UploadToStorage(req.FileName, req.ContentType, req.FileBytes)
	if err != nil {
		*res = DbResultFile{ID: 0, Id: 0, Status: "0", Details: "", Errors: err.Error()}
		return nil
	}

	// 2) Save metadata in DB
	sp := "v1_CustomerRole_BusinessModule_ValidateKYCDocumentUpload"

	var detailsStr string
	id, status, errorsStr, dbErr := dbGetSPResult(
		ctx, sp, &detailsStr,
		"CustomersId", req.CustomersId,
		"SiteUsersId", req.SiteUsersId,
		"Filename", req.FileName,
		"DocumentId", req.DocumentId,
		"CustomersBusinessDocumentsId", 0,
		// "DocumentId", req.DocumentTypeId, // if needed, add to struct and uncomment
		// "CustomersBusinessDocumentsId", 0, // if needed, add to struct and uncomment
	)

	if dbErr != nil {
		*res = DbResultFile{ID: 0, Id: 0, Status: "0", Details: "", Errors: dbErr.Error()}
		return nil
	}

	// 3) Microservice -> gateway safe (RAW JSON string)
	var errorsArr []ErrorResultFile
	if strings.TrimSpace(errorsStr) != "" {
		errorsArr = parseLegacyDbErrorString(errorsStr)
	}
	finalErrors := errorsStr
	if len(errorsArr) > 0 {
		b, _ := json.Marshal(errorsArr)
		finalErrors = string(b)
		status = "0"
	} else {
		status = "1"
	}
	*res = DbResultFile{
		ID:      id,
		Id:      id,
		Status:  status,
		Details: detailsStr, // RAW JSON STRING
		Errors:  finalErrors,
	}
	return nil
}

func (s *BusinessService) SaveVerificationForm(req *SaveVerificationFormRequest, res *SaveVerificationFormResult) error {
	if res == nil {
		return fmt.Errorf("SaveVerificationForm: nil response pointer")
	}

	// Validate request + command
	if errs := ValidateSaveVerificationForm(req); len(errs) > 0 {
		id := int64(0)
		if req != nil {
			id = req.CustomersId
		}
		*res = SaveVerificationFormResult{
			ID:     id,
			Status: "0",
			Details: SaveVerificationFormResponse{
				InnerErrors:     "",
				ListInnerErrors: make([]ErrorItem, 0),
			},
			Errors: errs,
		}
		return nil
	}

	// Parse command payload (keep sub-parts as RawMessage so we don't coerce types)
	var cmd SaveVerificationFormCommandPayload
	if err := json.Unmarshal(req.Command, &cmd); err != nil {
		*res = SaveVerificationFormResult{
			ID:     req.CustomersId,
			Status: "0",
			Details: SaveVerificationFormResponse{
				InnerErrors:     "",
				ListInnerErrors: make([]ErrorItem, 0),
			},
			Errors: []ErrorItem{{FieldName: "command", MessageCode: "Invalid_Json"}},
		}
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	sp := "v3_CustomerRole_BusinessModule_SaveVerificationForm"

	rawOrNull := func(r json.RawMessage) string {
		s := strings.TrimSpace(string(r))
		if s == "" {
			return "null"
		}
		return s
	}

	// ownerValidationResult / registrationValidationResult (placeholder until rules are ported)
	ownerValidation := ResultDto{Id: 0, Errors: []ErrorResultFile{}}
	regValidation := ResultDto{Id: 0, Errors: []ErrorResultFile{}}
	jsonOwnerValidation, _ := json.Marshal(ownerValidation)
	jsonRegValidation, _ := json.Marshal(regValidation)

	// OwnershipGraph + OwnershipTree JSON (best-effort)
	jsonOwnershipGraph := "{}"
	jsonOwnershipTree := "null"
	var ownerInfo VerificationOwnerInformation
	if strings.TrimSpace(string(cmd.OwnerInformation)) != "" && strings.TrimSpace(string(cmd.OwnerInformation)) != "null" {
		_ = json.Unmarshal(cmd.OwnerInformation, &ownerInfo)
	}
	if len(ownerInfo.BeneficialOwnersStructure) > 0 {
		bos := strings.TrimSpace(string(ownerInfo.BeneficialOwnersStructure))
		if bos != "" && bos != "null" {
			jsonOwnershipTree = bos
			var children []OwnershipTreeNode
			_ = json.Unmarshal(ownerInfo.BeneficialOwnersStructure, &children)
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
	container := AzureBlobSecureRootContainerName()
	basePath := businessFormDocumentPath(req.CustomersId)
	if Storage != nil {
		for _, bo := range ownerInfo.IndividualBeneficialOwners {
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

		var opsInfo OperationsInfo
		if strings.TrimSpace(string(cmd.OperationsInformation)) != "" && strings.TrimSpace(string(cmd.OperationsInformation)) != "null" {
			_ = json.Unmarshal(cmd.OperationsInformation, &opsInfo)
			fn := strings.TrimSpace(opsInfo.FinancialInstitutionFormFileName)
			if fn != "" {
				blobName := fmt.Sprintf("%d_%d_%s", req.CustomersId, req.SiteUsersId, fn)
				exists, _ := Storage.CheckFileExists(basePath+blobName, container)
				if !exists {
					missing = append(missing, fn)
				}
			}
		}
	}

	addedBy := strings.TrimSpace(req.AddedBy)
	if addedBy == "" {
		addedBy = strings.TrimSpace(cmd.AddedBy)
	}
	if addedBy == "" {
		addedBy = "adeel"
	}

	// Build JsonData with PascalCase root keys (to match .NET SP payload expectation)
	jsonDataObj := map[string]any{
		"BusinessVerificationStep": cmd.BusinessVerificationStep,
		"RegistrationInformation":  json.RawMessage(rawOrNull(cmd.RegistrationInformation)),
		"OperationsInformation":    json.RawMessage(rawOrNull(cmd.OperationsInformation)),
		"OwnerInformation":         json.RawMessage(rawOrNull(cmd.OwnerInformation)),
		"Terms":                    json.RawMessage(rawOrNull(cmd.Terms)),
		"AccountCurrencySelection": json.RawMessage(rawOrNull(cmd.AccountCurrencySelection)),
		"DocumentsUpload":          json.RawMessage(rawOrNull(cmd.DocumentsUpload)),
		"addedBy":                  addedBy,
	}
	jsonDataBytes, _ := json.MarshalIndent(jsonDataObj, "", "    ")
	jsonData := string(jsonDataBytes)

	// IMPORTANT: SP parameters are PascalCase (and include JsonData + validation/ownership payloads)
	params := map[string]any{
		"SiteUsersID":                      req.SiteUsersId,
		"CustomersID":                      req.CustomersId,
		"BusinessVerificationStep":         cmd.BusinessVerificationStep,
		"AddedBy":                          addedBy,
		"JsonData":                         jsonData,
		"JsonOwnerValidationResult":        string(jsonOwnerValidation),
		"JsonRegistrationValidationResult": string(jsonRegValidation),
		"JsonOwnershipGraph":               jsonOwnershipGraph,
		"JsonOwnershipTree":                jsonOwnershipTree,
		"MissingFilenames":                 strings.Join(missing, ","),
	}

	dbRes, err := ExecSPDbResult(ctx, sp, params)
	if err != nil {
		*res = SaveVerificationFormResult{
			ID:      req.CustomersId,
			Status:  "0",
			Details: SaveVerificationFormResponse{InnerErrors: "", ListInnerErrors: make([]ErrorItem, 0)},
			Errors:  []ErrorItem{{FieldName: "Error", MessageCode: err.Error()}},
		}
		return nil
	}

	// Build wrapper result
	outRes := SaveVerificationFormResult{ID: req.CustomersId, Status: strings.TrimSpace(dbRes.Status), Errors: make([]ErrorItem, 0)}
	// Details is returned as a JSON string from the SP
	details := SaveVerificationFormResponse{ListInnerErrors: make([]ErrorItem, 0)}
	if detailsStr, ok := dbRes.Details.(string); ok {
		detailsStr = strings.TrimSpace(detailsStr)
		if detailsStr != "" && detailsStr != "null" {
			_ = json.Unmarshal([]byte(detailsStr), &details)
		}
	}

	// Merge/normalize SP errors (so caller still receives problems even if details didn't include them)
	if errStr, ok := dbRes.Errors.(string); ok {
		errStr = strings.TrimSpace(errStr)
		if errStr != "" && errStr != "[]" {
			details.InnerErrors = errStr
			norm := normalizeErrorsToSlice(errStr)
			if len(norm) > 0 {
				details.ListInnerErrors = append(details.ListInnerErrors, norm...)
				details.ListInnerErrors = dedupErrorItems(details.ListInnerErrors)
				outRes.Errors = append(outRes.Errors, norm...)
				outRes.Errors = dedupErrorItems(outRes.Errors)
			}
		}
	}

	if details.ListInnerErrors == nil {
		details.ListInnerErrors = make([]ErrorItem, 0)
	}
	if outRes.Errors == nil {
		outRes.Errors = make([]ErrorItem, 0)
	}
	if outRes.Status == "" {
		if len(outRes.Errors) == 0 {
			outRes.Status = "1"
		} else {
			outRes.Status = "0"
		}
	}
	outRes.Details = details
	*res = outRes
	return nil
}

// RPC: "BusinessService.GetVerificationForm" (also exposed via alias "BusinessV3Service" if registered)
func (s *BusinessService) GetVerificationForm(req *GetVerificationFormRequest, res *DbResultRPC) error {
	if res == nil {
		return fmt.Errorf("GetVerificationForm: nil response pointer")
	}
	if req == nil {
		bad := []ErrorResultDto{{ErrorType: "BadRequest", FieldName: "Json", MessageCode: "Invalid_Json"}}
		b, _ := json.Marshal(bad)
		*res = DbResultRPC{ID: 0, Id: 0, Status: "0", Details: "", Errors: string(b)}
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	sp := "V3_CustomerRole_BusinessModule_GetVerificationForm"
	var detailsStr string
	id, status, errorsStr, err := dbGetSPResult(
		ctx,
		sp,
		&detailsStr,
		"SiteUsersID", req.SiteUsersId,
		"CustomersID", req.CustomersId,
	)
	if err != nil {
		*res = DbResultRPC{ID: 0, Id: 0, Status: "0", Details: "", Errors: err.Error()}
		return nil
	}

	// Enrich details like .NET: ListInnerErrors from InnerErrors
	// Enrich details like .NET: ListInnerErrors from InnerErrors + build formData from rawFormData
	if strings.TrimSpace(detailsStr) != "" {
		var dto GetVerificationFormDto
		if json.Unmarshal([]byte(detailsStr), &dto) == nil {

			// 1) innerErrors -> listInnerErrors (always [] not null)
			dto.ListInnerErrors = make([]ErrorResultDto, 0)
			if strings.TrimSpace(dto.InnerErrors) != "" {
				dto.ListInnerErrors = append(dto.ListInnerErrors, ParseInnerErrors(dto.InnerErrors)...)
			}

			// 2) rawFormData -> formData struct
			// If DB didn't send formData or it's null, compute it
			if dto.FormData == nil {
				if fd, ok := ParseRawFormData(dto.RawFormData); ok {
					dto.FormData = fd
				}
			}

			// marshal back into detailsStr (still string for gob-safe DbResultRPC)
			b, _ := json.Marshal(dto)
			detailsStr = string(b)
		}
	}

	*res = DbResultRPC{ID: id, Id: id, Status: status, Details: detailsStr, Errors: errorsStr}
	if strings.TrimSpace(res.Errors) == "" {
		res.Status = "1"
	}
	return nil
}

// RPC: "BusinessV3Service.GetBusinessVerificationStatus" (alias points to BusinessService)
func (s *BusinessService) GetBusinessVerificationStatus(req *DocumentRequest, res *BusinessVerificationStatusResult) error {
	if res == nil {
		return fmt.Errorf("GetBusinessVerificationStatus: nil response pointer")
	}

	// req is *struct{}; nil req is possible in rpc
	if req == nil {
		*res = BusinessVerificationStatusResult{
			ID:     0,
			Status: "0",
			Details: BusinessVerificationDetails{
				RejectedDocuments: RejectedDocuments{},
			},
			Errors: []ErrorItem{{FieldName: "Json", MessageCode: "Invalid_Json"}},
		}
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	sp := "v2_CustomerRole_BusinessModule_GetBusinessVerificationStatus"

	var detailsStr string
	id, status, errorsStr, err := dbGetSPResult(ctx, sp, &detailsStr,
		"CustomersId", req.CustomersId,
		"SiteUsersId", req.SiteUsersId)
	if err != nil {
		*res = BusinessVerificationStatusResult{
			ID:     0,
			Status: "0",
			Details: BusinessVerificationDetails{
				RejectedDocuments: RejectedDocuments{},
			},
			Errors: []ErrorItem{{FieldName: "Error", MessageCode: err.Error()}},
		}
		return nil
	}

	// ✅ Details JSON -> typed struct
	var details BusinessVerificationDetails
	if strings.TrimSpace(detailsStr) != "" {
		if e := json.Unmarshal([]byte(detailsStr), &details); e != nil {
			*res = BusinessVerificationStatusResult{
				ID:     id,
				Status: "0",
				Details: BusinessVerificationDetails{
					RejectedDocuments: RejectedDocuments{},
				},
				Errors: []ErrorItem{{FieldName: "Details", MessageCode: "Invalid_Details_JSON"}},
			}
			return nil
		}
	} else {
		// ensure nested struct isn't nil-ish
		details.RejectedDocuments = RejectedDocuments{}
	}

	// ✅ Errors normalize to []ErrorItem (handles "", "[]", JSON array, legacy text)
	errs := normalizeErrorsToSlice(errorsStr)

	// ✅ Status rule
	statusOut := strings.TrimSpace(status)
	if len(errs) == 0 {
		statusOut = "1"
	} else {
		statusOut = "0"
	}

	*res = BusinessVerificationStatusResult{
		ID:      id,
		Status:  statusOut,
		Details: details,
		Errors:  errs,
	}
	return nil
}

// RPC: "BusinessV3Service.GetDropdownData" (alias points to BusinessService)
func (s *BusinessService) GetDropdownData(req *GetDropdownDataRequest, res *BusinessRegistrationMetadataResult) error {
	if res == nil {
		return fmt.Errorf("GetDropdownData: nil response pointer")
	}
	if req == nil {
		// bad := []ErrorResultDto{{ErrorType: "BadRequest", FieldName: "Json", MessageCode: "Invalid_Json"}}
		// b, _ := json.Marshal(bad)
		*res = BusinessRegistrationMetadataResult{ID: 0, Details: BusinessRegistrationMetadataDetails{}, Status: "0", Errors: []ErrorItem{{FieldName: "Json", MessageCode: "Invalid_Json"}}}
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	sp := "v2_CustomerRole_BusinessModule_GetDropdownData"
	var detailsStr string
	id, status, errorsStr, err := dbGetSPResult(
		ctx,
		sp,
		&detailsStr,
		"CustomersId", req.CustomersId,
		"SiteUsersId", req.SiteUsersId,
	)
	if err != nil {
		*res = BusinessRegistrationMetadataResult{ID: 0, Status: "0", Details: BusinessRegistrationMetadataDetails{}, Errors: []ErrorItem{{FieldName: "Error", MessageCode: err.Error()}}}
		return nil
	}

	errs := normalizeErrorsToSlice(errorsStr)

	// ✅ Status rule
	statusOut := strings.TrimSpace(status)
	if len(errs) == 0 {
		statusOut = "1"
	} else {
		statusOut = "0"
	}
	var details BusinessRegistrationMetadataDetails
	if err := json.Unmarshal([]byte(detailsStr), &details); err != nil {
		details = BusinessRegistrationMetadataDetails{}
	}
	*res = BusinessRegistrationMetadataResult{
		ID:      id,
		Status:  statusOut,
		Details: details,
		Errors:  errs,
	}
	return nil
}

// RPC method: "BusinessService.AddKYCDocumentV3"
func (s *BusinessService) AddKYCDocumentV3(req *AddKYCDocumentRequest, res *DbResultRPC) error {
	if res == nil {
		return fmt.Errorf("AddKYCDocumentV3: nil response pointer")
	}

	if errs := validateAddKycDocument(req); len(errs) > 0 {
		*res = DbResultRPC{ID: 0, Id: 0, Status: "0", Details: "", Errors: toErrorsJSON(errs)}
		return nil
	}

	ctx, cancel := withTimeout()
	defer cancel()

	path := businessKycDocPath(req.CustomersId)

	// DELETE MODE
	if len(req.FileBytes) == 0 {
		if req.CustomersBusinessDocumentsId == nil {
			*res = DbResultRPC{ID: 0, Id: 0, Status: "0", Details: "", Errors: toErrorsJSON([]ErrorResult{{ErrorType: "BadRequest", FieldName: "File", MessageCode: "Required"}})}
			return nil
		}

		spDetails := "v1_CustomerRole_BusinessModule_GetKYCDocumentDetails"
		var detailsStr string
		fmt.Printf("[AddKYCDocumentV3] Calling SP: %s\nParams: %+v\n", spDetails, map[string]interface{}{
			"@CustomersId":                  req.CustomersId,
			"@SiteUsersId":                  req.SiteUsersId,
			"@DocumentId":                   req.DocumentId,
			"@CustomersBusinessDocumentsId": *req.CustomersBusinessDocumentsId,
		})
		_, _, _, err := dbGetSPResult(ctx, spDetails, &detailsStr, map[string]interface{}{
			"@CustomersId":                  req.CustomersId,
			"@SiteUsersId":                  req.SiteUsersId,
			"@DocumentId":                   req.DocumentId,
			"@CustomersBusinessDocumentsId": *req.CustomersBusinessDocumentsId,
		})
		if err != nil {
			*res = DbResultRPC{ID: 0, Id: 0, Status: "0", Details: "", Errors: err.Error()}
			return nil
		}

		var dto struct {
			Details UploadedDocumentDetails `json:"details"`
		}
		var raw any
		_ = json.Unmarshal([]byte(detailsStr), &raw)
		fileName := ""
		var d UploadedDocumentDetails
		if json.Unmarshal([]byte(detailsStr), &d) == nil && d.FileName != "" {
			fileName = d.FileName
		} else {
			if json.Unmarshal([]byte(detailsStr), &dto) == nil && dto.Details.FileName != "" {
				fileName = dto.Details.FileName
			}
		}
		if fileName != "" {
			blobName := businessKycDocName(req.CustomersId, req.SiteUsersId, fileName)
			_ = StorageDeleteIfExists(path+blobName, AzureBlobSecureRootContainerName())
		}

		spDelete := "v1_CustomerRole_BusinessModule_DeleteKYCDocument"
		var delDetails string
		fmt.Printf("[AddKYCDocumentV3] Calling SP: %s\nParams: %+v\n", spDelete, map[string]interface{}{
			"@CustomersId":                  req.CustomersId,
			"@SiteUsersId":                  req.SiteUsersId,
			"@CustomersBusinessDocumentsId": *req.CustomersBusinessDocumentsId,
		})
		delId, delStatus, delErrStr, delErr := dbGetSPResult(ctx, spDelete, &delDetails, map[string]interface{}{
			"@CustomersId":                  req.CustomersId,
			"@SiteUsersId":                  req.SiteUsersId,
			"@CustomersBusinessDocumentsId": *req.CustomersBusinessDocumentsId,
		})
		if delErr != nil {
			*res = DbResultRPC{ID: 0, Id: 0, Status: "0", Details: "", Errors: delErr.Error()}
			return nil
		}
		*res = DbResultRPC{ID: delId, Id: delId, Status: delStatus, Details: delDetails, Errors: delErrStr}
		if strings.TrimSpace(res.Errors) == "" {
			res.Status = "1"
		}
		return nil
	}

	// UPLOAD MODE
	if e := validateFileTypeAndExt(req.ContentType, req.FileName); e != nil {
		*res = DbResultRPC{ID: 0, Id: 0, Status: "0", Details: "", Errors: toErrorsJSON([]ErrorResult{*e})}
		return nil
	}

	spValidate := "v1_CustomerRole_BusinessModule_ValidateKYCDocumentUpload"
	var validateDetails string
	fmt.Printf("[AddKYCDocumentV3] Calling SP: %s\nParams: %+v\n", spValidate, map[string]interface{}{
		"@CustomersId":                  req.CustomersId,
		"@SiteUsersId":                  req.SiteUsersId,
		"@DocumentId":                   req.DocumentId,
		"@Filename":                     req.FileName,
		"@CustomersBusinessDocumentsId": req.CustomersBusinessDocumentsId,
	})
	_, vStatus, vErrorsStr, vErr := dbGetSPResult(ctx, spValidate, &validateDetails, map[string]interface{}{
		"@CustomersId":                  req.CustomersId,
		"@SiteUsersId":                  req.SiteUsersId,
		"@DocumentId":                   req.DocumentId,
		"@Filename":                     req.FileName,
		"@CustomersBusinessDocumentsId": req.CustomersBusinessDocumentsId,
	})
	if vErr != nil {
		*res = DbResultRPC{ID: 0, Id: 0, Status: "0", Details: "", Errors: vErr.Error()}
		return nil
	}
	if vStatus != "1" {
		*res = DbResultRPC{ID: 0, Id: 0, Status: vStatus, Details: validateDetails, Errors: vErrorsStr}
		if strings.TrimSpace(res.Errors) == "" {
			res.Status = "1"
		}
		return nil
	}

	blobName := businessKycDocName(req.CustomersId, req.SiteUsersId, req.FileName)
	ok, storageErrs := StorageSaveOrOverwrite(path+blobName, req.ContentType, req.FileBytes, AzureBlobSecureRootContainerName())
	if !ok {
		*res = DbResultRPC{
			ID: 0, Id: 0, Status: "0", Details: "",
			Errors: toErrorsJSON([]ErrorResult{{ErrorType: "BadRequest", FieldName: "File", MessageCode: "Upload_Failed"}}),
		}
		_ = storageErrs
		return nil
	}

	spComplete := "v1_CustomerRole_BusinessModule_CompleteKYCDocumentUpload"
	var completeDetails string
	fmt.Printf("[AddKYCDocumentV3] Calling SP: %s\nParams: %+v\n", spComplete, map[string]interface{}{
		"@CustomersId":                  req.CustomersId,
		"@SiteUsersId":                  req.SiteUsersId,
		"@DocumentId":                   req.DocumentId,
		"@Note":                         req.Note,
		"@bDocumentUploaded":            true,
		"@Filename":                     req.FileName,
		"@CustomersBusinessDocumentsId": req.CustomersBusinessDocumentsId,
	})
	cId, cStatus, cErrorsStr, cErr := dbGetSPResult(ctx, spComplete, &completeDetails, map[string]interface{}{
		"@CustomersId":                  req.CustomersId,
		"@SiteUsersId":                  req.SiteUsersId,
		"@DocumentId":                   req.DocumentId,
		"@Note":                         req.Note,
		"@bDocumentUploaded":            true,
		"@Filename":                     req.FileName,
		"@CustomersBusinessDocumentsId": req.CustomersBusinessDocumentsId,
	})
	if cErr != nil {
		_ = StorageDeleteIfExists(path+blobName, AzureBlobSecureRootContainerName())
		*res = DbResultRPC{ID: 0, Id: 0, Status: "0", Details: "", Errors: cErr.Error()}
		return nil
	}
	if cStatus != "1" {
		_ = StorageDeleteIfExists(path+blobName, AzureBlobSecureRootContainerName())
	}
	*res = DbResultRPC{ID: cId, Id: cId, Status: cStatus, Details: completeDetails, Errors: cErrorsStr}
	if strings.TrimSpace(res.Errors) == "" {
		res.Status = "1"
	}
	return nil
}
