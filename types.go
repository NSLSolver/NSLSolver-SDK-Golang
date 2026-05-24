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
	Success bool    `json:"success"`
	Token   string  `json:"token"`
	// Cost is the USD amount deducted from the account balance for this solve.
	Cost float64 `json:"cost"`
	Type string  `json:"type"`
}

// ChallengeResult is the response from a successful Challenge solve.
type ChallengeResult struct {
	Success   bool             `json:"success"`
	Cookies   ChallengeCookies `json:"cookies"`
	UserAgent string           `json:"user_agent"`
	// Token is set when the challenge page returns a Turnstile-style token instead of (or alongside) cookies.
	Token string  `json:"token,omitempty"`
	Cost  float64 `json:"cost"`
	Type  string  `json:"type"`
}

// ChallengeCookies holds the cookies returned from a challenge solve.
type ChallengeCookies struct {
	CFClearance string `json:"cf_clearance"`
}

// KasadaParams are the parameters for solving a Kasada challenge.
type KasadaParams struct {
	URL          string       `json:"url"`
	UserAgent    string       `json:"user_agent"`
	UAVersion    int          `json:"ua_version"`
	KasadaConfig KasadaConfig `json:"kasada_config"`
	Proxy        string       `json:"proxy,omitempty"`
}

// KasadaConfig holds the Kasada-specific configuration.
type KasadaConfig struct {
	PJSPath    string `json:"p_js_path"`
	FPHost     string `json:"fp_host"`
	TLHost     string `json:"tl_host"`
	CDConstant string `json:"cd_constant,omitempty"`
}

// KasadaResult is the response from a successful Kasada solve.
type KasadaResult struct {
	Success bool          `json:"success"`
	Headers KasadaHeaders `json:"headers"`
	Cost    float64       `json:"cost"`
	Type    string        `json:"type"`
}

// KasadaHeaders holds the headers returned from a Kasada solve.
type KasadaHeaders struct {
	XKpsdkCT string `json:"x-kpsdk-ct"`
	XKpsdkCD string `json:"x-kpsdk-cd"`
	XKpsdkV  string `json:"x-kpsdk-v"`
	XKpsdkH  string `json:"x-kpsdk-h"`
}

// AkamaiParams are the parameters for solving an Akamai Bot Manager challenge.
// UserAgent and Proxy are both required — the _abck cookie is bound to the
// proxy's egress IP and to the submitted UA.
type AkamaiParams struct {
	URL       string `json:"url"`
	UserAgent string `json:"user_agent"`
	Proxy     string `json:"proxy"`
}

// AkamaiResult is the response from a successful Akamai solve.
// Cookies includes _abck and the rest of the jar (bm_sz, ak_bmsc, …).
type AkamaiResult struct {
	Success bool              `json:"success"`
	Cookies map[string]string `json:"cookies"`
	Cost    float64           `json:"cost"`
	Type    string            `json:"type"`
}

// BalanceResult holds account balance, plan flags, and live CPM (captchas-per-minute) usage.
type BalanceResult struct {
	Success            bool     `json:"success"`
	Balance            float64  `json:"balance"`
	Unlimited          bool     `json:"unlimited"`
	AllowedTypes       []string `json:"allowed_types"`
	// MaxCPM is the per-key captchas-per-minute ceiling. 0 means uncapped.
	MaxCPM int `json:"max_cpm"`
	// CurrentCPM is how many tokens have been consumed in the rolling minute.
	CurrentCPM int `json:"current_cpm"`
	// CPMLimit mirrors MaxCPM for symmetry with monitoring dashboards.
	CPMLimit int `json:"cpm_limit"`
	// UnlimitedExpiresAt is set when Unlimited is true and an expiry was configured.
	UnlimitedExpiresAt string `json:"unlimited_expires_at,omitempty"`
}

type solveRequest struct {
	Type         string        `json:"type"`
	SiteKey      string        `json:"site_key,omitempty"`
	URL          string        `json:"url"`
	Action       string        `json:"action,omitempty"`
	CData        string        `json:"cdata,omitempty"`
	Proxy        string        `json:"proxy,omitempty"`
	UserAgent    string        `json:"user_agent,omitempty"`
	UAVersion    int           `json:"ua_version,omitempty"`
	KasadaConfig *KasadaConfig `json:"kasada_config,omitempty"`
}
