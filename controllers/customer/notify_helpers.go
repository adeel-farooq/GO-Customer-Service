package customer

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type SignificantPartyTokenHelper interface {
	GenerateSignificantPartiesJumioToken(customersId, significantPartyId int64) (string, error)
}

type HMACTokenHelper struct {
	Secret string
}

func (h HMACTokenHelper) GenerateSignificantPartiesJumioToken(customersId, significantPartyId int64) (string, error) {
	if strings.TrimSpace(h.Secret) == "" {
		return "", fmt.Errorf("token secret missing")
	}
	payload := fmt.Sprintf("%d:%d:%d", customersId, significantPartyId, time.Now().Unix())
	mac := hmac.New(sha256.New, []byte(h.Secret))
	_, _ = mac.Write([]byte(payload))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	token := base64.RawURLEncoding.EncodeToString([]byte(payload)) + "." + sig
	return token, nil
}

type SignificantPartyNotifier interface {
	SendSignificantPartyJumioEmail(toEmail, companyName, jumioLink, domain string) error
}

type NoopNotifier struct{}

func (NoopNotifier) SendSignificantPartyJumioEmail(toEmail, companyName, jumioLink, domain string) error {
	return nil
}

func EnsureHTTP(u string) string {
	u = strings.TrimSpace(u)
	if u == "" {
		return u
	}
	if strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://") {
		return u
	}
	return "https://" + u
}

func BuildJumioLink(baseURL, token string) string {
	baseURL = EnsureHTTP(baseURL)
	baseURL = strings.TrimRight(baseURL, "/")
	return baseURL + "/idverification?token=" + url.QueryEscape(token)
}
