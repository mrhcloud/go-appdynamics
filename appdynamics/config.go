package appdynamics

import (
	"github.com/mrhcloud/go-appdynamics/appdynamics/credentials"
	"net/http"
	"time"
)

const UseServiceDefaultRetries = -1

type RequestRetryer interface{}

type Config struct {
	// Enables verbose error printing of all credential chain errors.
	// Should be used when wanting to see all errors while attempting to
	// retrieve credentials
	CredentialsChainVerboseErrors *bool

	// The credentials object to use when signing requests. Defaults to a
	// chain of credential providers to search for credentials in environment
	// variables or shared credential file
	Credentials *credentials.Credentials

	Endpoint *string

	EnforceShouldRetryCheck *bool

	DisableSSL *bool

	HttpClient *http.Client

	LogLevel *LogLevelType

	Logger Logger

	MaxRetries *int

	Retryer RequestRetryer

	DisableParamValidation *bool

	SleepDelay func(time.Duration)

	DisableRestProtocolURICleaning *bool
}