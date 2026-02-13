package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/jeheskielSunloy77/libra-link/apps/tui/internal/api/gen"
)

const (
	defaultAccessCookieName  = "access_token"
	defaultRefreshCookieName = "refresh_token"
)

type Client struct {
	client    *gen.ClientWithResponses
	http      *http.Client
	baseURL   *url.URL
	access    string
	refresh   string
	userID    string
	accessCK  string
	refreshCK string
	mu        sync.RWMutex
}

func NewClient(baseURL string, timeout time.Duration) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{
		Timeout: timeout,
		Jar:     jar,
	}

	parsedBase, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	genClient, err := gen.NewClientWithResponses(baseURL, gen.WithHTTPClient(httpClient))
	if err != nil {
		return nil, err
	}

	return &Client{
		client:    genClient,
		http:      httpClient,
		baseURL:   parsedBase,
		accessCK:  defaultAccessCookieName,
		refreshCK: defaultRefreshCookieName,
	}, nil
}

func (c *Client) SetSession(accessToken, refreshToken, userID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.access = accessToken
	c.refresh = refreshToken
	c.userID = userID
	c.writeCookie(c.accessCK, accessToken)
	c.writeCookie(c.refreshCK, refreshToken)
}

func (c *Client) Session() (accessToken, refreshToken, userID string) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.access, c.refresh, c.userID
}

func (c *Client) ClearSession() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.access = ""
	c.refresh = ""
	c.userID = ""
	c.writeCookie(c.accessCK, "")
	c.writeCookie(c.refreshCK, "")
}

func (c *Client) Register(ctx context.Context, email, username, password string) (*User, error) {
	resp, err := c.client.AuthRegisterWithResponse(ctx, gen.AuthRegisterJSONRequestBody{
		Email:    openapi_types.Email(email),
		Username: username,
		Password: password,
	})
	if err != nil {
		return nil, err
	}
	if resp.JSON201 == nil {
		return nil, apiError("register", resp.StatusCode(), resp.Body)
	}

	c.captureTokens(resp.HTTPResponse)
	user := &User{ID: resp.JSON201.Id.String(), Email: string(resp.JSON201.Email), Username: resp.JSON201.Username}
	c.setUserID(user.ID)
	return user, nil
}

func (c *Client) Login(ctx context.Context, identifier, password string) (*User, error) {
	resp, err := c.client.AuthLoginWithResponse(ctx, gen.AuthLoginJSONRequestBody{
		Identifier: identifier,
		Password:   password,
	})
	if err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, apiError("login", resp.StatusCode(), resp.Body)
	}

	c.captureTokens(resp.HTTPResponse)
	user := &User{ID: resp.JSON200.Id.String(), Email: string(resp.JSON200.Email), Username: resp.JSON200.Username}
	c.setUserID(user.ID)
	return user, nil
}

func (c *Client) Refresh(ctx context.Context) (*User, error) {
	resp, err := c.client.AuthRefreshWithResponse(ctx, gen.AuthRefreshJSONRequestBody{}, c.withRefreshCookie())
	if err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, apiError("refresh", resp.StatusCode(), resp.Body)
	}

	c.captureTokens(resp.HTTPResponse)
	user := &User{ID: resp.JSON200.Id.String(), Email: string(resp.JSON200.Email), Username: resp.JSON200.Username}
	c.setUserID(user.ID)
	return user, nil
}

func (c *Client) Logout(ctx context.Context) error {
	resp, err := c.client.AuthLogoutWithResponse(ctx, gen.AuthLogoutJSONRequestBody{}, c.withBearer(), c.withRefreshCookie())
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		return apiError("logout", resp.StatusCode(), resp.Body)
	}
	c.ClearSession()
	return nil
}

func (c *Client) Me(ctx context.Context) (*User, error) {
	resp, err := c.client.AuthMeWithResponse(ctx, c.withBearer())
	if err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, apiError("me", resp.StatusCode(), resp.Body)
	}
	user := &User{ID: resp.JSON200.Id.String(), Email: string(resp.JSON200.Email), Username: resp.JSON200.Username}
	c.setUserID(user.ID)
	return user, nil
}

func (c *Client) GoogleAuthURL() string {
	if c.baseURL == nil {
		return ""
	}
	u := *c.baseURL
	u.Path = strings.TrimSuffix(u.Path, "/") + "/api/v1/auth/google"
	return u.String()
}

