// Copyright (c) 2021 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package legal

import (
	"bytes"
	"github.com/AccelByte/iam-go-sdk"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
)

const (
	testClientID = "testClientID"
	testClientIDA = "testClientID_A"
	policyVersionA = "policyVersionA"
	countryA = "countryA"
	namespaceA = "namespaceA"
	policyVersionB = "policyVersionB"
	countryB = "countryB"
	namespaceB = "namespaceB"
	policyVersionC = "policyVersionC"
	policyVersionD = "policyVersionD"
	policyVersionE = "policyVersionE"
	policyVersionF = "policyVersionF"
	affectedClientTest = `{
   "affectedClient":{
      "all":[
         {
            "policyVersionId":"policyVersionC",
            "country":"countryA",
            "namespace":"namespaceA"
         },
         {
            "policyVersionId":"policyVersionF",
            "country":"countryB",
            "namespace":"namespaceB"
         }
      ],
      "testClientID":[
         {
            "policyVersionId":"policyVersionB",
            "country":"countryB",
            "namespace":"namespaceB"
         },
         {
            "policyVersionId":"policyVersionA",
            "country":"countryA",
            "namespace":"namespaceA"
         },
         {
            "policyVersionId":"policyVersionD",
            "country":"countryA",
            "namespace":"namespaceA"
         }
      ],
      "testClientID_A":[
          {
            "policyVersionId":"policyVersionE",
            "country":"countryA",
            "namespace":"namespaceA"
         }
      ]
   }
}`
)

var testClient *DefaultLegalClient

func init() {
	testClient = &DefaultLegalClient{
		legalConfig:               &LegalConfig{},
		policyVersion:             nil,
		policyVersionCache:        cache.New(cache.DefaultExpiration, cache.DefaultExpiration),
		policyVersionRefreshError: nil,
		remotePolicyValidation:    nil,
		httpClient:                nil,
	}

	testClient.policyVersionCache.Set(
		testClientID,
		[]PolicyVersion {
			{
				PolicyVersionID:policyVersionA,
				Country:countryA,
				Namespace: namespaceA,
			},
			{
				PolicyVersionID:policyVersionB,
				Country:countryB,
				Namespace: namespaceB,
			},
			{
				PolicyVersionID:policyVersionD,
				Country:countryA,
				Namespace: namespaceA,
			},
		},
		cache.DefaultExpiration)

	testClient.policyVersionCache.Set(
		allAffectedClientID,
		[]PolicyVersion {
			{
				PolicyVersionID:policyVersionC,
				Country:countryA,
				Namespace: namespaceA,
			},
			{
				PolicyVersionID:policyVersionF,
				Country:countryB,
				Namespace: namespaceB,
			},
		},
		cache.DefaultExpiration)

	testClient.policyVersionCache.Set(
		testClientIDA,
		[]PolicyVersion {
			{
				PolicyVersionID:policyVersionE,
				Country:countryA,
				Namespace: namespaceA,
			},
		},
		cache.DefaultExpiration)

	testClient.remotePolicyValidation =
		func(listPolicyVersion []string, clientID, country, namespace string) (b bool, e error) {
			return true, nil
		}
}

func Test_NewDefaultLegalClient(t *testing.T) {
	conf := &LegalConfig{}
	c := NewDefaultLegalClient(conf)

	defaultLegalClient := c.(*DefaultLegalClient)

	assert.Equal(t, defaultPolicyVersionCacheTime, defaultLegalClient.legalConfig.PolicyVersionRefreshInterval)
}

func TestDefaultLegalClient_StartCachingCrucialLegal(t *testing.T) {
	mockHTTPClient := &httpClientMock{
		doMock: func(req *http.Request) (*http.Response, error) {
			body := affectedClientTest

			return &http.Response{
				Status:     http.StatusText(http.StatusOK),
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
				Header:     http.Header{},
			}, nil
		},
	}

	conf := &LegalConfig{}
	c := NewDefaultLegalClient(conf)
	defaultLegalClient := c.(*DefaultLegalClient)
	defaultLegalClient.httpClient = mockHTTPClient

	err := defaultLegalClient.StartLocalCachingCrucial()

	assert.NoError(t, err, "start caching crucial legal success")
}

