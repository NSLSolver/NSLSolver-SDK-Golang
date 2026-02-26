package nslsolver

// TurnstileParams are the parameters for solving a Cloudflare Turnstile captcha.
type TurnstileParams struct {
	SiteKey   string `json:"site_key"`
	URL       string `json:"url"`
	Action    string `json:"action,omitempty"`
	CData     string `json:"cdata,omitempty"`
	Proxy     string `json:"proxy,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}

// ChallengeParams are the parameters for solving a Cloudflare Challenge page.
type ChallengeParams struct {
	URL       string `json:"url"`
	Proxy     string `json:"proxy"`
	UserAgent string `json:"user_agent,omitempty"`
}

// TurnstileResult is the response from a successful Turnstile solve.
type TurnstileResult struct {
	Success bool   `json:"success"`
	Token   string `json:"token"`
	Type    string `json:"type"`
}

// ChallengeResult is the response from a successful Challenge solve.
type ChallengeResult struct {
	Success   bool             `json:"success"`
	Cookies   ChallengeCookies `json:"cookies"`
	UserAgent string           `json:"user_agent"`
	Type      string           `json:"type"`
}

// ChallengeCookies holds the cookies returned from a challenge solve.
type ChallengeCookies struct {
	CFClearance string `json:"cf_clearance"`
}

// BalanceResult holds account balance and configuration info.
type BalanceResult struct {
	Balance      float64  `json:"balance"`
	MaxThreads   int      `json:"max_threads"`
	AllowedTypes []string `json:"allowed_types"`
}

type solveRequest struct {
	Type      string `json:"type"`
	SiteKey   string `json:"site_key,omitempty"`
	URL       string `json:"url"`
	Action    string `json:"action,omitempty"`
	CData     string `json:"cdata,omitempty"`
	Proxy     string `json:"proxy,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}