func (c *Client) StartGoogleDeviceAuth(ctx context.Context) (*GoogleDeviceStart, error) {
	resp, err := c.client.AuthGoogleDeviceStartWithResponse(ctx, gen.AuthGoogleDeviceStartJSONRequestBody{})
	if err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, apiError("google device start", resp.StatusCode(), resp.Body)
	}
	return &GoogleDeviceStart{
		DeviceCode:      resp.JSON200.DeviceCode,
		AuthURL:         resp.JSON200.AuthUrl,
		ExpiresAt:       resp.JSON200.ExpiresAt,
		IntervalSeconds: resp.JSON200.IntervalSeconds,
	}, nil
}

func (c *Client) PollGoogleDeviceAuth(ctx context.Context, deviceCode string) (*GoogleDevicePoll, error) {
	resp, err := c.client.AuthGoogleDevicePollWithResponse(ctx, gen.AuthGoogleDevicePollJSONRequestBody{
		DeviceCode: deviceCode,
	})
	if err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, apiError("google device poll", resp.StatusCode(), resp.Body)
	}

	result := &GoogleDevicePoll{Status: string(resp.JSON200.Status)}
	if resp.JSON200.Result != nil {
		user := &User{
			ID:       resp.JSON200.Result.User.Id.String(),
			Email:    string(resp.JSON200.Result.User.Email),
			Username: resp.JSON200.Result.User.Username,
		}
		result.User = user
		result.AccessToken = resp.JSON200.Result.Token.Token
		result.RefreshToken = resp.JSON200.Result.RefreshToken.Token
		c.SetSession(result.AccessToken, result.RefreshToken, user.ID)
	}
	return result, nil
}

func (c *Client) ListEbooks(ctx context.Context, limit int) ([]Ebook, error) {
	if limit <= 0 {
		limit = 20
	}
	params := &gen.EbookGetManyParams{Limit: &limit}
	resp, err := c.client.EbookGetManyWithResponse(ctx, params, c.withBearer())
	if err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, apiError("list ebooks", resp.StatusCode(), resp.Body)
	}

	items := make([]Ebook, 0, len(resp.JSON200.Data))
	for _, row := range resp.JSON200.Data {
		item := Ebook{
			ID:         row.Id.String(),
			Title:      row.Title,
			Format:     string(row.Format),
			StorageKey: row.StorageKey,
			ImportedAt: row.ImportedAt,
		}
		if row.Description != nil {
			item.Description = *row.Description
		}
		items = append(items, item)
	}
	return items, nil
}

func (c *Client) ListShares(ctx context.Context, limit int) ([]Share, error) {
	if limit <= 0 {
		limit = 20
	}
	params := &gen.ShareGetManyParams{Limit: &limit}
	resp, err := c.client.ShareGetManyWithResponse(ctx, params, c.withBearer())
	if err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, apiError("list shares", resp.StatusCode(), resp.Body)
	}

	items := make([]Share, 0, len(resp.JSON200.Data))
	for _, row := range resp.JSON200.Data {
		entry := Share{
			ID:          row.Id.String(),
			EbookID:     row.EbookId.String(),
			OwnerUserID: row.OwnerUserId.String(),
			Status:      string(row.Status),
			Visibility:  string(row.Visibility),
		}
		if row.TitleOverride != nil {
			entry.Title = *row.TitleOverride
		}
		items = append(items, entry)
	}
	return items, nil
}

func (c *Client) BorrowShare(ctx context.Context, shareID string) error {
	uuidVal, err := parseUUID(shareID)
	if err != nil {
		return err
	}
	resp, err := c.client.ShareBorrowWithResponse(ctx, uuidVal, gen.ShareBorrowJSONRequestBody{LegalAcknowledged: gen.True}, c.withBearer())
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusCreated {
		return apiError("borrow share", resp.StatusCode(), resp.Body)
	}
	return nil
}

func (c *Client) UpsertReview(ctx context.Context, shareID string, rating int, review string) error {
	uuidVal, err := parseUUID(shareID)
	if err != nil {
		return err
	}
	body := gen.ShareUpsertReviewJSONRequestBody{Rating: rating}
	if strings.TrimSpace(review) != "" {
		body.ReviewText = &review
	}
	resp, err := c.client.ShareUpsertReviewWithResponse(ctx, uuidVal, body, c.withBearer())
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusOK {
		return apiError("upsert review", resp.StatusCode(), resp.Body)
	}
	return nil
}

