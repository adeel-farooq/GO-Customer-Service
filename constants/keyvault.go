package constants

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
)

const FromKeyVaultPlaceholder = "FROM_KEY_VAULT"

type KeyVaultOptions struct {
	Enabled   bool
	VaultName string
	Timeout   time.Duration
}

// ApplyKeyVaultPlaceholders fetches ONLY secrets for fields that still equal "FROM_KEY_VAULT".
func ApplyKeyVaultPlaceholders(ctx context.Context, cfg *Config, opts KeyVaultOptions) error {
	if !opts.Enabled {
		return nil
	}
	if strings.TrimSpace(opts.VaultName) == "" {
		return fmt.Errorf("keyvault enabled but VaultName is empty")
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 30 * time.Second
	}

	client, err := initializeKeyVaultClient(opts.VaultName)
	if err != nil {
		return err
	}

	// 1) Find all config paths where value == FROM_KEY_VAULT
	paths := findFromKeyVaultPaths(cfg)
	if len(paths) == 0 {
		return nil
	}

	// 2) Fetch and apply each secret to that config path
	patch := make(map[string]any)

	prefix := strings.TrimSpace(cfg.AzureKeyVault.SecretNamePrefix)
	for _, path := range paths {
		secretName := strings.ReplaceAll(path, ".", "--")
		if prefix != "" {
			secretName = prefix + "-" + secretName
		}

		val := getSecret(ctx, client, secretName, opts.Timeout)
		// fmt.Printf("KeyVault fetch for secret '%s' returned value %s\n", secretName, val)
		if val == "" {
			// leave placeholder as-is; Validate() will fail if required
			continue
		}

		setNestedValue(patch, path, val)
	}

	// 3) Merge patch into cfg via JSON roundtrip
	b, _ := json.Marshal(patch)
	if err := json.Unmarshal(b, cfg); err != nil {
		return fmt.Errorf("keyvault merge failed: %w", err)
	}

	return nil
}

// findFromKeyVaultPaths walks cfg using json tags and returns dot-paths where value == "FROM_KEY_VAULT".
func findFromKeyVaultPaths(cfg *Config) []string {
	var out []string
	rv := reflect.ValueOf(cfg).Elem()
	rt := rv.Type()

	var walk func(v reflect.Value, t reflect.Type, prefix string)
	walk = func(v reflect.Value, t reflect.Type, prefix string) {
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.PkgPath != "" { // unexported
				continue
			}

			tag := f.Tag.Get("json")
			if tag == "" || tag == "-" {
				continue
			}
			name := strings.Split(tag, ",")[0]
			if name == "" {
				continue
			}

			path := name
			if prefix != "" {
				path = prefix + "." + name
			}

			fv := v.Field(i)

			// Follow pointers
			for fv.Kind() == reflect.Pointer {
				if fv.IsNil() {
					// nil pointer can't be "FROM_KEY_VAULT", skip
					return
				}
				fv = fv.Elem()
			}

			switch fv.Kind() {
			case reflect.Struct:
				// Recurse into nested structs (e.g. ConnectionStrings, NiumSettings, etc)
				walk(fv, fv.Type(), path)

			case reflect.String:
				if fv.String() == FromKeyVaultPlaceholder {
					out = append(out, path)
				}
			}
		}
	}

	walk(rv, rt, "")
	return out
}

// setNestedValue sets patch["A"]["B"]["C"] = value for path "A.B.C".
func setNestedValue(root map[string]any, path string, value any) {
	parts := strings.Split(path, ".")
	cur := root
	for i := 0; i < len(parts)-1; i++ {
		k := parts[i]
		next, ok := cur[k]
		if !ok {
			m := make(map[string]any)
			cur[k] = m
			cur = m
			continue
		}
		m, ok := next.(map[string]any)
		if !ok {
			m = make(map[string]any)
			cur[k] = m
		}
		cur = m
	}
	cur[parts[len(parts)-1]] = value
}

func ApplyKeyVault(ctx context.Context, cfg *Config, opts KeyVaultOptions) error {
	if !opts.Enabled {
		return nil
	}
	if strings.TrimSpace(opts.VaultName) == "" {
		return fmt.Errorf("keyvault enabled but VaultName is empty")
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 30 * time.Second
	}

	client, err := initializeKeyVaultClient(opts.VaultName)
	if err != nil {
		return err
	}

	// Build a dynamic object tree from secrets, then merge into cfg.
	values := map[string]any{}

	pager := client.NewListSecretsPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("keyvault list secrets failed: %w", err)
		}

		for _, item := range page.Value {
			// SecretItem doesn't have Name; parse from ID
			if item.ID == nil {
				continue
			}

			secretName := item.ID.Name()
			if secretName == "" {
				continue
			}

			secretValue := getSecret(ctx, client, secretName, opts.Timeout)
			if secretValue == "" {
				continue
			}

			// .NET convention: `--` means nested config key
			key := strings.ReplaceAll(secretName, "--", ".")
			key = strings.ReplaceAll(key, fmt.Sprintf("%s-", cfg.AzureKeyVault.SecretNamePrefix), "")
			// fmt.Printf("key: %s\n", key)                 /* testing */
			// fmt.Printf("secretValue: %s\n", secretValue) /* testing */
			setNestedValue(values, key, secretValue)
		}
	}

	b, _ := json.Marshal(values)
	if err := json.Unmarshal(b, cfg); err != nil {
		return fmt.Errorf("keyvault merge into config failed: %w", err)
	}
	return nil
}

// func setNestedValue(m map[string]any, key string, val string) {
// 	parts := strings.Split(key, ".")
// 	cur := m
// 	for i := 0; i < len(parts)-1; i++ {
// 		p := parts[i]
// 		next, ok := cur[p]
// 		if !ok {
// 			nm := map[string]any{}
// 			cur[p] = nm
// 			cur = nm
// 			continue
// 		}
// 		asMap, ok := next.(map[string]any)
// 		if !ok {
// 			// if conflict, overwrite with map
// 			nm := map[string]any{}
// 			cur[p] = nm
// 			cur = nm
// 			continue
// 		}
// 		cur = asMap
// 	}
// 	cur[parts[len(parts)-1]] = val
// }

func initializeKeyVaultClient(vaultName string) (*azsecrets.Client, error) {
	vaultURL := fmt.Sprintf("https://%s.vault.azure.net/", vaultName)

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		fmt.Printf("initializeKeyVaultClient(keyvault.go) failed to create Azure credential: %v", err)
		return nil, fmt.Errorf("failed to create Azure credential: %w", err)
	}
	client, err := azsecrets.NewClient(vaultURL, cred, nil)
	if err != nil {
		fmt.Printf("initializeKeyVaultClient(keyvault.go) failed to create Key Vault client: %v", err)
		return nil, fmt.Errorf("failed to create Key Vault client: %w", err)
	}
	return client, nil
}

// optional semantics: return "" if missing/fails
func getSecret(ctx context.Context, client *azsecrets.Client, secretName string, timeout time.Duration) string {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	resp, err := client.GetSecret(ctxWithTimeout, secretName, "", nil)
	if err == nil && resp.Value != nil {
		return *resp.Value
	}
	return ""
}
