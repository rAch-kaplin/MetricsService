package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"

	log "github.com/rAch-kaplin/mipt-golang-course/MetricsService/pkg/logger"
)

func GetHash(key, data []byte) ([]byte, error) {
	h := hmac.New(sha256.New, key)

	if _, err := h.Write(data); err != nil {
		log.Error().Err(err).Msg("invalid hashing sha256")
		return nil, fmt.Errorf("invalid hashing sha256 %v", err)
	}

	return h.Sum(nil), nil
}

func CheckHash(key, data, expdHash []byte) bool {
	hash, err := GetHash(key, data)
	if err != nil {
		log.Error().Err(err).Msg("cat't get hash")
		return false
	}

	return hmac.Equal(hash, expdHash)
}
