package api

import (
	"context"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

func InitAuth(ctx context.Context, mode string, key string, alg string, jwksURL string) (AuthState, error) {
	authState := AuthState{}
	authState.ValidationType = mode
	authState.Alg = jwa.KeyAlgorithmFrom(alg)

	if mode == "LOCAL" {
		switch algPrefix := alg[:2]; algPrefix {
		case "RS", "ES":

			decodedKey, err := base64.StdEncoding.DecodeString(key)
			if err != nil {
				return authState, fmt.Errorf("failed to decode base64 public key: %w", err)
			}

			pubKey, err := x509.ParsePKIXPublicKey(decodedKey)
			if err != nil {
				return authState, fmt.Errorf("failed to parse DER encoded public key: %w", err)
			}

			switch algPrefix {
			case "RS":
				if _, ok := pubKey.(*rsa.PublicKey); !ok {
					return authState, errors.New("public key is not of type RSA for RS* algorithm")
				}
			case "ES":
				if _, ok := pubKey.(*ecdsa.PublicKey); !ok {
					return authState, errors.New("public key is not of type ECDSA for ES* algorithm")
				}
			}
			authState.Key = pubKey
		case "HS":
			authState.Key = []byte(key)
		default:
			return authState, fmt.Errorf("unsupported algorithm family: %s", alg)
		}
		return authState, nil
	}

	// JWKS mode
	c, err := InitCache(jwksURL, ctx)
	if err != nil {
		return authState, err
	}

	authState.Cache = c
	authState.JwksURL = jwksURL

	return authState, nil
}

func InitCache(jwksURL string, ctx context.Context) (*jwk.Cache, error) {

	cache := jwk.NewCache(ctx)
	cache.Register(jwksURL, jwk.WithMinRefreshInterval(15*time.Minute))

	_, err := cache.Get(ctx, jwksURL)

	if err != nil {
		return nil, err
	}

	return cache, nil
}

func (a AppState) RegisterAuthMiddleware(ctx huma.Context, next func(huma.Context)) {

	authHeader := ctx.Header("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		huma.WriteErr(*a.HumaInstance, ctx, http.StatusUnauthorized, "invalid or missing token", fmt.Errorf("missing bearer token"))

		return
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

	// var t jwt.Token
	var err error

	if a.Auth.ValidationType == "LOCAL" {
		_, err = jwt.Parse([]byte(tokenStr), jwt.WithKey(a.Auth.Alg, a.Auth.Key))
		if err != nil {
			huma.WriteErr(*a.HumaInstance, ctx, http.StatusUnauthorized, "invalid or missing token", fmt.Errorf("missing bearer token"))
			return
		}
	} else {
		keyset, err := a.Auth.Cache.Get(ctx.Context(), a.Auth.JwksURL)
		if err != nil {
			huma.WriteErr(*a.HumaInstance, ctx, http.StatusUnauthorized, "invalid or missing token", fmt.Errorf("missing bearer token"))
			return
		}
		_, err = jwt.Parse([]byte(tokenStr), jwt.WithKeySet(keyset))
		if err != nil {
			huma.WriteErr(*a.HumaInstance, ctx, http.StatusUnauthorized, "invalid or missing token", fmt.Errorf("missing bearer token"))
			return
		}
	}

	next(ctx)
}
