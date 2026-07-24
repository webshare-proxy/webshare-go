package webshare

import (
	"context"
	"net/http"
	"time"
)

// ProfileService exposes the user profile operations.
type ProfileService struct {
	client *Client
}

// Profile is the user profile of the account.
type Profile struct {
	// ID is the unique identifier of the profile instance.
	ID int `json:"id"`
	// Email is the user's email address. Read-only.
	Email string `json:"email"`
	// FirstName is the user's first name. Can be empty.
	FirstName string `json:"first_name"`
	// LastName is the user's last name. Can be empty.
	LastName string `json:"last_name"`
	// LastLogin is the date the user last logged in. Read-only.
	LastLogin time.Time `json:"last_login"`
	// Timezone is the user's preferred timezone.
	Timezone string `json:"timezone"`
	// SubscribedBandwidthUsageNotifications toggles bandwidth usage emails.
	SubscribedBandwidthUsageNotifications bool `json:"subscribed_bandwidth_usage_notifications"`
	// SubscribedSubscriptionNotifications toggles subscription update
	// emails.
	SubscribedSubscriptionNotifications bool `json:"subscribed_subscription_notifications"`
	// SubscribedProxyUsageStatistics toggles proxy usage statistics emails.
	SubscribedProxyUsageStatistics bool `json:"subscribed_proxy_usage_statistics"`
	// SubscribedUsageWarnings toggles proxy usage warning emails.
	SubscribedUsageWarnings bool `json:"subscribed_usage_warnings"`
	// SubscribedGuidesAndTips toggles guides-and-tips emails.
	SubscribedGuidesAndTips bool `json:"subscribed_guides_and_tips"`
	// SubscribedSurveyEmails toggles survey emails.
	SubscribedSurveyEmails bool `json:"subscribed_survey_emails"`
	// TrackingID is a unique user ID for identifying the user with external
	// services. Read-only.
	TrackingID string `json:"tracking_id"`
	// AnnounceKitUserToken is a token for the AnnounceKit widget. Observed
	// on the live API; absent from the documented object.
	AnnounceKitUserToken string `json:"announce_kit_user_token,omitempty"`
	// HelpscoutBeaconSignature is a signature for the Help Scout beacon.
	// Observed on the live API; absent from the documented object.
	HelpscoutBeaconSignature string `json:"helpscout_beacon_signature,omitempty"`
	// IntercomSignature is a signature for the Intercom widget. Observed on
	// the live API; absent from the documented object.
	IntercomSignature string `json:"intercom_signature,omitempty"`
	// IsVIPCustomer reports whether the account is a VIP customer. Observed
	// on the live API; absent from the documented object.
	IsVIPCustomer bool `json:"is_vip_customer,omitempty"`
	// CreatedAt is the registration date. Read-only.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is when this instance was last updated.
	UpdatedAt time.Time `json:"updated_at"`
}

