package test_juno

//types to implementation container/heap
//container sorted by expiration

//ExpiringSign key and time of expiration
type ExpiringSign struct {
	expiration int64
	key        string
}

type ExpSignHeap []ExpiringSign

func (h ExpSignHeap) Len() int {
	return len(h)
}

func (h ExpSignHeap) Less(i, j int) bool {
	return h[i].expiration < h[j].expiration
}

func (h ExpSignHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *ExpSignHeap) Push(x interface{}) {
	*h = append(*h, x.(ExpiringSign))
}

func (h *ExpSignHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
