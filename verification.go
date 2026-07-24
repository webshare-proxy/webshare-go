package webshare

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"iter"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// VerificationService exposes the account verification (compliance)
// operations, grouped into nested subresources.
type VerificationService struct {
	// Flows exposes the verification flow operations.
	Flows *VerificationFlowsService
	// Questions exposes the verification question operations.
	Questions *VerificationQuestionsService
	// Appeals exposes the suspension appeal operations.
	Appeals *VerificationAppealsService
	// AbuseReports exposes the abuse report operations.
	AbuseReports *VerificationAbuseReportsService

	client *Client
}

// VerificationFlowType is the kind of verification flow.
type VerificationFlowType string

// Verification flow types.
const (
	// FlowAcceptableUseViolation is a verification for acceptable use
	// policy violations.
	FlowAcceptableUseViolation VerificationFlowType = "acceptable_use_violation"
	// FlowAbuseReport is a verification for abuse reports.
	FlowAbuseReport VerificationFlowType = "abuse_report"
	// FlowFraudulentPayment is a verification for fraudulent payments.
	FlowFraudulentPayment VerificationFlowType = "fraudulent_payment"
)

// VerificationFlowState is the state of a verification flow.
type VerificationFlowState string

// Verification flow states.
const (
	// FlowInflow means the verification is in progress.
	FlowInflow VerificationFlowState = "inflow"
	// FlowSuccessful means the verification succeeded.
	FlowSuccessful VerificationFlowState = "successful_verification"
	// FlowFailed means the verification failed.
	FlowFailed VerificationFlowState = "failed_verification"
)

// VerificationFile is a file attached to evidence or an answer.
type VerificationFile struct {
	// ID is the unique identifier of the file.
	ID int `json:"id"`
	// File is the stored file reference.
	File string `json:"file"`
	// CreatedAt is when the file was uploaded.
	CreatedAt time.Time `json:"created_at"`
}

// VerificationEvidence is the evidence submitted for a verification flow.
type VerificationEvidence struct {
	// ID is the unique identifier of the evidence.
	ID int `json:"id"`
	// Explanation is the explanation submitted with the evidence.
	Explanation string `json:"explanation"`
	// CreatedAt is when the evidence was created.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is when the evidence was last updated.
	UpdatedAt time.Time `json:"updated_at"`
	// Files lists the submitted files.
	Files []VerificationFile `json:"files"`
}

// VerificationFlow is one account verification.
type VerificationFlow struct {
	// ID is the unique identifier of the verification instance.
	ID int `json:"id"`
	// Type is the verification type.
	Type VerificationFlowType `json:"type"`
	// State is the current state of the verification.
	State VerificationFlowState `json:"state"`
	// StartedAt is when this verification started.
	StartedAt time.Time `json:"started_at"`
	// UpdatedAt is when this instance was last updated.
	UpdatedAt time.Time `json:"updated_at"`
	// NeedsEvidence reports whether the verification requires evidence.
	NeedsEvidence bool `json:"needs_evidence"`
	// Evidence holds the submitted evidence. May be nil.
	Evidence *VerificationEvidence `json:"evidence"`
	// IDVerificationRestoresAccess reports whether completing ID
	// verification restores proxy access.
	IDVerificationRestoresAccess bool `json:"id_verification_restores_access"`
	// IDVerificationRequired reports whether this verification requires ID
	// verification.
	IDVerificationRequired bool `json:"id_verification_required"`
}

// VerificationFlowsService exposes the verification flow operations.
type VerificationFlowsService struct {
	client *Client
}

// VerificationFlowListParams are the parameters for
// VerificationFlowsService.List.
type VerificationFlowListParams struct {
	// Page is the page number.
	Page *int
	// PageSize is the number of results per page.
	PageSize *int
}

// List returns the account verifications in paginated format.
func (s *VerificationFlowsService) List(ctx context.Context, params VerificationFlowListParams, opts ...RequestOption) (*Page[VerificationFlow], error) {
	q := url.Values{}
	setInt(q, "page", params.Page)
	setInt(q, "page_size", params.PageSize)
	return getPage[VerificationFlow](ctx, s.client, "/api/v2/verification/flow/", q, opts)
}

