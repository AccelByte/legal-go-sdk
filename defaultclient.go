// Copyright (c) 2021 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package legal

import (
	"github.com/pkg/errors"
	"net/http"
	"time"

	"github.com/AccelByte/iam-go-sdk"
	"github.com/patrickmn/go-cache"
)

const (
	crucialPolicyVersionPath = "/public/policies/version/allCrucial"
	allAffectedClientID      = "all"

	defaultPolicyVersionCacheTime = 60 * time.Second
	maxBackOffTime                = 60 * time.Second
)

type LegalConfig struct {
	LegalBaseURL                 string
	PublisherNamespace           string
	PolicyVersionRefreshInterval time.Duration
	Debug                        bool
}

type DefaultLegalClient struct {
	legalConfig               *LegalConfig
	policyVersion             map[string][]PolicyVersion
	policyVersionCache        *cache.Cache
	policyVersionRefreshError error
	remotePolicyValidation    func(listPolicyVersion []string, clientID, country, namespace string) (bool, error)
	// for mocking the HTTP call
	httpClient HTTPClient
}

var debug bool

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewDefaultLegalClient creates new Legal DefaultClient
func NewDefaultLegalClient(config *LegalConfig) LegalClient {
	if config.PolicyVersionRefreshInterval <= 0 {
		config.PolicyVersionRefreshInterval = defaultPolicyVersionCacheTime
	}

	client := &DefaultLegalClient{
		legalConfig: config,
		policyVersionCache: cache.New(
			config.PolicyVersionRefreshInterval,
			2*config.PolicyVersionRefreshInterval,
		),
		httpClient: &http.Client{},
	}

	client.remotePolicyValidation = client.remoteValidatePolicyVersion

	debug = config.Debug

	log("NewDefaultClient: debug enabled")

	return client
}

func (client *DefaultLegalClient) StartCachingCrucialLegal() error {
	err := client.getCrucialPolicyVersion()
	if err != nil {
		return logAndReturnErr(
			errors.WithMessage(err, "StartCachingCrucialLegal: unable to get crucial legal"))
	}

	go client.refreshCrucialPolicyVersion()

	log("StartCachingCrucialLegal: caching crucial legal start")

	return nil
}

func (client *DefaultLegalClient) ValidatePolicyVersions(claims *iam.JWTClaims) (bool, error) {
	// Check for affected clientID
	if cachedCrucialPolicyVersion, found := client.policyVersionCache.Get(claims.ClientID); found {
		if !validate(claims.AcceptedPolicyVersion,  cachedCrucialPolicyVersion.([]PolicyVersion), claims.Country, claims.Namespace, client.legalConfig.PublisherNamespace) {
			return false, nil
		}
	}

	// check for all affected clientID
	if cachedCrucialPolicyVersion, found := client.policyVersionCache.Get(allAffectedClientID); found {
		if !validate(claims.AcceptedPolicyVersion, cachedCrucialPolicyVersion.([]PolicyVersion), claims.Country, claims.Namespace, client.legalConfig.PublisherNamespace) {
			return false, nil
		}
	}

	// cache not found, do remoteValidation

	log("remote policy version validation start")
	return client.remotePolicyValidation(claims.AcceptedPolicyVersion, claims.ClientID, claims.Country, claims.Namespace)

}

func (client *DefaultLegalClient) HealthCheck() bool {
	if client.policyVersionRefreshError != nil {
		logErr(client.policyVersionRefreshError, "HealthCheck: error in Policy Version refresh")
		return false
	}

	log("HealthCheck: all OK")

	return true
}

func contains(listOfPolicyVersion []string, targetPolicyVersion string) bool {
	for _, policyVersion := range listOfPolicyVersion {
		if policyVersion == targetPolicyVersion {
			return true
		}
	}

	return false
}

func validate(policyVersions []string, requiredPolicyVersions []PolicyVersion, country, namespace, publisherNamespace string) bool {
	// check if required policy versions is empty, return true
	if len(requiredPolicyVersions) == 0 {
		return true
	}

	// check namespace equal to publisher namespace, if not the same check legal in publisher too
	if namespace != publisherNamespace {
		for _, requiredPolicyVersion := range requiredPolicyVersions {
			if requiredPolicyVersion.Country != country ||
				(requiredPolicyVersion.Namespace != namespace &&
				requiredPolicyVersion.Namespace != publisherNamespace) {
				continue
			}
			if contains(policyVersions, requiredPolicyVersion.PolicyVersionID) {
				continue
			}
			return false
		}

		return true
	}

	// namespace equal to publisher namespace check only the same namespace where the user login
	for _, requiredPolicyVersion := range requiredPolicyVersions {
		if requiredPolicyVersion.Country != country ||
			requiredPolicyVersion.Namespace != namespace {
			continue
		}
		if contains(policyVersions, requiredPolicyVersion.PolicyVersionID) {
			continue
		}
		return false
	}

	return true
}
