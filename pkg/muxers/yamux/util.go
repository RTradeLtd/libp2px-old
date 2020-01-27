package yamux

// asyncNotify is used to signal a waiting goroutine
func asyncNotify(ch chan struct{}) {
	select {
	case ch <- struct{}{}:
	default:
	}
}

// min computes the minimum of a set of values
func min(values ...uint32) uint32 {
	m := values[0]
	for _, v := range values[1:] {
		if v < m {
			m = v
		}
	}
	return m
}
