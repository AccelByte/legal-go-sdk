# Legal Go SDK

This is AccelByte's Legal Go SDK for integrating with Legal in Go projects.

## Usage

### Importing package

```go
import "github.com/AccelByte/legal-go-sdk"
```

### Creating default Legal client

```go
cfg := &legal.Config{
    LegalBaseURL: "<Legal URL>",
}

client := legal.NewDefaultLegalClient(cfg)
```

It's recommended that you store the **interface** rather than the type since it enables you to mock the client during tests.

```go
var client legal.LegalClient

client := legal.NewDefaultLegalClient(cfg)
```

So during tests, you can replace the `client` with:

```go
var client legal.Client

client := iam.NewMockLegalClient() // or create your own mock implementation that suits your test case
```

**Note**

By default, the client can only do crucial policy version validation by requesting to Legal service.

To enable local validation, you need to call:

```go
client.StartCachingCrucialLegal()
```

Then the client will automatically get all latest crucial policy version and refreshing them periodically.
This enables you to do local policy version validation.

### Validating Policy Version

#### Validating locally using cached policy versions:

```go
claims, _ := client.ValidatePolicyVersions(policyVersions, clientID, country, namespace)
```

**Note**

If no policy versions cached for the affected clientID, it will try to call Legal to do remote validation

### Health check

Whenever the Legal service went unhealthy, the client will know by detecting if any of the automated refresh goroutines has error.

You can check the health by:

```go
client.HealthCheck()
```
