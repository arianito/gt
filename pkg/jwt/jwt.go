package jwt

import (
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
	"log"
	"time"
)

type AuthClaims struct {
	UserId     string
	TokenId    string
	Gateway    string
	ClientName string
	UserAgent string
	ExpiresAt  time.Time
}


type Authenticator interface {
	Create(claims *AuthClaims) (string, error)
	Validate(token string) bool
	Parse(token string) (*AuthClaims, error)
	Renew(token string, expiresAt time.Time) (string, error)
}


func NewClient(client *Client) Authenticator {
	a := new(Client)
	pubKey, err := base64.StdEncoding.DecodeString(client.PublicKey)
	if err != nil {
		log.Fatal("Failed to parse public key: ", err)
		return nil
	}
	pukKey, err := base64.StdEncoding.DecodeString(client.PrivateKey)
	if err != nil {
		log.Fatal("Failed to parse private key: ", err)
		return nil
	}
	a.publicKey, err = crypto.ParseRSAPublicKeyFromPEM(pubKey)
	if err != nil {
		log.Fatal("Failed to load public key: ", err)
		return nil
	}
	a.privateKey, err = crypto.ParseRSAPrivateKeyFromPEM(pukKey)
	if err != nil {
		log.Fatal("Failed to load private key: ", err)
		return nil
	}
	return a
}

type Client struct {
	PublicKey string
	PrivateKey string
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
}
func (a *Client) Create(c *AuthClaims) (string, error) {
	claims := jws.Claims{}
	claims.Set("userId", c.UserId)
	claims.Set("tokenId", c.TokenId)
	claims.Set("gateway", c.Gateway)
	claims.Set("userAgent", c.UserAgent)
	claims.Set("clientName", c.ClientName)
	claims.SetExpiration(c.ExpiresAt)
	claims.SetIssuedAt(time.Now())
	j := jws.NewJWT(claims, crypto.SigningMethodRS256)
	output, err := j.Serialize(a.privateKey)
	if err != nil {
		return "", err
	}
	return string(output), nil
}
func (a *Client) Validate(token string) bool {
	j, err := jws.ParseJWT([]byte(token))
	if err != nil {
		return false
	}
	if err = j.Validate(a.publicKey, crypto.SigningMethodRS256); err != nil {
		return false
	}
	return true
}
func (a *Client) Parse(token string) (*AuthClaims, error) {
	claims := new(AuthClaims)
	j, err := jws.ParseJWT([]byte(token))
	if err != nil {
		return nil, errors.New("1001 failed parse token, "+ err.Error())
	}
	err = j.Validate(a.publicKey, crypto.SigningMethodRS256)
	if err != nil {
		return nil, errors.New("1002 failed validate token, "+ err.Error())
	}
	jwtClaims := j.Claims()
	claims.UserId = jwtClaims.Get("userId").(string)
	claims.TokenId = jwtClaims.Get("tokenId").(string)
	claims.Gateway = jwtClaims.Get("gateway").(string)
	claims.UserAgent = jwtClaims.Get("userAgent").(string)
	claims.ClientName = jwtClaims.Get("clientName").(string)
	claims.ExpiresAt, _ = jwtClaims.Expiration()
	return claims, nil
}
func (a *Client) Renew(token string, expiresAt time.Time) (string, error) {
	claims, err := a.Parse(token)
	if err != nil {
		return "", err
	}
	claims.ExpiresAt = expiresAt
	return a.Create(claims)
}
