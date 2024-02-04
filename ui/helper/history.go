package helper

type HistoryManager struct {
	history  []string
	position int
}

func NewHistoryManager() *HistoryManager {
	return &HistoryManager{
		history:  []string{""},
		position: 0,
	}
}

func (h *HistoryManager) Add(text string) {
	h.history = append(h.history, text)
	h.position = len(h.history)
}

func (h *HistoryManager) Current() string {
	return h.history[h.position]
}

func (h *HistoryManager) Length() int {
	return len(h.history)
}

func (h *HistoryManager) Previous() string {
	if len(h.history) == 0 {
		return ""
	}
	if h.position == 0 {
		h.position = len(h.history)
	}
	h.position--
	return h.history[h.position]
}

func (h *HistoryManager) Next() string {
	if len(h.history) == 0 {
		return ""
	}
	h.position++
	if h.position >= len(h.history) {
		h.position = 0
	}
	return h.history[h.position]
}
