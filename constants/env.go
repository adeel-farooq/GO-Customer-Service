package constants

import (
	"os"
	"strconv"
	"strings"
)

func applyEnvOverrides(cfg *Config) {
	// Connection string overrides (optional)
	if v := envStr("DEFAULT_CONNECTION", ""); v != "" {
		cfg.ConnectionStrings.DefaultConnection = v
	}
	if v := envStr("PORT", ""); v != "" {
		cfg.PORT = v
	}

	// NIUM overrides (optional)
	if v := envStr("NIUM_BASE_URL", ""); v != "" {
		cfg.NiumSettings.BaseUrl = v
	}
	if v := envStr("NIUM_CLIENT_KEY", ""); v != "" {
		cfg.NiumSettings.ClientKey = v
	}
	if v := envStr("NIUM_CLIENT_SECRET", ""); v != "" {
		cfg.NiumSettings.ClientSecret = v
	}
	if v := envStr("NIUM_USE_MOCK_API", ""); v != "" {
		cfg.NiumSettings.UseMockApi = strings.EqualFold(v, "true") || v == "1"
	}

	// Sentry override (optional)
	if v := envStr("SENTRY_DSN", ""); v != "" {
		cfg.Sentry.Dsn = v
	}

	// Customer registration overrides (optional)
	if v := envStr("DEFAULT_DOMAIN", ""); v != "" {
		cfg.ApplicationSettings.DefaultDomain = v
	}
	if v := envStr("EMAIL_VERIFICATION_CODE_EXPIRY_MINUTES", ""); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.ApplicationSettings.EmailVerificationCodeExpiryTimeMinutes = n
		}
	}

	// KeyVault overrides (optional)
	if v := envStr("KEY_VAULT_NAME", ""); v != "" {
		cfg.AzureKeyVault.KeyVaultName = v
	}
	if v := envStr("KEYVAULT_ENABLED", ""); v != "" {
		cfg.AzureKeyVault.Enabled = strings.EqualFold(v, "true") || v == "1"
	}
	if v := envStr("KEYVAULT_TIMEOUT_MS", ""); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.AzureKeyVault.TimeoutMs = n
		}
	}
}

func envStr(key, def string) string {
	if v := os.Getenv(key); strings.TrimSpace(v) != "" {
		return v
	}
	return def
}

func envBool(key string, def bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	return strings.EqualFold(v, "true") || v == "1" || strings.EqualFold(v, "yes") || strings.EqualFold(v, "True")
}
