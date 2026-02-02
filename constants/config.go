package constants

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Environment string
	PORT        string

	ApplicationSettings struct {
		ProjectName                            string `json:"ProjectName"`
		DefaultDomain                          string `json:"DefaultDomain"`
		EmailVerificationCodeExpiryTimeMinutes int    `json:"EmailVerificationCodeExpiryTimeMinutes"`

		MaxThreads  int `json:"MaxThreads"`
		BatchSize   int `json:"BatchSize"`
		MaxTryCount int `json:"MaxTryCount"`
	} `json:"ApplicationSettings"`

	ConnectionStrings struct {
		DefaultConnection string `json:"DefaultConnection"`
	} `json:"ConnectionStrings"`

	NiumSettings struct {
		BaseUrl                     string `json:"BaseUrl"`
		ClientKey                   string `json:"ClientKey"`
		ClientSecret                string `json:"ClientSecret"`
		UseMockApi                  bool   `json:"UseMockApi"`
		OperationalUsdAccountNumber int64  `json:"OperationalUsdAccountNumber"`
	} `json:"NiumSettings"`

	CrossRiverBankSettings struct {
		RemittanceInformation struct {
			Name                 string `json:"Name"`
			ClientKey            string `json:"ClientKey"`
			AddressLine1         string `json:"AddressLine1"`
			AddressLine2         string `json:"AddressLine2"`
			TownCity             string `json:"TownCity"`
			StateProvince        string `json:"StateProvince"`
			Postcode             string `json:"Postcode"`
			CountryISO2          string `json:"CountryISO2"`
			IdentificationNumber string `json:"IdentificationNumber"`
			IdentificationType   string `json:"IdentificationType"`
		} `json:"RemittanceInformation"`
	} `json:"CrossRiverBankSettings"`

	RendimentoSettings struct {
		BaseUrl         string `json:"BaseUrl"`
		Email           string `json:"Email"`
		Password        string `json:"Password"`
		TotpSecret      string `json:"TotpSecret"`
		PartnerId       int    `json:"PartnerId"`
		SourceAccountId int64  `json:"SourceAccountId"`
		WebhookApiKey   string `json:"WebhookApiKey"`
		UseMockApi      bool   `json:"UseMockApi"`
		IbaneraCNPJ     string `json:"IbaneraCNPJ"`
		NatureCode      string `json:"NatureCode"`
		PurposeCode     struct {
			PIXINT string `json:"PIXINT"`
		} `json:"PurposeCode"`
	} `json:"RendimentoSettings"`

	CorpaySettings struct {
		BaseUrl                     string `json:"BaseUrl"`
		Email                       string `json:"Email"`
		Password                    string `json:"Password"`
		TotpSecret                  string `json:"TotpSecret"`
		PartnerId                   int    `json:"PartnerId"`
		UseMockApi                  bool   `json:"UseMockApi"`
		PendingRetryIntervalMinutes int    `json:"PendingRetryIntervalMinutes"`
		OperationalAccountNumbers   struct {
			USD int64 `json:"USD"`
			AUD int64 `json:"AUD"`
			SGD int64 `json:"SGD"`
			CAD int64 `json:"CAD"`
			EUR int64 `json:"EUR"`
			GBP int64 `json:"GBP"`
		} `json:"OperationalAccountNumbers"`
	} `json:"CorpaySettings"`

	LightnetSettings struct {
		BaseUrl         string `json:"BaseUrl"`
		Email           string `json:"Email"`
		Password        string `json:"Password"`
		TotpSecret      string `json:"TotpSecret"`
		PartnerId       int    `json:"PartnerId"`
		SourceAccountId string `json:"SourceAccountId"`
		WebhookApiKey   string `json:"WebhookApiKey"`
		UseMockApi      bool   `json:"UseMockApi"`
	} `json:"LightnetSettings"`

	ConveraSettings struct {
		BaseUrl         string `json:"BaseUrl"`
		Email           string `json:"Email"`
		Password        string `json:"Password"`
		TotpSecret      string `json:"TotpSecret"`
		PartnerId       int    `json:"PartnerId"`
		SourceAccountId int64  `json:"SourceAccountId"`
		WebhookApiKey   string `json:"WebhookApiKey"`
		UseMockApi      bool   `json:"UseMockApi"`
	} `json:"ConveraSettings"`

	SWIFTCodesSettings struct {
		BaseUrl string `json:"BaseUrl"`
		ApiKey  string `json:"ApiKey"`
	} `json:"SWIFTCodesSettings"`

	Sentry struct {
		Dsn                    string `json:"Dsn"`
		IncludeRequestPayload  bool   `json:"IncludeRequestPayload"`
		SendDefaultPii         bool   `json:"SendDefaultPii"`
		MinimumBreadcrumbLevel string `json:"MinimumBreadcrumbLevel"`
		MinimumEventLevel      string `json:"MinimumEventLevel"`
		AttachStackTrace       bool   `json:"AttachStackTrace"`
		Debug                  bool   `json:"Debug"`
		DiagnosticsLevel       string `json:"DiagnosticsLevel"`
	} `json:"Sentry"`

	// Add this section in appsettings if you want KV config from file
	AzureKeyVault struct {
		Enabled               bool   `json:"bEnabled"`
		KeyVaultName          string `json:"KeyVaultName"`          // "my-vault-name" (NOT full URL)
		SecretNamePrefix      string `json:"SecretNamePrefix"`      // optional
		ReloadIntervalMinutes int    `json:"ReloadIntervalMinutes"` // optional
		TimeoutMs             int    `json:"TimeoutMs"`             // optional
	} `json:"AzureKeyVault"`
}

