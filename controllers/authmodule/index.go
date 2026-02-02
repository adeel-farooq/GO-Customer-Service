package authmodule

import (
	"context"
	"fmt"
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
