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
	"golang.org/x/crypto/bcrypt"
)

type service struct {
	mux       sync.Mutex
	name      string
	ctx       context.Context
	cfg       *config.Config
	logger    log.Logger
	auth      auth.SecServer
	providers []sys.Provider
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

	fromName := s.Config().Mailers.Providers[0].Name
	fromEmail := s.Config().Mailers.Providers[0].Name

	m := makeEmail(fromName, fromEmail, to, cc, bcc, subject, body)
	s.logger.Log("message", fmt.Sprintf("%+v", m))

	return errors.New("not implemented")
}

// Providers returns service providers.
func (s *service) Providers() []sys.Provider {
	return s.providers
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

func passwordMatches(digest, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(digest), []byte(password))
	if err != nil {
		return false
	}
	return true
}
