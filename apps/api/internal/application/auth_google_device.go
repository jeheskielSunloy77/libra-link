package application

import (
	"context"
	"strings"
	"time"

	"github.com/jeheskielSunloy77/libra-link/internal/app/errs"
)

func (s *authService) StartGoogleDeviceAuth(ctx context.Context) (*GoogleDeviceAuthStart, error) {
	if !s.googleConfigReady() {
		return nil, errs.NewBadRequestError("Google login is not configured", false, nil, nil)
	}

	deviceCode, err := generateStateToken()
	if err != nil {
		return nil, errs.NewInternalServerError()
	}
	state, err := generateStateToken()
	if err != nil {
		return nil, errs.NewInternalServerError()
	}

	now := s.currentTime()
	expiresAt := now.Add(googleDevicePendingTTL)
	authURL := s.googleOAuthConfig.AuthCodeURL(state)

	s.deviceStateMu.Lock()
	defer s.deviceStateMu.Unlock()

	s.cleanupDeviceSessionsLocked(now)
	s.deviceAuthStates[deviceCode] = &googleDeviceSession{
		DeviceCode: deviceCode,
		State:      state,
		ExpiresAt:  expiresAt,
		Status:     GoogleDeviceAuthPending,
	}
	s.deviceStateIndex[state] = deviceCode

	interval := int(s.devicePollInterval.Seconds())
	if interval < 1 {
		interval = 1
	}

	return &GoogleDeviceAuthStart{
		DeviceCode:      deviceCode,
		AuthURL:         authURL,
		ExpiresAt:       expiresAt,
		IntervalSeconds: interval,
	}, nil
}

func (s *authService) CompleteGoogleDeviceAuth(ctx context.Context, code, state, userAgent, ipAddress string) error {
	if !s.googleConfigReady() {
		return errs.NewBadRequestError("Google login is not configured", false, nil, nil)
	}
	if strings.TrimSpace(code) == "" || strings.TrimSpace(state) == "" {
		return errs.NewBadRequestError("Invalid Google login request", false, nil, nil)
	}

	deviceCode, session, err := s.getDeviceSessionByState(state)
	if err != nil {
		return err
	}

	token, err := s.googleOAuthConfig.Exchange(ctx, code)
	if err != nil {
		s.markDeviceSessionFailed(deviceCode, session, "exchange failed")
		return errs.NewUnauthorizedError("Invalid Google token", false)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || strings.TrimSpace(rawIDToken) == "" {
		s.markDeviceSessionFailed(deviceCode, session, "missing id token")
		return errs.NewUnauthorizedError("Invalid Google token", false)
	}

	claims, err := s.googleTokenValidator(ctx, rawIDToken, s.googleClientID)
	if err != nil {
		s.markDeviceSessionFailed(deviceCode, session, "token validation failed")
		return errs.NewUnauthorizedError("Invalid Google token", false)
	}

	result, err := s.loginWithGoogleClaims(
		ctx,
		claims.Subject,
		stringClaim(claims.Claims, "email"),
		boolClaim(claims.Claims, "email_verified"),
		userAgent,
		ipAddress,
	)
	if err != nil {
		s.markDeviceSessionFailed(deviceCode, session, err.Error())
		return err
	}

	s.deviceStateMu.Lock()
	session.Status = GoogleDeviceAuthApproved
	session.Result = result
	session.LastError = ""
	s.deviceAuthStates[deviceCode] = session
	s.deviceStateMu.Unlock()

	return nil
}

func (s *authService) PollGoogleDeviceAuth(ctx context.Context, deviceCode string) (*GoogleDeviceAuthPollResult, error) {
	_ = ctx
	if strings.TrimSpace(deviceCode) == "" {
		return nil, errs.NewBadRequestError("deviceCode is required", true, nil, nil)
	}

	now := s.currentTime()

	s.deviceStateMu.Lock()
	defer s.deviceStateMu.Unlock()

	s.cleanupDeviceSessionsLocked(now)

	session, ok := s.deviceAuthStates[deviceCode]
	if !ok {
		return nil, errs.NewBadRequestError("device code is invalid or expired", true, nil, nil)
	}

	if session.ExpiresAt.Before(now) {
		session.Status = GoogleDeviceAuthExpired
		s.deviceAuthStates[deviceCode] = session
	}

	switch session.Status {
	case GoogleDeviceAuthPending:
		return &GoogleDeviceAuthPollResult{Status: GoogleDeviceAuthPending}, nil
	case GoogleDeviceAuthExpired:
		return &GoogleDeviceAuthPollResult{Status: GoogleDeviceAuthExpired}, nil
	case GoogleDeviceAuthFailed:
		return &GoogleDeviceAuthPollResult{Status: GoogleDeviceAuthFailed}, nil
	case GoogleDeviceAuthApproved:
		result := session.Result
		delete(s.deviceStateIndex, session.State)
		delete(s.deviceAuthStates, deviceCode)
		return &GoogleDeviceAuthPollResult{Status: GoogleDeviceAuthApproved, Result: result}, nil
	default:
		return &GoogleDeviceAuthPollResult{Status: GoogleDeviceAuthPending}, nil
	}
}

func (s *authService) getDeviceSessionByState(state string) (string, *googleDeviceSession, error) {
	now := s.currentTime()

	s.deviceStateMu.Lock()
	defer s.deviceStateMu.Unlock()

	s.cleanupDeviceSessionsLocked(now)

	deviceCode, ok := s.deviceStateIndex[state]
	if !ok {
		return "", nil, errs.NewBadRequestError("Invalid Google login state", false, nil, nil)
	}

	session, ok := s.deviceAuthStates[deviceCode]
	if !ok {
		return "", nil, errs.NewBadRequestError("Invalid Google login state", false, nil, nil)
	}

	if session.ExpiresAt.Before(now) {
		session.Status = GoogleDeviceAuthExpired
		s.deviceAuthStates[deviceCode] = session
		return "", nil, errs.NewBadRequestError("Device authorization expired", false, nil, nil)
	}

	return deviceCode, session, nil
}

func (s *authService) markDeviceSessionFailed(deviceCode string, session *googleDeviceSession, reason string) {
	s.deviceStateMu.Lock()
	defer s.deviceStateMu.Unlock()
	session.Status = GoogleDeviceAuthFailed
	session.LastError = reason
	s.deviceAuthStates[deviceCode] = session
}

func (s *authService) cleanupDeviceSessionsLocked(now time.Time) {
	for code, session := range s.deviceAuthStates {
		if session.ExpiresAt.Add(30 * time.Minute).Before(now) {
			delete(s.deviceStateIndex, session.State)
			delete(s.deviceAuthStates, code)
		}
	}
}

func (s *authService) currentTime() time.Time {
	if s.now != nil {
		return s.now().UTC()
	}
	return time.Now().UTC()
}

func stringClaim(claims map[string]interface{}, key string) string {
	if claims == nil {
		return ""
	}
	if value, ok := claims[key].(string); ok {
		return value
	}
	return ""
}

func boolClaim(claims map[string]interface{}, key string) bool {
	if claims == nil {
		return false
	}
	if value, ok := claims[key].(bool); ok {
		return value
	}
	return false
}