func TestDefaultLegalClient_ValidatePolicyVersionsTrue(t *testing.T) {
	policyVersions := make([]string, 0)
	policyVersions = append(policyVersions, "policyVersionA")
	policyVersions = append(policyVersions, "policyVersionC")
	policyVersions = append(policyVersions, "policyVersionD")

	mockHTTPClient := &httpClientMock{
		doMock: func(req *http.Request) (*http.Response, error) {
			body := affectedClientTest

			return &http.Response{
				Status:     http.StatusText(http.StatusOK),
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
				Header:     http.Header{},
			}, nil
		},
	}

	conf := &LegalConfig{
		PublisherNamespace: namespaceB,
	}
	c := NewDefaultLegalClient(conf)
	defaultLegalClient := c.(*DefaultLegalClient)
	defaultLegalClient.httpClient = mockHTTPClient
	
	jwtClaimsTest := &iam.JWTClaims{
		Namespace: namespaceA,
		AcceptedPolicyVersion: policyVersions,
		Country: countryA,
		ClientID: testClientID,
	}

	err := defaultLegalClient.StartLocalCachingCrucial()
	valid, err := defaultLegalClient.ValidatePolicyVersions(jwtClaimsTest)

	assert.NoError(t, err, "error in validating policy versions")
	assert.True(t, valid, "policy versions is not valid")
}

func TestDefaultLegalClient_ValidatePolicyVersionsFalse(t *testing.T) {
	policyVersions := make([]string, 0)
	policyVersions = append(policyVersions, "policyVersionB")

	mockHTTPClient := &httpClientMock{
		doMock: func(req *http.Request) (*http.Response, error) {
			body := affectedClientTest

			return &http.Response{
				Status:     http.StatusText(http.StatusOK),
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
				Header:     http.Header{},
			}, nil
		},
	}

	conf := &LegalConfig{}
	c := NewDefaultLegalClient(conf)
	defaultLegalClient := c.(*DefaultLegalClient)
	defaultLegalClient.httpClient = mockHTTPClient

	jwtClaimsTest := &iam.JWTClaims{
		Namespace: namespaceA,
		AcceptedPolicyVersion: policyVersions,
		Country: countryA,
		ClientID: testClientID,
	}

	err := defaultLegalClient.StartLocalCachingCrucial()
	valid, err := defaultLegalClient.ValidatePolicyVersions(jwtClaimsTest)

	assert.NoError(t, err, "error in validating policy versions")
	assert.False(t, valid, "not all policy version signed should be false")
}

func TestDefaultLegalClient_ValidatePolicyVersionsPartialSign(t *testing.T) {
	policyVersions := make([]string, 0)
	policyVersions = append(policyVersions, "policyVersionB")

	mockHTTPClient := &httpClientMock{
		doMock: func(req *http.Request) (*http.Response, error) {
			body := affectedClientTest

			return &http.Response{
				Status:     http.StatusText(http.StatusOK),
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
				Header:     http.Header{},
			}, nil
		},
	}

	conf := &LegalConfig{}
	c := NewDefaultLegalClient(conf)
	defaultLegalClient := c.(*DefaultLegalClient)
	defaultLegalClient.httpClient = mockHTTPClient

	jwtClaimsTest := &iam.JWTClaims{
		Namespace: namespaceB,
		AcceptedPolicyVersion: policyVersions,
		Country: countryB,
		ClientID: testClientID,
	}

	err := defaultLegalClient.StartLocalCachingCrucial()
	valid, err := defaultLegalClient.ValidatePolicyVersions(jwtClaimsTest)

	assert.NoError(t, err, "error in validating policy versions")
	assert.False(t, valid, "not all policy version signed")
}