// ListAll returns a lazy iterator over every verification flow across all
// pages.
func (s *VerificationFlowsService) ListAll(ctx context.Context, params VerificationFlowListParams, opts ...RequestOption) iter.Seq2[VerificationFlow, error] {
	return iterPages(ctx, func(ctx context.Context) (*Page[VerificationFlow], error) {
		return s.List(ctx, params, opts...)
	})
}

// Get retrieves an account verification.
func (s *VerificationFlowsService) Get(ctx context.Context, id int, opts ...RequestOption) (*VerificationFlow, error) {
	out := &VerificationFlow{}
	if err := s.client.doJSON(ctx, http.MethodGet, "/api/v2/verification/flow/"+strconv.Itoa(id)+"/", nil, nil, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// File is a file to upload in a multipart request.
type File struct {
	// Name is the file name reported to the API.
	Name string
	// Reader supplies the file contents.
	Reader io.Reader
}

// SubmitEvidenceParams are the parameters for
// VerificationFlowsService.SubmitEvidence.
type SubmitEvidenceParams struct {
	// Explanation is the explanation submitted as part of the verification.
	Explanation string
	// Files lists the files submitted as part of the verification.
	Files []File
}

// SubmitEvidence submits evidence for a verification. The request is encoded
// as multipart/form-data.
func (s *VerificationFlowsService) SubmitEvidence(ctx context.Context, id int, params SubmitEvidenceParams, opts ...RequestOption) (*VerificationFlow, error) {
	fields := map[string]string{}
	if params.Explanation != "" {
		fields["explanation"] = params.Explanation
	}
	body, contentType, err := encodeMultipart(fields, "files", params.Files)
	if err != nil {
		return nil, err
	}
	out := &VerificationFlow{}
	path := "/api/v2/verification/flow/" + strconv.Itoa(id) + "/submit_evidence/"
	if err := s.client.doMultipart(ctx, http.MethodPost, path, body, contentType, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// SubmitSecurityCodeParams are the parameters for
// VerificationFlowsService.SubmitSecurityCode.
type SubmitSecurityCodeParams struct {
	// SecurityCode is the two-character code found in Webshare charges on
	// the user's bank statement.
	SecurityCode string `json:"security_code"`
}

// SubmitSecurityCode submits a security code for a fraudulent payment
// verification flow.
func (s *VerificationFlowsService) SubmitSecurityCode(ctx context.Context, id int, params SubmitSecurityCodeParams, opts ...RequestOption) (*VerificationFlow, error) {
	out := &VerificationFlow{}
	path := "/api/v2/verification/flow/" + strconv.Itoa(id) + "/submit_verification_code/"
	if err := s.client.doJSON(ctx, http.MethodPost, path, nil, params, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// VerificationAnswer is the answer submitted for a verification question.
type VerificationAnswer struct {
	// ID is the unique identifier of the answer.
	ID int `json:"id"`
	// Answer is the answer text.
	Answer string `json:"answer"`
	// CreatedAt is when the answer was created.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is when the answer was last updated.
	UpdatedAt time.Time `json:"updated_at"`
	// Files lists the attachments submitted with the answer.
	Files []VerificationFile `json:"files"`
}

// VerificationQuestion is a question submitted by the compliance team.
type VerificationQuestion struct {
	// ID is the unique identifier of the question.
	ID int `json:"id"`
	// Question is the question text.
	Question string `json:"question"`
	// CreatedAt is when the question was created.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is when the question was last updated.
	UpdatedAt time.Time `json:"updated_at"`
	// Flow is the ID of the related verification flow.
	Flow int `json:"flow"`
	// Answer holds the submitted answer, or nil when unanswered.
	Answer *VerificationAnswer `json:"answer"`
}

// VerificationQuestionsService exposes the verification question operations.
type VerificationQuestionsService struct {
	client *Client
}

// VerificationQuestionListParams are the parameters for
// VerificationQuestionsService.List.
type VerificationQuestionListParams struct {
	// FlowType matches questions whose verification flow has the given
	// type.
	FlowType VerificationFlowType
	// FlowState matches questions whose verification flow is in the given
	// state.
	FlowState VerificationFlowState
	// AnswerIsNull set to true shows only unanswered questions; false shows
	// only answered ones.
	AnswerIsNull *bool
	// FlowStartedAtGTE bounds the flow start date from below.
	FlowStartedAtGTE *time.Time
	// FlowStartedAtLTE bounds the flow start date from above.
	FlowStartedAtLTE *time.Time
	// Question matches questions with the given question text.
	Question string
	// AnswerAnswer matches questions with the given answer text.
	AnswerAnswer string
	// Page is the page number.
	Page *int
	// PageSize is the number of results per page.
	PageSize *int
}

// List returns the compliance questions in paginated format.
func (s *VerificationQuestionsService) List(ctx context.Context, params VerificationQuestionListParams, opts ...RequestOption) (*Page[VerificationQuestion], error) {
	q := url.Values{}
	setString(q, "flow__type", string(params.FlowType))
	setString(q, "flow__state", string(params.FlowState))
	setBool(q, "answer__isnull", params.AnswerIsNull)
	setTime(q, "flow__started_at__gte", params.FlowStartedAtGTE)
	setTime(q, "flow__started_at__lte", params.FlowStartedAtLTE)
	setString(q, "question", params.Question)
	setString(q, "answer__answer", params.AnswerAnswer)
	setInt(q, "page", params.Page)
	setInt(q, "page_size", params.PageSize)
	return getPage[VerificationQuestion](ctx, s.client, "/api/v2/verification/question/", q, opts)
}

// ListAll returns a lazy iterator over every verification question across
// all pages.
func (s *VerificationQuestionsService) ListAll(ctx context.Context, params VerificationQuestionListParams, opts ...RequestOption) iter.Seq2[VerificationQuestion, error] {
	return iterPages(ctx, func(ctx context.Context) (*Page[VerificationQuestion], error) {
		return s.List(ctx, params, opts...)
	})
}

// SubmitAnswerParams are the parameters for
// VerificationQuestionsService.SubmitAnswer.
type SubmitAnswerParams struct {
	// Answer is the answer to the question.
	Answer string
	// Files lists optional attachments to submit with the answer.
	Files []File
}

// SubmitAnswer submits an answer for a verification question with optional
// attachments. The request is encoded as multipart/form-data.
func (s *VerificationQuestionsService) SubmitAnswer(ctx context.Context, questionID int, params SubmitAnswerParams, opts ...RequestOption) (*VerificationAnswer, error) {
	fields := map[string]string{}
	if params.Answer != "" {
		fields["answer"] = params.Answer
	}
	body, contentType, err := encodeMultipart(fields, "files", params.Files)
	if err != nil {
		return nil, err
	}
	out := &VerificationAnswer{}
	path := "/api/v2/verification/question/" + strconv.Itoa(questionID) + "/answer/"
	if err := s.client.doMultipart(ctx, http.MethodPost, path, body, contentType, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// AppealState is the state of a suspension appeal.
type AppealState string

// Appeal states.
const (
	// AppealApproved means the appeal was approved.
	AppealApproved AppealState = "approved"
	// AppealRejected means the appeal was rejected.
	AppealRejected AppealState = "rejected"
	// AppealSubmitted means the appeal is awaiting review.
	AppealSubmitted AppealState = "submitted"
)

// VerificationAppeal is a suspension appeal.
type VerificationAppeal struct {
	// ID is the unique identifier of the appeal.
	ID int `json:"id"`
	// Appeal is the appeal text.
	Appeal string `json:"appeal"`
	// State is the appeal state.
	State AppealState `json:"state"`
	// CreatedAt is when the appeal was created.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is when the appeal was last updated.
	UpdatedAt time.Time `json:"updated_at"`
}

// VerificationAppealsService exposes the suspension appeal operations.
type VerificationAppealsService struct {
	client *Client
}

// VerificationAppealListParams are the parameters for
// VerificationAppealsService.List.
type VerificationAppealListParams struct {
	// State matches only appeals in the given state.
	State AppealState
	// Page is the page number.
	Page *int
	// PageSize is the number of results per page.
	PageSize *int
}

// List returns the appeals submitted for the account in paginated format.
func (s *VerificationAppealsService) List(ctx context.Context, params VerificationAppealListParams, opts ...RequestOption) (*Page[VerificationAppeal], error) {
	q := url.Values{}
	setString(q, "state", string(params.State))
	setInt(q, "page", params.Page)
	setInt(q, "page_size", params.PageSize)
	return getPage[VerificationAppeal](ctx, s.client, "/api/v2/verification/appeal/", q, opts)
}

// ListAll returns a lazy iterator over every appeal across all pages.
func (s *VerificationAppealsService) ListAll(ctx context.Context, params VerificationAppealListParams, opts ...RequestOption) iter.Seq2[VerificationAppeal, error] {
	return iterPages(ctx, func(ctx context.Context) (*Page[VerificationAppeal], error) {
		return s.List(ctx, params, opts...)
	})
}

// VerificationAppealCreateParams are the parameters for
// VerificationAppealsService.Create.
type VerificationAppealCreateParams struct {
	// Appeal is the appeal text to submit for the account suspension.
	Appeal string `json:"appeal"`
}

// Create submits an appeal for an account suspension. Only one appeal can be
// submitted at a time.
func (s *VerificationAppealsService) Create(ctx context.Context, params VerificationAppealCreateParams, opts ...RequestOption) (*VerificationAppeal, error) {
	out := &VerificationAppeal{}
	if err := s.client.doJSON(ctx, http.MethodPost, "/api/v2/verification/appeal/", nil, params, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// AbuseReport is an abuse report raised against the account.
type AbuseReport struct {
	// ID is the unique identifier of the abuse report.
	ID int `json:"id"`
	// Content is the content of the abuse report.
	Content string `json:"content"`
	// Flow is the related account verification flow ID. May be nil.
	Flow *int `json:"flow"`
	// CreatedAt is when the report was created.
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt is when the report was last updated.
	UpdatedAt time.Time `json:"updated_at"`
}

// VerificationAbuseReportsService exposes the abuse report operations.
type VerificationAbuseReportsService struct {
	client *Client
}

// AbuseReportListParams are the parameters for
// VerificationAbuseReportsService.List.
type AbuseReportListParams struct {
	// Page is the page number.
	Page *int
	// PageSize is the number of results per page.
	PageSize *int
}

// List returns the abuse reports raised against the account in paginated
// format.
func (s *VerificationAbuseReportsService) List(ctx context.Context, params AbuseReportListParams, opts ...RequestOption) (*Page[AbuseReport], error) {
	q := url.Values{}
	setInt(q, "page", params.Page)
	setInt(q, "page_size", params.PageSize)
	return getPage[AbuseReport](ctx, s.client, "/api/v2/verification/abuse_report/", q, opts)
}

// ListAll returns a lazy iterator over every abuse report across all pages.
func (s *VerificationAbuseReportsService) ListAll(ctx context.Context, params AbuseReportListParams, opts ...RequestOption) iter.Seq2[AbuseReport, error] {
	return iterPages(ctx, func(ctx context.Context) (*Page[AbuseReport], error) {
		return s.List(ctx, params, opts...)
	})
}

// Suspension describes when and why the account was suspended.
type Suspension struct {
	// CreatedAt is when the account was suspended.
	CreatedAt time.Time `json:"created_at"`
	// Reason is the suspension reason.
	Reason VerificationFlowType `json:"reason"`
}

// GetSuspension returns when the account was suspended and why. This
// endpoint works even while the account is suspended.
func (s *VerificationService) GetSuspension(ctx context.Context, opts ...RequestOption) (*Suspension, error) {
	out := &Suspension{}
	if err := s.client.doJSON(ctx, http.MethodGet, "/api/v2/verification/suspension/", nil, nil, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// VerificationCategory describes one verification category.
type VerificationCategory struct {
	// Description describes the category. Can be user visible.
	Description string `json:"description"`
	// RequestThreshold previously held the triggering request threshold;
	// it is now always nil.
	RequestThreshold *int `json:"request_threshold"`
	// IDVerificationRequired reports whether the category requires ID
	// verification when triggered.
	IDVerificationRequired bool `json:"id_verification_required"`
	// IDVerificationRestoresAccess reports whether completing ID
	// verification restores proxy access.
	IDVerificationRestoresAccess bool `json:"id_verification_restores_access"`
}

// GetCategories retrieves the verification categories that may trigger
// verification flows. The response is a map keyed by category name.
func (s *VerificationService) GetCategories(ctx context.Context, opts ...RequestOption) (map[string]VerificationCategory, error) {
	out := map[string]VerificationCategory{}
	if err := s.client.doJSON(ctx, http.MethodGet, "/api/v2/verification/categories/", nil, nil, &out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// ProxyState describes the verification limit applied to the account's
// proxies.
type ProxyState string

// Proxy states.
const (
	// ProxyStateActive means proxies work normally.
	ProxyStateActive ProxyState = "active"
	// ProxyStateLimited means proxies are slower than usual and may error.
	ProxyStateLimited ProxyState = "limited"
	// ProxyStatePaused means proxies are currently not working.
	ProxyStatePaused ProxyState = "paused"
)

// VerificationLimits describes limits the account may have received.
type VerificationLimits struct {
	// ProxyState is the current proxy limit state.
	ProxyState ProxyState `json:"proxy_state"`
}

// GetLimits retrieves the verification limits applied to the account.
func (s *VerificationService) GetLimits(ctx context.Context, opts ...RequestOption) (*VerificationLimits, error) {
	out := &VerificationLimits{}
	if err := s.client.doJSON(ctx, http.MethodGet, "/api/v2/verification/limits/", nil, nil, out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// VerificationThreshold describes one verification trigger threshold.
type VerificationThreshold struct {
	// Description describes the trigger. Can be user visible.
	Description string `json:"description"`
	// IDVerificationRequired reports whether the threshold requires ID
	// verification when triggered.
	IDVerificationRequired bool `json:"id_verification_required"`
	// IDVerificationRestoresAccess reports whether completing ID
	// verification restores proxy access.
	IDVerificationRestoresAccess bool `json:"id_verification_restores_access"`
	// RequestCount is the number of proxy requests matching this trigger.
	RequestCount int `json:"request_count"`
	// RequestThreshold previously held the triggering request count; it is
	// now always nil.
	RequestThreshold *int `json:"request_threshold"`
	// Triggered reports whether the threshold has been triggered.
	Triggered bool `json:"triggered"`
}

// GetThresholds retrieves the thresholds that may trigger an acceptable use
// verification flow. The response is a map keyed by category name.
func (s *VerificationService) GetThresholds(ctx context.Context, opts ...RequestOption) (map[string]VerificationThreshold, error) {
	out := map[string]VerificationThreshold{}
	if err := s.client.doJSON(ctx, http.MethodGet, "/api/v2/verification/thresholds/", nil, nil, &out, opts); err != nil {
		return nil, err
	}
	return out, nil
}

// encodeMultipart builds a multipart/form-data body from text fields and
// files.
func encodeMultipart(fields map[string]string, fileField string, files []File) ([]byte, string, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			return nil, "", fmt.Errorf("webshare: encoding multipart field %q: %w", key, err)
		}
	}
	for _, file := range files {
		part, err := writer.CreateFormFile(fileField, file.Name)
		if err != nil {
			return nil, "", fmt.Errorf("webshare: encoding multipart file %q: %w", file.Name, err)
		}
		if _, err := io.Copy(part, file.Reader); err != nil {
			return nil, "", fmt.Errorf("webshare: reading multipart file %q: %w", file.Name, err)
		}
	}
	if err := writer.Close(); err != nil {
		return nil, "", fmt.Errorf("webshare: finalizing multipart body: %w", err)
	}
	return buf.Bytes(), writer.FormDataContentType(), nil
}
