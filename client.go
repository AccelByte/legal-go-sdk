// Copyright (c) 2021 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package legal

import "github.com/AccelByte/iam-go-sdk"

type LegalClient interface {
	StartCachingCrucialLegal() error

	ValidatePolicyVersions(claims *iam.JWTClaims) (bool, error)

	HealthCheck() bool
}