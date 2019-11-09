package hash

import (
	"golang.org/x/crypto/bcrypt"
	"math/rand"
	"time"
)

type Hash interface {
	Hash(psw string) string
	CheckHash(hashed, psw string) bool
	GenerateCode(size int) string
	GenerateFixed() string
	RandomString(n int) string
}

func NewClient() Hash {
	return new(Client)
}

type Client struct {
}

func (h *Client) GenerateCode(size int) string {
	out := make([]byte, size)
	nano := time.Now().UnixNano()
	src := rand.NewSource(nano)
	rng := rand.New(src)
	for i := 0; i < size; i++ {
		out[i] = h.mapChar(rng.Uint64() + uint64(nano))
	}
	return string(out)
}

func (h *Client) GenerateFixed() string {
	return h.GenerateCode(4)
}

func (*Client) Hash(password string) string {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	return string(hashedPassword)
}
func (*Client) CheckHash(hashed string, password string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password)); err != nil {
		return false
	}
	return true
}

func (*Client) RandomString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func (*Client) mapChar(a uint64) byte {
	return byte(a%10) + 48
}