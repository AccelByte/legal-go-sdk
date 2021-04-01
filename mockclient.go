// Copyright (c) 2021 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package legal

import "github.com/AccelByte/iam-go-sdk"

type MockLegalClient struct {
	Healthy bool
}

func (client MockLegalClient) HealthCheck() bool {
	return client.Healthy
}

func (client MockLegalClient) StartCachingCrucialLegal() error {
	return nil
}

func (client MockLegalClient) ValidatePolicyVersions(claims *iam.JWTClaims) (bool, error) {
	return true, nil
}

func NewMockLegalClient() LegalClient {
	return &MockLegalClient{
		Healthy: true,
	}
}