func (c *Client) ReportShare(ctx context.Context, shareID, reason, details string) error {
	uuidVal, err := parseUUID(shareID)
	if err != nil {
		return err
	}
	r := gen.ShareCreateReportJSONBodyReason(reason)
	body := gen.ShareCreateReportJSONRequestBody{Reason: r}
	if strings.TrimSpace(details) != "" {
		body.Details = &details
	}
	resp, err := c.client.ShareCreateReportWithResponse(ctx, uuidVal, body, c.withBearer())
	if err != nil {
		return err
	}
	if resp.StatusCode() != http.StatusCreated {
		return apiError("report share", resp.StatusCode(), resp.Body)
	}
	return nil
}

func (c *Client) GetPreferences(ctx context.Context) (*Preferences, error) {
	resp, err := c.client.ReaderGetUserPreferencesWithResponse(ctx, c.withBearer())
	if err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, apiError("get preferences", resp.StatusCode(), resp.Body)
	}
	data := resp.JSON200.Data
	prefs := &Preferences{
		UserID:            data.UserId.String(),
		ReadingMode:       string(data.ReadingMode),
		ZenRestoreOnOpen:  data.ZenRestoreOnOpen,
		ThemeMode:         string(data.ThemeMode),
		ThemeOverrides:    data.ThemeOverrides,
		TypographyProfile: string(data.TypographyProfile),
		RowVersion:        data.RowVersion,
	}
	return prefs, nil
}

func (c *Client) PatchPreferences(ctx context.Context, input Preferences) (*Preferences, error) {
	mode := gen.ReaderPatchUserPreferencesJSONBodyReadingMode(input.ReadingMode)
	theme := gen.ReaderPatchUserPreferencesJSONBodyThemeMode(input.ThemeMode)
	typo := gen.ReaderPatchUserPreferencesJSONBodyTypographyProfile(input.TypographyProfile)
	body := gen.ReaderPatchUserPreferencesJSONRequestBody{
		ReadingMode:       &mode,
		ZenRestoreOnOpen:  &input.ZenRestoreOnOpen,
		ThemeMode:         &theme,
		ThemeOverrides:    &input.ThemeOverrides,
		TypographyProfile: &typo,
	}

	resp, err := c.client.ReaderPatchUserPreferencesWithResponse(ctx, body, c.withBearer())
	if err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, apiError("patch preferences", resp.StatusCode(), resp.Body)
	}
	data := resp.JSON200.Data
	patched := &Preferences{
		UserID:            data.UserId.String(),
		ReadingMode:       string(data.ReadingMode),
		ZenRestoreOnOpen:  data.ZenRestoreOnOpen,
		ThemeMode:         string(data.ThemeMode),
		ThemeOverrides:    data.ThemeOverrides,
		TypographyProfile: string(data.TypographyProfile),
		RowVersion:        data.RowVersion,
	}
	return patched, nil
}

func (c *Client) GetReaderState(ctx context.Context) (*ReaderState, error) {
	resp, err := c.client.ReaderGetUserReaderStateWithResponse(ctx, c.withBearer())
	if err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, apiError("get reader state", resp.StatusCode(), resp.Body)
	}
	data := resp.JSON200.Data
	state := &ReaderState{
		UserID:      data.UserId.String(),
		ReadingMode: string(data.ReadingMode),
		RowVersion:  data.RowVersion,
	}
	if data.CurrentEbookId != nil {
		state.CurrentEbookID = data.CurrentEbookId.String()
	}
	if data.CurrentLocation != nil {
		state.CurrentLocation = *data.CurrentLocation
	}
	if data.LastOpenedAt != nil {
		state.LastOpenedAt = data.LastOpenedAt
	}
	return state, nil
}

func (c *Client) PatchReaderState(ctx context.Context, input ReaderState) (*ReaderState, error) {
	mode := gen.ReaderPatchUserReaderStateJSONBodyReadingMode(input.ReadingMode)
	body := gen.ReaderPatchUserReaderStateJSONRequestBody{ReadingMode: &mode}

	if strings.TrimSpace(input.CurrentLocation) != "" {
		body.CurrentLocation = &input.CurrentLocation
	}
	if strings.TrimSpace(input.CurrentEbookID) != "" {
		parsed, err := parseUUID(input.CurrentEbookID)
		if err != nil {
			return nil, err
		}
		body.CurrentEbookId = &parsed
	}
	if input.LastOpenedAt != nil {
		body.LastOpenedAt = input.LastOpenedAt
	}

	resp, err := c.client.ReaderPatchUserReaderStateWithResponse(ctx, body, c.withBearer())
	if err != nil {
		return nil, err
	}
	if resp.JSON200 == nil {
		return nil, apiError("patch reader state", resp.StatusCode(), resp.Body)
	}
	data := resp.JSON200.Data
	state := &ReaderState{
		UserID:      data.UserId.String(),
		ReadingMode: string(data.ReadingMode),
		RowVersion:  data.RowVersion,
	}
	if data.CurrentEbookId != nil {
		state.CurrentEbookID = data.CurrentEbookId.String()
	}
	if data.CurrentLocation != nil {
		state.CurrentLocation = *data.CurrentLocation
	}
	if data.LastOpenedAt != nil {
		state.LastOpenedAt = data.LastOpenedAt
	}
	return state, nil
}

