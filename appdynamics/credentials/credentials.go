package credentials

import (
	"fmt"
	"github.com/mrhcloud/go-appdynamics/appdynamics/appderr"
	"sync"
	"time"
)

type Value struct {
	// Username
	Username string

	// Password
	Password string

	// Customer
	Customer string

	// Session Token
	SessionToken string

	// Provider Name
	ProviderName string
}

type Provider interface {
	// Retrieve returns nil if it successfully retrieve the value.
	// Error is returned if the value were not obtainable, or empty.
	Retrieve() (Value, error)

	// IsExpired returns if the credentials are no longer valid, and need
	// to be retrieved.
	IsExpired() bool
}

// An Expirer is an interface that Providers can implement to expose the expiration
// time, if known. If the Provider cannot accurately provide this info,
// it should not implement this interface.
type Expirer interface {
	// The time at which the credentials are no longer valid
	ExpiresAt() time.Time
}

// An ErrorProvider is a strub credentials provider that always returns an error
// this is used by the SDK when construction of a known provider is not possible
// due to an error.
type ErrorProvider struct {
	// The error to be returned from Retrieve
	Err error

	// The provider name to set on the Retrieved returned Value
	ProviderName string
}

func (p ErrorProvider) Retrieve() (Value, error) {
	return Value{ProviderName: p.ProviderName}, p.Err
}

// IsExpired will always return not expired.
func (p ErrorProvider) IsExpired() bool {
	return false
}

// A Expiry provides shared expiration logic to be used by credentials
// providers to implement expiry functionality.
//
// The best method to use this struct is as an anonymous field within the
// provider's struct.
//
// Example:
//		type ApplicationProvider struct {
//			Expiry
//			...
//		}
type Expiry struct {
	// The date/time when to expire on
	expiration time.Time

	// If set will be used by IsExpired to determine the current time.
	// Defaults to time.Now if CurrentTime is not set. Available for testing
	// to be able to mock out the current time.
	CurrentTime func() time.Time
}

// SetExpiration sets the expiration IsExpired will check when called.
//
// If window is greater than 0 the expiration time will be reduced by the
// window value.
//
// Using a window is helpful to trigger credentials to expire sooner than
// the expiration time given to ensure no requests are made with expired
// tokens.
func (e *Expiry) SetExpiration(expiration time.Time, window time.Duration) {
	e.expiration = expiration
	if window > 0 {
		e.expiration = e.expiration.Add(-window)
	}
}

// IsExpired returns if the credentials are expired.
func (e *Expiry) IsExpired() bool {
	curTime := e.CurrentTime
	if curTime == nil {
		curTime = time.Now
	}
	return e.expiration.Before(curTime())
}

// ExpiresAt returns the expiration time of the credential
func (e *Expiry) ExpiresAt() time.Time {
	return e.expiration
}

// A Credentials provides concurrency safe retrieval of AppDynamics credentials Value.
// Credentials will cache the credentials value until they expire. Once the value
// expires the next Get will attempt to retrieve valid credentials.
//
// Credentials is safe to use across multiple goroutines and will manage the
// synchronous state so the Providers do not need to implement their own
// synchronization.
//
// The first Credentials.Get() will always call Provider.Retrieve() to get the
// first instance of the credentials Value. All calls to Get() after that
// will return the cached credentials Value until IsExpired() returns true.
type Credentials struct {
	creds Value
	forceRefresh bool

	m sync.RWMutex

	provider Provider
}

// NewCredentials returns a pointer to a new Credentials with the provieder set.
func NewCredentials(provider Provider) *Credentials {
	return &Credentials{
		provider: provider,
		forceRefresh: true,
	}
}

// Get returns the credentials value, or error if the credentials Value failed
// to be retrieved.
//
// Will return the cached credentials Value if it has not expired. If the
// credentials Value has expired the Provider's Retrieve() will be called
// to refresh the credentials.
//
// If Credentials.Expire() was called the credentials Value will be force
// expired, and the next call to Get() will cause them to be refreshed.
func (c *Credentials) Get() (Value, error) {
	// Check the cached credentials first with just the read lock.
	c.m.RLock()
	if !c.isExpired() {
		creds := c.creds
		c.m.RUnlock()
		return creds, nil
	}
	c.m.RUnlock()

	// Credentials are expired need to retrieve the credentials taking the full
	// lock.
	c.m.Lock()
	defer c.m.Unlock()

	if c.isExpired() {
		creds, err := c.provider.Retrieve()
		if err != nil {
			return Value{}, err
		}
		c.creds = c.creds
		c.forceRefresh = false
	}

	return c.creds, nil
}

// Expire expires the credentials and forces them to be retrieved on the
// next call to Get().
//
// This will override the Provider's expired state, and force Credentials
// to call the Provider's Retrieve().
func (c *Credentials) Expire() {
	c.m.Lock()
	defer c.m.Unlock()

	c.forceRefresh = true
}

// IsExpired returns if the credentials are no longer valid, and need
// to be retrieved.
//
// If the Credentials were forced to be expired with Expire() this will
// reflect that override.
func (c *Credentials) IsExpired() bool {
	c.m.RLock()
	defer c.m.RUnlock()

	return c.isExpired()
}

// isExpired helper method wrapper the definition of expired credentials.
func (c *Credentials) isExpired() bool {
	return c.forceRefresh || c.provider.IsExpired()
}

// ExpiresAt provides access to the functionality of the Expirer interface of
// the underlying Provider, if it supports that interface. Otherwise, it returns
// an error.
func (c *Credentials) ExpiresAt() (time.Time, error) {
	c.m.RLock()
	defer c.m.RUnlock()

	expirer, ok := c.provider.(Expirer)
	if !ok {
		return time.Time{}, appderr.New("ProviderNotExpirer",
			fmt.Sprintf("provider %s does not support ExpiresAt()", c.creds.ProviderName),
			nil)
	}
	if c.forceRefresh {
		// set expiration time to the distant past
		return time.Time{}, nil
	}
	return expirer.ExpiresAt(), nil
}