package webshare

import (
	"context"
	"net/http"
	"time"
)

// AuthService exposes session and account lifecycle operations. Several of
// these endpoints are recaptcha-gated and documented as usable only from the
// Webshare dashboard; they are included for completeness and marked in their
// doc comments.
type AuthService struct {
	client *Client
}

// AuthResult is returned by registration endpoints.
type AuthResult struct {
	// Token is a login token usable to authorize API requests.
	Token string `json:"token"`
	// LoggedInExistingUser reports whether an existing user was logged in
	// instead of a new account being created.
	LoggedInExistingUser bool `json:"logged_in_existing_user"`
}

// TokenResult is returned by endpoints that issue a login token.
type TokenResult struct {
	// Token is a login token usable to authorize API requests.
	Token string `json:"token"`
}

// RegisterParams are the parameters for AuthService.Register.
type RegisterParams struct {
	// Email is the address to register with. May require email validation
	// later on.
	Email string `json:"email"`
	// Password must be at least 8 characters, not too similar to the email,
	// not all numeric and not a common password.
	Password string `json:"password"`
	// Recaptcha is the recaptcha token.
	Recaptcha string `json:"recaptcha"`
	// TOSAccepted must be true to process the request.
	TOSAccepted bool `json:"tos_accepted"`
	// MarketingEmailAccepted opts in to marketing emails.
	MarketingEmailAccepted bool `json:"marketing_email_accepted"`
}