var Cfg Config

func InitializeEnvironmentVariables(baseDir string) (*Config, error) {
	// Context setup for background task execution
	ctx := context.Background()

	_ = godotenv.Load() // loads .env into OS env (ignore error if optional)

	// Environment detection (default to "development" if not provided)
	env := os.Getenv("ENV")
	if env == "" {
		env = "Development" // Default environment
	}

	// cfg = &Config{Environment: env}
	Cfg.Environment = env

	mergeFile := func(path string) error {
		b, err := os.ReadFile(path)
		if err != nil {
			// optional like .NET
			return nil
		}
		// Strip UTF-8 BOM if present
		b = stripBOM(b)

		return json.Unmarshal(b, &Cfg)
	}

	// 1) appsettings.json
	appSettingsPath := filepath.Join(baseDir, "/configs/appsettings.json")
	if err := mergeFile(appSettingsPath); err != nil {
		return nil, err
	}
	// 2) appsettings.{env}.json
	appSettingsEnvPath := filepath.Join(baseDir, fmt.Sprintf("/configs/appsettings.%s.json", env))
	if err := mergeFile(appSettingsEnvPath); err != nil {
		return nil, err
	}

	// 3) env overrides (things you want to come from environment)
	applyEnvOverrides(&Cfg)
	// 4) keyvault overrides (like .NET provider)
	kvEnabled := Cfg.AzureKeyVault.Enabled || envBool("KEYVAULT_ENABLED", false)
	kvName := firstNonEmpty(
		envStr("KEY_VAULT_NAME", ""),
		envStr("AZURE_KEYVAULT_NAME", ""),
		Cfg.AzureKeyVault.KeyVaultName,
	)

	timeout := 30 * time.Second
	if Cfg.AzureKeyVault.TimeoutMs > 0 {
		timeout = time.Duration(Cfg.AzureKeyVault.TimeoutMs) * time.Millisecond
	}

	if kvEnabled {
		if err := ApplyKeyVaultPlaceholders(ctx, &Cfg, KeyVaultOptions{
			Enabled:   true,
			VaultName: kvName,
			Timeout:   timeout,
		}); err != nil {
			return nil, err
		}
	}

	// 5) Replace placeholders in connection string if you use [[SQL_USERNAME]] etc.
	if strings.EqualFold(
		Cfg.ConnectionStrings.DefaultConnection,
		"FROM_KEY_VAULT",
	) {
		return nil, fmt.Errorf("ConnectionStrings.DefaultConnection was not overridden by Key Vault")
	}
	Cfg.ConnectionStrings.DefaultConnection = replaceConnPlaceholders(Cfg.ConnectionStrings.DefaultConnection)

	// 6) Validate required values
	if err := Cfg.Validate(); err != nil {
		return nil, err
	}

	// CfgJSON, _ := json.Marshal(Cfg)
	// fmt.Println(string(CfgJSON))

	return &Cfg, nil
}

func (c *Config) Validate() error {
	var missing []string

	// Example required checks (adjust to your needs)
	if isPlaceholderOrEmpty(c.ConnectionStrings.DefaultConnection) {
		missing = append(missing, "ConnectionStrings.DefaultConnection")
	}
	if isPlaceholderOrEmpty(c.NiumSettings.BaseUrl) {
		missing = append(missing, "NiumSettings.BaseUrl")
	}
	if isPlaceholderOrEmpty(c.NiumSettings.ClientKey) {
		missing = append(missing, "NiumSettings.ClientKey")
	}
	if isPlaceholderOrEmpty(c.NiumSettings.ClientSecret) {
		missing = append(missing, "NiumSettings.ClientSecret")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing config values (still empty/FROM_KEY_VAULT): %s", strings.Join(missing, ", "))
	}
	return nil
}

func isPlaceholderOrEmpty(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return true
	}
	return strings.EqualFold(s, "FROM_KEY_VAULT")
}

func replaceConnPlaceholders(conn string) string {
	user := envStr("SQL_USERNAME", "")
	pass := envStr("SQL_PASSWORD", "")

	if user != "" {
		conn = strings.ReplaceAll(conn, "[[SQL_USERNAME]]", user)
	}
	if pass != "" {
		conn = strings.ReplaceAll(conn, "[[SQL_PASSWORD]]", pass)
	}
	return conn
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// stripBOM removes UTF-8 BOM (Byte Order Mark) from the beginning of data
func stripBOM(data []byte) []byte {
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return data[3:]
	}
	return data
}
