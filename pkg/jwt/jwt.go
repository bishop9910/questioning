// Package jwt 手写 HS256 JSON Web Token 的签发与校验，不引第三方库。
package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// ErrInvalid 表示 token 格式不对、签名不匹配或已过期。
var ErrInvalid = errors.New("token 无效或已过期")

var headerPart = base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))

// Claims 是载荷，只放用户 ID 和过期时间。
type Claims struct {
	Subject   int64 `json:"sub"`
	IssuedAt  int64 `json:"iat"`
	ExpiresAt int64 `json:"exp"`
}

func sign(secret []byte, payload string) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// Sign 为 subject 签发有效期 ttl 的 token。
func Sign(secret []byte, subject int64, ttl time.Duration) (string, error) {
	now := time.Now()
	body, err := json.Marshal(Claims{
		Subject:   subject,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(ttl).Unix(),
	})
	if err != nil {
		return "", err
	}
	payload := headerPart + "." + base64.RawURLEncoding.EncodeToString(body)
	return payload + "." + sign(secret, payload), nil
}

// Parse 校验签名与有效期，返回用户 ID。
func Parse(secret []byte, token string) (int64, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return 0, ErrInvalid
	}
	payload := parts[0] + "." + parts[1]
	if !hmac.Equal([]byte(sign(secret, payload)), []byte(parts[2])) {
		return 0, ErrInvalid
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return 0, ErrInvalid
	}
	var c Claims
	if err := json.Unmarshal(raw, &c); err != nil {
		return 0, ErrInvalid
	}
	if c.ExpiresAt <= time.Now().Unix() || c.Subject <= 0 {
		return 0, ErrInvalid
	}
	return c.Subject, nil
}
