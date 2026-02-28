# NSLSolver Go SDK

Go client for the [NSLSolver](https://nslsolver.com) captcha solving API. Supports Cloudflare Turnstile and Challenge page solves with automatic retries and exponential backoff. No external dependencies.

## Installation

```bash
go get github.com/nslsolver/nslsolver-go
```

## Usage

```go
client := nslsolver.NewClient("your-api-key")
ctx := context.Background()
```

### Turnstile

```go
result, err := client.SolveTurnstile(ctx, nslsolver.TurnstileParams{
    SiteKey: "0x4AAAAAAAB...",
    URL:     "https://example.com",
})
if err != nil {
    log.Fatal(err)
}
fmt.Println(result.Token)
```

Optional fields: `Action`, `CData`, `Proxy`, `UserAgent`.

### Challenge

Requires a proxy. Returns clearance cookies and the user agent used during the solve.

```go
result, err := client.SolveChallenge(ctx, nslsolver.ChallengeParams{
    URL:   "https://example.com/protected",
    Proxy: "http://user:pass@host:port",
})
if err != nil {
    log.Fatal(err)
}
fmt.Println(result.Cookies.CFClearance)
```

### Balance

```go
balance, err := client.GetBalance(ctx)
fmt.Printf("$%.2f, %d threads\n", balance.Balance, balance.MaxThreads)
```

## Client Options

```go
client := nslsolver.NewClient("key",
    nslsolver.WithTimeout(60 * time.Second),
    nslsolver.WithMaxRetries(5),
    nslsolver.WithHTTPClient(customHTTPClient),
    nslsolver.WithBaseURL("https://custom.example.com"),
)
```

Defaults: 120s timeout, 3 retries.

## Error Handling

All API errors are returned as `*nslsolver.APIError` with `StatusCode`, `Message`, and `Retryable` fields. Helper functions let you match specific cases:

```go
result, err := client.SolveTurnstile(ctx, params)
if err != nil {
    switch {
    case nslsolver.IsAuthError(err):       // 401
    case nslsolver.IsBalanceError(err):     // 402
    case nslsolver.IsNotAllowedError(err):  // 403
    case nslsolver.IsRateLimitError(err):   // 429
    case nslsolver.IsBackendError(err):     // 503
    case nslsolver.IsBadRequestError(err):  // 400
    }
}
```

Retryable errors (429, 503) are automatically retried with exponential backoff before being returned. Fatal errors (400, 401, 402, 403) fail immediately.

## Documentation

For more information, check out the full documentation at https://docs.nslsolver.com

## License

MIT
