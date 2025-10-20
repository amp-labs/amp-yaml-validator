package catalog

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
)

// ErrNotSupported indicates that the catalog does not support object enumeration
var ErrNotSupported = errors.New("object enumeration not supported by catalog")

// CatalogProvider defines an interface for accessing provider catalog information.
// This abstraction allows dependency injection for testing.
type CatalogProvider interface {
	// GetProviderInfo retrieves provider information by name
	GetProviderInfo(ctx context.Context, providerName string) (*providers.ProviderInfo, error)

	// GetProviderSupport retrieves provider capabilities
	GetProviderSupport(ctx context.Context, providerName string) (*providers.Support, error)

	// GetModuleInfo retrieves module information for a given provider and module ID
	GetModuleInfo(ctx context.Context, providerName string, moduleID string) (*providers.ModuleInfo, error)

	// ListObjects returns a list of known object names for a provider/module.
	// Returns ErrNotSupported if the catalog does not expose object schemas.
	ListObjects(ctx context.Context, providerName string, moduleID string) ([]string, error)

	// Ping checks if the catalog is accessible
	Ping(ctx context.Context) error
}

// DefaultCatalogProvider implements CatalogProvider using the connectors package catalog.
// It caches the catalog on first access and is thread-safe.
type DefaultCatalogProvider struct {
	catalog *providers.CatalogWrapper
	once    sync.Once
	err     error
}

// NewDefaultCatalogProvider creates a new default catalog provider that uses
// the providers.ReadCatalog() function to load provider information.
func NewDefaultCatalogProvider() CatalogProvider {
	return &DefaultCatalogProvider{}
}

// loadCatalog loads the catalog once using sync.Once for thread-safe initialization.
func (p *DefaultCatalogProvider) loadCatalog() {
	p.once.Do(func() {
		p.catalog, p.err = providers.ReadCatalog()
	})
}

// GetProviderInfo retrieves provider information by name.
func (p *DefaultCatalogProvider) GetProviderInfo(ctx context.Context, providerName string) (*providers.ProviderInfo, error) {
	p.loadCatalog()

	if p.err != nil {
		return nil, fmt.Errorf("failed to load catalog: %w", p.err)
	}

	info, ok := p.catalog.Catalog[providerName]
	if !ok {
		return nil, fmt.Errorf("provider %s not found in catalog", providerName)
	}

	return &info, nil
}

// GetProviderSupport retrieves provider capabilities.
func (p *DefaultCatalogProvider) GetProviderSupport(ctx context.Context, providerName string) (*providers.Support, error) {
	info, err := p.GetProviderInfo(ctx, providerName)
	if err != nil {
		return nil, err
	}

	return &info.Support, nil
}

// GetModuleInfo retrieves module information for a given provider and module ID.
func (p *DefaultCatalogProvider) GetModuleInfo(ctx context.Context, providerName string, moduleID string) (*providers.ModuleInfo, error) {
	info, err := p.GetProviderInfo(ctx, providerName)
	if err != nil {
		return nil, err
	}

	if info.Modules == nil {
		return nil, fmt.Errorf("provider %s does not have modules", providerName)
	}

	// Direct lookup without fallback
	mods := *info.Modules

	mod, ok := mods[common.ModuleID(moduleID)]
	if !ok {
		return nil, fmt.Errorf("module %s not found for provider %s", moduleID, providerName)
	}

	return &mod, nil
}

// ListObjects returns a list of known object names for a provider/module.
// Currently not supported by the connectors catalog - returns ErrNotSupported.
func (p *DefaultCatalogProvider) ListObjects(ctx context.Context, providerName string, moduleID string) ([]string, error) {
	// The connectors package does not currently expose object schemas in the catalog.
	// This is a best-effort implementation that returns ErrNotSupported to allow
	// graceful degradation in validators.
	return nil, ErrNotSupported
}

// Ping checks if the catalog is accessible by attempting to load it.
func (p *DefaultCatalogProvider) Ping(ctx context.Context) error {
	p.loadCatalog()
	return p.err
}

// MockCatalogProvider is a mock implementation of CatalogProvider for testing.
type MockCatalogProvider struct {
	catalog map[string]providers.ProviderInfo
	// objects maps "provider:module" keys to lists of valid object names
	objects map[string][]string
}

// NewMockCatalogProvider creates a new mock catalog provider with the given catalog data.
func NewMockCatalogProvider(catalog map[string]providers.ProviderInfo) CatalogProvider {
	return &MockCatalogProvider{
		catalog: catalog,
		objects: make(map[string][]string),
	}
}

// NewMockCatalogProviderWithObjects creates a new mock catalog provider with catalog data and object lists.
func NewMockCatalogProviderWithObjects(catalog map[string]providers.ProviderInfo, objects map[string][]string) CatalogProvider {
	return &MockCatalogProvider{
		catalog: catalog,
		objects: objects,
	}
}

// GetProviderInfo retrieves provider information from the mock catalog.
func (m *MockCatalogProvider) GetProviderInfo(ctx context.Context, providerName string) (*providers.ProviderInfo, error) {
	info, ok := m.catalog[providerName]
	if !ok {
		return nil, fmt.Errorf("provider %s not found in catalog", providerName)
	}

	return &info, nil
}

// GetProviderSupport retrieves provider capabilities from the mock catalog.
func (m *MockCatalogProvider) GetProviderSupport(ctx context.Context, providerName string) (*providers.Support, error) {
	info, err := m.GetProviderInfo(ctx, providerName)
	if err != nil {
		return nil, err
	}

	return &info.Support, nil
}

// GetModuleInfo retrieves module information from the mock catalog.
func (m *MockCatalogProvider) GetModuleInfo(ctx context.Context, providerName string, moduleID string) (*providers.ModuleInfo, error) {
	info, err := m.GetProviderInfo(ctx, providerName)
	if err != nil {
		return nil, err
	}

	if info.Modules == nil {
		return nil, fmt.Errorf("provider %s does not have modules", providerName)
	}

	// Direct lookup without fallback
	mods := *info.Modules

	mod, ok := mods[common.ModuleID(moduleID)]
	if !ok {
		return nil, fmt.Errorf("module %s not found for provider %s", moduleID, providerName)
	}

	return &mod, nil
}

// ListObjects returns a list of known object names for a provider/module from the mock catalog.
// The mock can optionally provide object lists for testing. If not provided, returns ErrNotSupported.
func (m *MockCatalogProvider) ListObjects(ctx context.Context, providerName string, moduleID string) ([]string, error) {
	if m.objects == nil {
		return nil, ErrNotSupported
	}

	// Create key for lookup
	key := providerName
	if moduleID != "" {
		key = fmt.Sprintf("%s:%s", providerName, moduleID)
	}

	objects, ok := m.objects[key]
	if !ok {
		return nil, ErrNotSupported
	}

	return objects, nil
}

// Ping always returns nil for the mock catalog provider.
func (m *MockCatalogProvider) Ping(ctx context.Context) error {
	return nil
}
