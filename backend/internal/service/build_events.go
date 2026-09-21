package service

import (
	"sync"

	"devops-platform/internal/model"
)

// BuildEvent is pushed to browser clients via SSE when a build job changes status.
type BuildEvent struct {
	ProjectID   uint   `json:"project_id"`
	BuildID     uint   `json:"build_id"`
	BuildNumber uint   `json:"build_number"`
	Status      string `json:"status"`
	Branch      string `json:"branch"`
	CommitSHA   string `json:"commit_sha"`
}

type buildEventHub struct {
	mu   sync.RWMutex
	subs map[chan BuildEvent]struct{}
}

// BuildEvents fans out build status changes to subscribed SSE clients (in-process).
var BuildEvents = &buildEventHub{subs: make(map[chan BuildEvent]struct{})}

func (h *buildEventHub) Subscribe() chan BuildEvent {
	ch := make(chan BuildEvent, 16)
	h.mu.Lock()
	h.subs[ch] = struct{}{}
	h.mu.Unlock()
	return ch
}

func (h *buildEventHub) Unsubscribe(ch chan BuildEvent) {
	h.mu.Lock()
	if _, ok := h.subs[ch]; ok {
		delete(h.subs, ch)
		close(ch)
	}
	h.mu.Unlock()
}

func (h *buildEventHub) Publish(ev BuildEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.subs {
		select {
		case ch <- ev:
		default:
			// slow client: drop rather than block the build goroutine
		}
	}
}

func PublishBuildEvent(job model.BuildJob) {
	BuildEvents.Publish(BuildEvent{
		ProjectID:   job.ProjectID,
		BuildID:     job.ID,
		BuildNumber: job.BuildNumber,
		Status:      job.Status,
		Branch:      job.Branch,
		CommitSHA:   job.CommitSHA,
	})
}