func TestDefaultLegalClient_ValidatePolicyVersionsEmptyRequiredPolicyVersions(t *testing.T) {
	policyVersions := make([]string, 0)
	policyVersions = append(policyVersions, "policyVersionB")

	mockHTTPClient := &httpClientMock{
		doMock: func(req *http.Request) (*http.Response, error) {
			body := `{ "affectedClient": { "all": [] }}`

			return &http.Response{
				Status:     http.StatusText(http.StatusOK),
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
				Header:     http.Header{},
			}, nil
		},
	}

	conf := &LegalConfig{}
	c := NewDefaultLegalClient(conf)
	defaultLegalClient := c.(*DefaultLegalClient)
	defaultLegalClient.httpClient = mockHTTPClient

	jwtClaimsTest := &iam.JWTClaims{
		Namespace: namespaceB,
		AcceptedPolicyVersion: policyVersions,
		Country: countryB,
		ClientID: testClientID,
	}

	err := defaultLegalClient.StartLocalCachingCrucial()
	valid, err := defaultLegalClient.ValidatePolicyVersions(jwtClaimsTest)

	assert.NoError(t, err, "error in validating policy versions")
	assert.True(t, valid, "empty required policy versions")
}

func TestDefaultLegalClient_ValidatePolicyVersionsSpecificClientHaveCrucialAllClientHaveCrucialFailed(t *testing.T) {
	policyVersions := make([]string, 0)
	policyVersions = append(policyVersions, "policyVersionA")

	mockHTTPClient := &httpClientMock{
		doMock: func(req *http.Request) (*http.Response, error) {
			body := affectedClientTest

			return &http.Response{
				Status:     http.StatusText(http.StatusOK),
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
				Header:     http.Header{},
			}, nil
		},
	}

	conf := &LegalConfig{}
	c := NewDefaultLegalClient(conf)
	defaultLegalClient := c.(*DefaultLegalClient)
	defaultLegalClient.httpClient = mockHTTPClient

	jwtClaimsTest := &iam.JWTClaims{
		Namespace: namespaceA,
		AcceptedPolicyVersion: policyVersions,
		Country: countryA,
		ClientID: testClientID,
	}

	err := defaultLegalClient.StartLocalCachingCrucial()
	valid, err := defaultLegalClient.ValidatePolicyVersions(jwtClaimsTest)

	assert.NoError(t, err, "error in validating policy versions")
	assert.False(t, valid, "not all policy version signed")
}

func TestDefaultLegalClient_ValidatePolicyVersionsSpecificClientHaveCrucialAllClientHaveCrucialSuccess(t *testing.T) {
	policyVersions := make([]string, 0)
	policyVersions = append(policyVersions, "policyVersionA")
	policyVersions = append(policyVersions, "policyVersionC")
	policyVersions = append(policyVersions, "policyVersionD")

	mockHTTPClient := &httpClientMock{
		doMock: func(req *http.Request) (*http.Response, error) {
			body := affectedClientTest

			return &http.Response{
				Status:     http.StatusText(http.StatusOK),
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewBufferString(body)),
				Header:     http.Header{},
			}, nil
		},
	}

	conf := &LegalConfig{}
	c := NewDefaultLegalClient(conf)
	defaultLegalClient := c.(*DefaultLegalClient)
	defaultLegalClient.httpClient = mockHTTPClient

	jwtClaimsTest := &iam.JWTClaims{
		Namespace: namespaceA,
		AcceptedPolicyVersion: policyVersions,
		Country: countryA,
		ClientID: testClientID,
	}

	err := defaultLegalClient.StartLocalCachingCrucial()
	valid, err := defaultLegalClient.ValidatePolicyVersions(jwtClaimsTest)

	assert.NoError(t, err, "error in validating policy versions")
	assert.True(t, valid, "not all policy version signed")
}

type httpClientMock struct {
	http.Client
	doMock func(req *http.Request) (*http.Response, error)
}

func (c *httpClientMock) Do(req *http.Request) (*http.Response, error) {
	return c.doMock(req)
}