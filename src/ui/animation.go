package ui

func MoveTowards(current, target, amount float32) float32 {
	if current < target {
		return min(current+amount, target)
	}
	return max(current-amount, target)
}

func Smoothstep(value float32) float32 {
	value = max(float32(0), min(value, float32(1)))
	return value * value * (3 - 2*value)
}
