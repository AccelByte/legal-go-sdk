// Copyright (c) 2021 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package legal

import (
	"encoding/json"
	"github.com/patrickmn/go-cache"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/pkg/errors"
)

// nolint: funlen
func (client *DefaultLegalClient) remoteValidatePolicyVersion(listPolicyVersion []string, clientID, country, namespace string) (bool, error) {
	req, err := http.NewRequest("GET", client.legalConfig.LegalBaseURL + crucialPolicyVersionPath, nil)
	if err != nil {
		return false, errors.Wrap(err, "getCrucialPolicyVersion: unable to create new Crucial policy request")
	}

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = maxBackOffTime

	var responseStatusCode int

	var responseBodyBytes []byte

	//nolint: dupl
	err = backoff.Retry(
		func() error {
			var e error

			resp, e := client.httpClient.Do(req)
			if e != nil {
				return backoff.Permanent(e)
			}
			defer resp.Body.Close()

			responseStatusCode = resp.StatusCode
			if resp.StatusCode >= http.StatusInternalServerError {
				return errors.Errorf("getCrucialPolicyVersion: endpoint returned status code : %v", responseStatusCode)
			}

			responseBodyBytes, e = ioutil.ReadAll(resp.Body)
			if e != nil {
				return errors.Wrap(e, "getCrucialPolicyVersion: unable to read response body")
			}

			return nil
		},
		b,
	)

	if err != nil {
		return false, errors.Wrap(err, "getCrucialPolicyVersion: unable to do HTTP request to get crucial policy version")
	}

	if responseStatusCode != http.StatusOK {
		return false, errors.Errorf("getCrucialPolicyVersion: unable to get crucial policy version: error code : %d, error message : %s",
			responseStatusCode, string(responseBodyBytes))
	}

	var getCrucialPolicyVersionResponse CrucialPolicyVersionResponse

	err = json.Unmarshal(responseBodyBytes, &getCrucialPolicyVersionResponse)
	if err != nil {
		return false, errors.Wrap(err, "getCrucialPolicyVersion: unable to unmarshal response body")
	}

	if getCrucialPolicyVersionResponse.AffectedClient == nil {
		return true, nil
	}

	// cache the client id result from remote call
	for clientID, affectedPolicyVersion := range getCrucialPolicyVersionResponse.AffectedClient {
		client.policyVersionCache.Set(clientID, affectedPolicyVersion, cache.DefaultExpiration)
	}

	// Check for affected clientID
	if !validate(listPolicyVersion, getCrucialPolicyVersionResponse.AffectedClient[clientID], country, namespace, client.legalConfig.PublisherNamespace) {
		return false, nil
	}

	if !validate(listPolicyVersion, getCrucialPolicyVersionResponse.AffectedClient[allAffectedClientID], country, namespace, client.legalConfig.PublisherNamespace) {
		return false, nil
	}

	// all policy versions is accepted, user eligible
	log("all crucial policy version accepted")

	return true, nil
}

func (client *DefaultLegalClient) getCrucialPolicyVersion() error {
	req, err := http.NewRequest("GET", client.legalConfig.LegalBaseURL + crucialPolicyVersionPath, nil)
	if err != nil {
		return errors.Wrap(err, "getCrucialPolicyVersion: unable to create new Crucial policy request")
	}

	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = maxBackOffTime

	var responseStatusCode int

	var responseBodyBytes []byte

	//nolint: dupl
	err = backoff.Retry(
		func() error {
			var e error

			resp, e := client.httpClient.Do(req)
			if e != nil {
				return backoff.Permanent(e)
			}
			defer resp.Body.Close()

			responseStatusCode = resp.StatusCode
			if resp.StatusCode >= http.StatusInternalServerError {
				return errors.Errorf("getCrucialPolicyVersion: endpoint returned status code : %v", responseStatusCode)
			}

			responseBodyBytes, e = ioutil.ReadAll(resp.Body)
			if e != nil {
				return errors.Wrap(e, "getCrucialPolicyVersion: unable to read response body")
			}

			return nil
		},
		b,
	)

	if err != nil {
		return errors.Wrap(err, "getCrucialPolicyVersion: unable to do HTTP request to get crucial policy version")
	}

	if responseStatusCode != http.StatusOK {
		return errors.Errorf("getCrucialPolicyVersion: unable to get crucial policy version: error code : %d, error message : %s",
			responseStatusCode, string(responseBodyBytes))
	}

	var getCrucialPolicyVersionResponse CrucialPolicyVersionResponse

	err = json.Unmarshal(responseBodyBytes, &getCrucialPolicyVersionResponse)
	if err != nil {
		return errors.Wrap(err, "getCrucialPolicyVersion: unable to unmarshal response body")
	}

	client.policyVersion = getCrucialPolicyVersionResponse.AffectedClient

	for clientID, affectedPolicyVersion := range getCrucialPolicyVersionResponse.AffectedClient {
		client.policyVersionCache.Set(clientID, affectedPolicyVersion, cache.DefaultExpiration)
	}

	return nil
}

func (client *DefaultLegalClient) refreshCrucialPolicyVersion() {
	backOffTime := time.Second
	time.Sleep(client.legalConfig.PolicyVersionRefreshInterval)

	for {
		client.policyVersionRefreshError = client.getCrucialPolicyVersion()
		if client.policyVersionRefreshError != nil {
			time.Sleep(backOffTime)

			if backOffTime < maxBackOffTime {
				backOffTime *= 2
			}

			continue
		}

		backOffTime = time.Second
		time.Sleep(client.legalConfig.PolicyVersionRefreshInterval)
	}
}