// Get retrieves the user profile.
func (s *ProfileService) Get(ctx context.Context, opts ...RequestOption) (*Profile, error) {
	out := &Profile{}
	if err := s.client.doJSON(ctx, http.MethodGet, "/api/v2/profile/", nil, nil, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// ProfileUpdateParams are the parameters for ProfileService.Update. Only set
// fields are sent.
type ProfileUpdateParams struct {
	// FirstName sets the user's first name.
	FirstName *string `json:"first_name,omitempty"`
	// LastName sets the user's last name.
	LastName *string `json:"last_name,omitempty"`
	// Timezone sets the user's preferred timezone.
	Timezone *string `json:"timezone,omitempty"`
	// SubscribedBandwidthUsageNotifications toggles bandwidth usage emails.
	SubscribedBandwidthUsageNotifications *bool `json:"subscribed_bandwidth_usage_notifications,omitempty"`
	// SubscribedSubscriptionNotifications toggles subscription update
	// emails.
	SubscribedSubscriptionNotifications *bool `json:"subscribed_subscription_notifications,omitempty"`
	// SubscribedProxyUsageStatistics toggles proxy usage statistics emails.
	SubscribedProxyUsageStatistics *bool `json:"subscribed_proxy_usage_statistics,omitempty"`
	// SubscribedUsageWarnings toggles proxy usage warning emails.
	SubscribedUsageWarnings *bool `json:"subscribed_usage_warnings,omitempty"`
	// SubscribedGuidesAndTips toggles guides-and-tips emails.
	SubscribedGuidesAndTips *bool `json:"subscribed_guides_and_tips,omitempty"`
	// SubscribedSurveyEmails toggles survey emails.
	SubscribedSurveyEmails *bool `json:"subscribed_survey_emails,omitempty"`
}

// Update partially updates the user profile.
func (s *ProfileService) Update(ctx context.Context, params ProfileUpdateParams, opts ...RequestOption) (*Profile, error) {
	out := &Profile{}
	if err := s.client.doJSON(ctx, http.MethodPatch, "/api/v2/profile/", nil, params, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// ProfilePreferences are the user preferences of the account.
type ProfilePreferences struct {
	// ID is the unique identifier of the preferences instance.
	ID int `json:"id"`
	// CustomerSatisfactionSurveyLastDismissedAt is when the customer
	// satisfaction survey was dismissed. May be nil.
	CustomerSatisfactionSurveyLastDismissedAt *time.Time `json:"customer_satisfaction_survey_last_dismissed_at"`
	// CustomerSatisfactionSurveyLastCompletedAt is when the customer
	// satisfaction survey was completed. May be nil.
	CustomerSatisfactionSurveyLastCompletedAt *time.Time `json:"customer_satisfaction_survey_last_completed_at"`
	// OnboardingActivityPageViewedAt is when the onboarding activity page
	// was viewed. May be nil.
	OnboardingActivityPageViewedAt *time.Time `json:"onboarding_activity_page_viewed_at"`
	// CreatedAt is when the preferences instance was created. Observed on
	// the live API; absent from the documented object.
	CreatedAt time.Time `json:"created_at,omitempty"`
	// UpdatedAt is when the preferences instance was last updated. Observed
	// on the live API; absent from the documented object.
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// GetPreferences retrieves the user preferences.
func (s *ProfileService) GetPreferences(ctx context.Context, opts ...RequestOption) (*ProfilePreferences, error) {
	out := &ProfilePreferences{}
	if err := s.client.doJSON(ctx, http.MethodGet, "/api/v2/profile/preferences/", nil, nil, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// ProfilePreferencesUpdateParams are the parameters for
// ProfileService.UpdatePreferences. Any subset of fields may be set.
type ProfilePreferencesUpdateParams struct {
	// CustomerSatisfactionSurveyLastDismissedAt records when the customer
	// satisfaction survey was dismissed.
	CustomerSatisfactionSurveyLastDismissedAt *time.Time `json:"customer_satisfaction_survey_last_dismissed_at,omitempty"`
	// CustomerSatisfactionSurveyLastCompletedAt records when the customer
	// satisfaction survey was completed.
	CustomerSatisfactionSurveyLastCompletedAt *time.Time `json:"customer_satisfaction_survey_last_completed_at,omitempty"`
	// OnboardingActivityPageViewedAt records when the onboarding activity
	// page was viewed.
	OnboardingActivityPageViewedAt *time.Time `json:"onboarding_activity_page_viewed_at,omitempty"`
}

// UpdatePreferences partially updates the user preferences.
func (s *ProfileService) UpdatePreferences(ctx context.Context, params ProfilePreferencesUpdateParams, opts ...RequestOption) (*ProfilePreferences, error) {
	out := &ProfilePreferences{}
	if err := s.client.doJSON(ctx, http.MethodPatch, "/api/v2/profile/preferences/", nil, params, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}