func (c *Client) StoreSyncEvent(ctx context.Context, input SyncEvent) error {
	entityType := gen.SyncStoreEventJSONBodyEntityType(input.EntityType)
	operation := gen.SyncStoreEventJSONBodyOperation(input.Operation)
	entityID, err := parseUUID(input.EntityID)
	if err != nil {
		return err
	}
	body := gen.SyncStoreEventJSONRequestBody{
		EntityType:      entityType,
		EntityId:        entityID,
		Operation:       operation,
		IdempotencyKey:  input.IdempotencyKey,
		ClientTimestamp: &input.ClientTS,
	}

	if input.BaseVersion != nil {
		body.BaseVersion = input.BaseVersion
	}
	if input.Payload != nil {
		payload := make(map[string]*interface{}, len(input.Payload))
		for key, value := range input.Payload {
			v := value
			payload[key] = &v
		}
		body.Payload = &payload
	}

	resp, err := c.client.SyncStoreEventWithResponse(ctx, body, c.withBearer())
	if err != nil {
		return err
	}
	if resp.JSON201 == nil {
		return apiError("store sync event", resp.StatusCode(), resp.Body)
	}
	return nil
}

func (c *Client) captureTokens(httpResp *http.Response) {
	if httpResp == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, cookie := range httpResp.Cookies() {
		nameLower := strings.ToLower(cookie.Name)
		switch {
		case strings.Contains(nameLower, "access"):
			c.accessCK = cookie.Name
			c.access = cookie.Value
		case strings.Contains(nameLower, "refresh"):
			c.refreshCK = cookie.Name
			c.refresh = cookie.Value
		}
	}

	if c.access != "" {
		c.writeCookie(c.accessCK, c.access)
	}
	if c.refresh != "" {
		c.writeCookie(c.refreshCK, c.refresh)
	}
}

func (c *Client) setUserID(userID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.userID = userID
}

func (c *Client) writeCookie(name, value string) {
	if c.http == nil || c.http.Jar == nil || c.baseURL == nil || name == "" {
		return
	}
	cookie := &http.Cookie{Name: name, Value: value, Path: "/"}
	if value == "" {
		cookie.MaxAge = -1
	}
	c.http.Jar.SetCookies(c.baseURL, []*http.Cookie{cookie})
}

func (c *Client) withBearer() gen.RequestEditorFn {
	return func(_ context.Context, req *http.Request) error {
		c.mu.RLock()
		token := c.access
		c.mu.RUnlock()
		if token == "" {
			return errors.New("missing access token")
		}
		req.Header.Set("Authorization", "Bearer "+token)
		return nil
	}
}

func (c *Client) withRefreshCookie() gen.RequestEditorFn {
	return func(_ context.Context, req *http.Request) error {
		c.mu.RLock()
		name := c.refreshCK
		token := c.refresh
		c.mu.RUnlock()
		if name == "" {
			name = defaultRefreshCookieName
		}
		if token == "" {
			return errors.New("missing refresh token")
		}
		req.Header.Add("Cookie", fmt.Sprintf("%s=%s", name, token))
		return nil
	}
}

func parseUUID(raw string) (openapi_types.UUID, error) {
	parsed, err := uuid.Parse(raw)
	if err != nil {
		return openapi_types.UUID{}, fmt.Errorf("invalid uuid %q", raw)
	}
	return openapi_types.UUID(parsed), nil
}

func apiError(operation string, status int, body []byte) error {
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		trimmed = http.StatusText(status)
	}
	if len(trimmed) > 280 {
		trimmed = trimmed[:280] + "..."
	}
	return fmt.Errorf("%s failed (%d): %s", operation, status, trimmed)
}
