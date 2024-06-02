package generator

import (
	"cmd/internal/experiment"
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"
	"unsafe"

	"github.com/google/uuid"
)

type RandomGenerator struct {
	generator *rand.Rand
}

func New() RandomGenerator {
	generator := rand.New(rand.NewSource(time.Now().UnixNano()))
	return RandomGenerator{generator: generator}
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6
	letterIdxMask = 1<<letterIdxBits - 1
	letterIdxMax  = 63 / letterIdxBits
)

func (g *RandomGenerator) GenerateRandomString(n uint64) string {
	b := make([]byte, n)
	for i, cache, remain := int(n)-1, g.generator.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = g.generator.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return *(*string)(unsafe.Pointer(&b))
}

func (g *RandomGenerator) GenerateRandomInt() uint64 {
	return uint64(math.Abs(float64(g.generator.Int63())))
}

func (g *RandomGenerator) GenerateUUID() uuid.UUID {
	return uuid.New()
}

func (g *RandomGenerator) GenerateTime(after time.Time) time.Time {
	now := time.Now()
	diff := now.Sub(after).Seconds()
	shift := g.GenerateRandomInt() % uint64(diff)
	return after.Add(time.Second * time.Duration(shift))
}

func (g *RandomGenerator) RandomByType(t experiment.Type) any {
	switch t {
	case experiment.Integer:
		return g.GenerateRandomInt()
	case experiment.String:
		return g.GenerateRandomString(8)
	case experiment.UUID:
		return g.GenerateUUID()
	case experiment.DateTime:
		return g.GenerateTime(time.Now().AddDate(0, 0, -1))
	case experiment.ByteArray:
		return []byte(g.GenerateRandomString(8))
	}
	if strings.HasPrefix(string(t), string(experiment.Varchar)) {
		return g.GenerateRandomString(8)
	}
	if strings.HasPrefix(string(t), string(experiment.Timestamp)) {
		return g.GenerateTime(time.Now().AddDate(0, 0, -1))
	}
	fmt.Println("WARN: unexpected type in random", t)
	return nil
}
