package customer

import (
	"context"
	"fmt"
	"log"
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