// Register registers a new account with email and password. Unauthenticated.
// This endpoint requires recaptcha validation: the docs mark it as usable
// only from the Webshare dashboard, not programmatically.
func (s *AuthService) Register(ctx context.Context, params RegisterParams, opts ...RequestOption) (*AuthResult, error) {
	out := &AuthResult{}
	if err := s.client.doJSON(ctx, http.MethodPost, "/api/v2/register/", nil, params, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// SocialProvider identifies a social login provider.
type SocialProvider string

// ProviderGoogle is currently the only supported social provider.
const ProviderGoogle SocialProvider = "google"

// RegisterSocialParams are the parameters for AuthService.RegisterSocial.
type RegisterSocialParams struct {
	// Provider is the social provider to register with.
	Provider SocialProvider `json:"provider"`
	// Code is the auth code received from the social provider.
	Code string `json:"code"`
	// RedirectURI must match the authorized redirect URIs configured with
	// the provider.
	RedirectURI string `json:"redirect_uri"`
	// TOSAccepted must be true to process the request.
	TOSAccepted bool `json:"tos_accepted"`
	// MarketingEmailAccepted opts in to marketing emails.
	MarketingEmailAccepted bool `json:"marketing_email_accepted"`
}

// RegisterSocial registers a new account with a social provider (Google
// OAuth2). Unauthenticated.
func (s *AuthService) RegisterSocial(ctx context.Context, params RegisterSocialParams, opts ...RequestOption) (*AuthResult, error) {
	out := &AuthResult{}
	if err := s.client.doJSON(ctx, http.MethodPost, "/api/v2/register/social/", nil, params, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// LoginParams are the parameters for AuthService.Login.
type LoginParams struct {
	// Email is the address previously registered with.
	Email string `json:"email"`
	// Password is the password previously registered with.
	Password string `json:"password"`
	// Recaptcha is the recaptcha token.
	Recaptcha string `json:"recaptcha"`
}

// Login logs in to an existing account with email and password.
// Unauthenticated. This endpoint requires recaptcha validation: the docs
// mark it as usable only from the Webshare dashboard, not programmatically.
func (s *AuthService) Login(ctx context.Context, params LoginParams, opts ...RequestOption) (*TokenResult, error) {
	out := &TokenResult{}
	if err := s.client.doJSON(ctx, http.MethodPost, "/api/v2/login/", nil, params, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// LoginSocialParams are the parameters for AuthService.LoginSocial.
type LoginSocialParams struct {
	// Provider is the social provider to log in with.
	Provider SocialProvider `json:"provider"`
	// Code is the auth code received from the social provider.
	Code string `json:"code"`
	// RedirectURI must match the authorized redirect URIs configured with
	// the provider.
	RedirectURI string `json:"redirect_uri"`
}

// LoginSocial logs in to an existing account with a social provider (Google
// OAuth2). Unauthenticated.
func (s *AuthService) LoginSocial(ctx context.Context, params LoginSocialParams, opts ...RequestOption) (*TokenResult, error) {
	out := &TokenResult{}
	if err := s.client.doJSON(ctx, http.MethodPost, "/api/v2/login/social/", nil, params, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// Logout invalidates the token used to make this request.
func (s *AuthService) Logout(ctx context.Context, opts ...RequestOption) error {
	return s.client.doJSON(ctx, http.MethodPost, "/api/v2/logout/", nil, nil, nil, opts)
}

// ChangePasswordParams are the parameters for AuthService.ChangePassword.
type ChangePasswordParams struct {
	// Password is the current password.
	Password string `json:"password"`
	// NewPassword is the new password; it must meet the registration
	// password requirements.
	NewPassword string `json:"new_password"`
}

// ChangePassword changes the current password. On success all API tokens
// except the current one are disabled. Accounts registered via Google OAuth
// may not have a password and must use the password reset flow instead.
func (s *AuthService) ChangePassword(ctx context.Context, params ChangePasswordParams, opts ...RequestOption) error {
	return s.client.doJSON(ctx, http.MethodPost, "/api/v2/changepassword/", nil, params, nil, opts)
}

// PasswordResetRequestParams are the parameters for
// AuthService.RequestPasswordReset.
type PasswordResetRequestParams struct {
	// Email is the address of the existing user.
	Email string `json:"email"`
	// Recaptcha is the recaptcha token.
	Recaptcha string `json:"recaptcha"`
}

// RequestPasswordReset requests a password reset email. Unauthenticated.
// Beyond basic validation this endpoint always succeeds, even for unknown
// email addresses.
func (s *AuthService) RequestPasswordReset(ctx context.Context, params PasswordResetRequestParams, opts ...RequestOption) error {
	return s.client.doJSON(ctx, http.MethodPost, "/api/v2/resetpassword/", nil, params, nil, opts)
}

// PasswordResetCompleteParams are the parameters for
// AuthService.CompletePasswordReset.
type PasswordResetCompleteParams struct {
	// Password is the new password.
	Password string `json:"password"`
	// PasswordResetToken is the token retrieved from the reset email.
	PasswordResetToken string `json:"password_reset_token"`
	// Recaptcha is the recaptcha token.
	Recaptcha string `json:"recaptcha"`
}

// CompletePasswordReset completes a password reset. Unauthenticated. On
// success all previous API tokens are invalidated and a new token is
// returned.
func (s *AuthService) CompletePasswordReset(ctx context.Context, params PasswordResetCompleteParams, opts ...RequestOption) (*TokenResult, error) {
	out := &TokenResult{}
	if err := s.client.doJSON(ctx, http.MethodPost, "/api/v2/resetpassword/complete/", nil, params, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// EmailChangeRequestParams are the parameters for
// AuthService.RequestEmailChange.
type EmailChangeRequestParams struct {
	// Password is the user's password.
	Password string `json:"password"`
	// NewEmail is the address to switch to. Validation checks whether the
	// email already exists in the system.
	NewEmail string `json:"new_email"`
}

// RequestEmailChange requests an email change; on success a confirmation
// email is sent to the new address.
func (s *AuthService) RequestEmailChange(ctx context.Context, params EmailChangeRequestParams, opts ...RequestOption) error {
	return s.client.doJSON(ctx, http.MethodPost, "/api/v2/changeemail/", nil, params, nil, opts)
}

// EmailChangeCompleteParams are the parameters for
// AuthService.CompleteEmailChange.
type EmailChangeCompleteParams struct {
	// ConfirmationCode is the code retrieved from the confirmation email.
	ConfirmationCode string `json:"confirmation_code"`
}

// CompleteEmailChange completes an email change. The user must be
// authenticated to complete the request.
func (s *AuthService) CompleteEmailChange(ctx context.Context, params EmailChangeCompleteParams, opts ...RequestOption) error {
	return s.client.doJSON(ctx, http.MethodPost, "/api/v2/changeemail/complete/", nil, params, nil, opts)
}

// ActivationStatus describes the state of the account activation.
type ActivationStatus struct {
	// EmailIsVerified reports whether the email is verified.
	EmailIsVerified bool `json:"email_is_verified"`
	// LastTimeEmailVerificationEmailSent is when the activation email was
	// last sent. May be nil.
	LastTimeEmailVerificationEmailSent *time.Time `json:"last_time_email_verification_email_sent"`
	// CreatedAt is when the activation status object was created.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is when the activation status was last updated.
	UpdatedAt time.Time `json:"updated_at"`
}

// GetActivation retrieves the current state of the account activation.
func (s *AuthService) GetActivation(ctx context.Context, opts ...RequestOption) (*ActivationStatus, error) {
	out := &ActivationStatus{}
	if err := s.client.doJSON(ctx, http.MethodGet, "/api/v2/activation/", nil, nil, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// ResendActivation re-sends the activation email. There is a limit on how
// often activation emails can be re-sent.
func (s *AuthService) ResendActivation(ctx context.Context, opts ...RequestOption) (*ActivationStatus, error) {
	out := &ActivationStatus{}
	if err := s.client.doJSON(ctx, http.MethodPost, "/api/v2/activation/resend/", nil, nil, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// ActivationCompleteParams are the parameters for
// AuthService.CompleteActivation.
type ActivationCompleteParams struct {
	// ActivationToken is the token retrieved from the activation email.
	ActivationToken string `json:"activation_token"`
}

// CompleteActivation completes the account activation and returns a new API
// token; existing tokens keep working. Authentication headers are optional
// but encouraged for this endpoint.
func (s *AuthService) CompleteActivation(ctx context.Context, params ActivationCompleteParams, opts ...RequestOption) (*TokenResult, error) {
	out := &TokenResult{}
	if err := s.client.doJSON(ctx, http.MethodPost, "/api/v2/activation/complete/", nil, params, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteAccountParams are the parameters for AuthService.DeleteAccount.
type DeleteAccountParams struct {
	// Password is the user's password.
	Password string `json:"password"`
	// Recaptcha is the recaptcha token.
	Recaptcha string `json:"recaptcha"`
}

// DeleteAccount deletes the Webshare account. This endpoint requires
// recaptcha validation: the docs mark it as usable only from the Webshare
// dashboard, not programmatically. After deletion every API request returns
// 403 with code account_deleted.
func (s *AuthService) DeleteAccount(ctx context.Context, params DeleteAccountParams, opts ...RequestOption) error {
	return s.client.doJSON(ctx, http.MethodPost, "/api/v2/deleteaccount/", nil, params, nil, opts)
}

// DeleteAccountSocialParams are the parameters for
// AuthService.DeleteAccountSocial.
type DeleteAccountSocialParams struct {
	// Provider is the social provider the account was registered with.
	Provider SocialProvider `json:"provider"`
	// Code is the auth code received from the social provider.
	Code string `json:"code"`
	// RedirectURI must match the authorized redirect URIs configured with
	// the provider.
	RedirectURI string `json:"redirect_uri"`
}

// DeleteAccountSocial deletes a Webshare account registered via a social
// provider. After deletion every API request returns 403 with code
// account_deleted.
func (s *AuthService) DeleteAccountSocial(ctx context.Context, params DeleteAccountSocialParams, opts ...RequestOption) error {
	return s.client.doJSON(ctx, http.MethodPost, "/api/v2/deleteaccount/social/", nil, params, nil, opts)
}
