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
	h.position = len(h.history) - 1
}

func (h *HistoryManager) Current() string {
	return h.history[h.position]
}

func (h *HistoryManager) Length() int {
	return len(h.history)
}

func (h *HistoryManager) Previous() string {
	h.position -= 1
	if h.position <= 0 {
		h.position = 0
		return ""
	}
	return h.history[h.position+1]
}

func (h *HistoryManager) Next() string {
	h.position += 1
	if h.position >= len(h.history) {
		h.position = len(h.history) - 1
		return ""
	}
	return h.history[h.position]
}
