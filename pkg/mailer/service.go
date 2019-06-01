/**
 * Copyright (c) 2019 Adrian K <adrian.git@kuguar.dev>
 *
 * This software is released under the MIT License.
 * https://opensource.org/licenses/MIT
 */

package mailer

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/adrianpk/poslan/internal/config"
	"github.com/adrianpk/poslan/internal/sys"
	"github.com/adrianpk/poslan/pkg/auth"
	"github.com/adrianpk/poslan/pkg/model"
	"github.com/go-kit/kit/log"
	"github.com/google/uuid"
	health "github.com/heptiolabs/healthcheck"
)

type service struct {
	mux       sync.Mutex
	name      string
	ctx       context.Context
	cfg       *config.Config
	logger    log.Logger
	auth      auth.SecServer
	providers []sys.Provider
	health    health.Handler
	ready     bool
	alive     bool
}

// SignIn lets a user sign in providing username and password.
func (s *service) SignIn(ctx context.Context, clientID, secret string) (string, error) {
	output, err := s.auth.Authenticate(clientID, secret)
	if err != nil {
		return "", err
	}
	return output, nil
}

// SignOut lets a user sign out.
func (s *service) SignOut(ctx context.Context, id uuid.UUID) error {
	// TODO: Close session implementation.
	return errors.New("not implemented")
}

// Send lets the user send a mail.
func (s *service) Send(ctx context.Context, to, cc, bcc, subject, body string) error {
	// FIX: take following two values from those stored in context
	// They are set by authentication middleware.
	fromName := s.Config().Mailer.Providers[0].Sender.Name
	fromEmail := s.Config().Mailer.Providers[0].Sender.Email

	p1, ok := s.ProviderByPriority(1)

	if !ok {
		return errors.New("no providers configured")
	}

	p2, ok2 := s.ProviderByPriority(2)

	e := makeEmail(fromName, fromEmail, to, cc, bcc, subject, body)
	resend, err := p1.Send(e)

	// Previous atempt failed and and a second provider enabled.
	if resend && ok2 || true {
		resend, err = p2.Send(e)
	}

	if err != nil {
		msg := fmt.Sprintf("Cannot send email with ID: '%s'.", e.ID)
		s.logger.Log(
			"level", config.LogLevel.Error,
			"package", "mailer",
			"method", "Send",
			"message", msg,
			"error", err.Error(),
		)
	}

	return err
}

// Providers returns service providers.
func (s *service) Providers() []sys.Provider {
	return s.providers
}

// Enable set to true the ready state.
func (s *service) Enable() {
	s.ready = true
}

// Disable set to false the ready state.
func (s *service) Disable() {
	s.ready = false
}

// IsAlive return true if the service is ready to serve requests.
// This function can be used by a readiness checker.
func (s *service) IsReady() bool {
	return s.ready
}

// IsAlive return true if the service is alive.
// This function can be used by a liveness checker.
func (s *service) IsAlive() bool {
	return s.alive
}

// Misc
// Context returns service context.
func (s *service) Context() context.Context {
	return s.ctx
}

// Config returns service config.
func (s *service) Config() *config.Config {
	return s.cfg
}

// Logger returns service imterface implemention logger.
func (s *service) Logger() log.Logger {
	return s.logger
}

// ProviderByPriority returns a provider by
// its prioririty (1..n)
func (s *service) ProviderByPriority(priority int) (p sys.Provider, ok bool) {
	for _, p := range s.providers {
		if priority == p.Priority() {
			return p, true
		}
	}
	return s.FallbackProvider()
}

// FallbackProvider returns the default provider
// if it was provided in config.
func (s *service) FallbackProvider() (p sys.Provider, ok bool) {
	if len(s.providers) > 0 {
		return s.providers[0], true
	}
	return nil, false
}

// Utility functions
func makeEmail(name, from, to, cc, bcc, subject, body string) *model.Email {
	return &model.Email{
		ID:      uuid.New(),
		Name:    name,
		From:    from,
		To:      to,
		CC:      cc,
		BCC:     bcc,
		Subject: subject,
		Body:    body,
		Charset: charset,
	}
}
