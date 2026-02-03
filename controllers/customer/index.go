package customer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

// DocumentGroupList RPC method
func (s *BusinessService) DocumentGroupList(req *GetDocumentGroupListRequest, res *DbResultRPC) error {
	if res == nil {
		return fmt.Errorf("DocumentGroupList: nil response pointer")
	}
	if req == nil {
		*res = DbResultRPC{ID: 0, Id: 0, Status: "0", Details: "", Errors: `[{"errorType":"BadRequest","fieldName":"Json","messageCode":"Invalid_Json"}]`}
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	sp := "v1_CustomerRole_BusinessModule_GetKYCDocumentList"
	var detailsStr string
	id, status, errorsStr, err := dbGetSPResult(ctx, sp, &detailsStr, map[string]interface{}{
		"CustomersId": req.CustomersId,
		"SiteUsersId": req.SiteUsersId,
	})
	if err != nil {
		*res = DbResultRPC{ID: 0, Id: 0, Status: "0", Details: "", Errors: err.Error()}
		return nil
	}

	if strings.TrimSpace(errorsStr) == "" {
		status = "1"
	}

	groupedJSON := detailsStr
	if strings.TrimSpace(detailsStr) != "" {
		if out, ok := buildGroupsFromDocumentList(detailsStr); ok {
			groupedJSON = out
		}
	}

	       // Parse errors to array of objects if not empty
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
	       *res = DbResultRPC{
		       ID:      id,
		       Id:      id,
		       Status:  status,
		       Details: groupedJSON,
		       Errors:  finalErrors,
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
