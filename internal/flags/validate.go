package flags

import (
	"fmt"

	"github.com/open-crypto-broker/crypto-broker-cli-go/internal/constant"
)

// ValidateFlagLoop validates loop flag value.
func ValidateFlagLoop(val int) error {
	if val == constant.NoLoopFlagValue {
		return nil
	}

	if val < constant.MinLoopFlagValue || val > constant.MaxLoopFlagValue {
		return fmt.Errorf("'loop' flag value must be between %d and %d", constant.MinLoopFlagValue, constant.MaxLoopFlagValue)
	}

	return nil
}

// ValidateFlagEncoding validates encoding flag value.
func ValidateFlagEncoding(val string) error {
	if val != constant.EncodingPEM && val != constant.EncodingB64 {
		return fmt.Errorf("'encoding' flag value must be %s or %s", constant.EncodingPEM, constant.EncodingB64)
	}

	return nil
}
