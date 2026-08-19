package deep

import "math/rand/v2"

// A WeightInitializer returns a (random) weight
type WeightInitializer func() float64

// NewUniform returns a uniform weight generator
func NewUniform(stdDev, mean float64, random *rand.Rand) WeightInitializer {
	return func() float64 { return Uniform(stdDev, mean, random) }
}

// Uniform samples a value from u(mean-stdDev/2,mean+stdDev/2)
func Uniform(stdDev, mean float64, random *rand.Rand) float64 {
    var v float64
    if random != nil {
        v = random.Float64()
    } else {
        v = rand.Float64()
    }
	return (v-0.5)*stdDev + mean

}

// NewNormal returns a normal weight generator
func NewNormal(stdDev, mean float64, random *rand.Rand) WeightInitializer {
	return func() float64 { return Normal(stdDev, mean, random) }
}

// Normal samples a value from N(μ, σ)
func Normal(stdDev, mean float64, random *rand.Rand) float64 {
    var v float64
    if random != nil {
        v = random.NormFloat64()
    } else {
        v = rand.NormFloat64()
    }
	return v*stdDev + mean
}
