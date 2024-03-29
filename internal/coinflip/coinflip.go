package coinflip

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func Flip() bool {
	/* #nosec G404 -- coinflip not used for cryptographic functions */
	if flipint := rand.Intn(2); flipint == 0 {
		return true
	}
	return false
}
