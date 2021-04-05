// Copyright (c) 2021 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package legal

type CrucialPolicyVersionResponse struct {
	AffectedClient map[string][]PolicyVersion
}

type PolicyVersion struct {
	PolicyVersionID string
	Country         string
	Namespace       string
}