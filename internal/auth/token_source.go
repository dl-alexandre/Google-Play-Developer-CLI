package auth

import (
	"encoding/json"
	"time"

	"golang.org/x/oauth2"

	gpdErrors "github.com/dl-alexandre/gpd/internal/errors"
)

type PersistedTokenSource struct {
	base       oauth2.TokenSource
	storage    SecureStorage
	storageKey string
	metadata   *TokenMetadata
}

func (s *PersistedTokenSource) Token() (*oauth2.Token, error) {
	token, err := s.base.Token()
	if err != nil {
		if apiErr := gpdErrors.ClassifyAuthError(err); apiErr != nil {
			return nil, apiErr
		}
		return nil, err
	}

	if token == nil || s.storage == nil {
		return token, nil
	}

	if token.RefreshToken == "" {
		if existing, err := s.storage.Retrieve(s.storageKey); err == nil && len(existing) > 0 {
			var storedToken StoredToken
			if err := json.Unmarshal(existing, &storedToken); err == nil && storedToken.RefreshToken != "" {
				token.RefreshToken = storedToken.RefreshToken
			}
		}
	}

	stored := StoredToken{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		Expiry:       token.Expiry.Format(time.RFC3339),
	}
	if s.metadata != nil {
		stored.Origin = s.metadata.Origin
		stored.Email = s.metadata.Email
		stored.Scopes = s.metadata.Scopes
	}
	data, err := json.Marshal(stored)
	if err == nil {
		_ = s.storage.Store(s.storageKey, data)
	}
	if s.metadata != nil {
		s.metadata.TokenExpiry = token.Expiry.Format(time.RFC3339)
		s.metadata.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
		_ = writeTokenMetadata(s.storageKey, s.metadata)
	}
	return token, nil
}
