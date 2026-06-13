package playback

import "time"

type NotifierGroup struct {
	notifiers []Notifier
}

func NewNotifierGroup(notifiers ...Notifier) *NotifierGroup {
	group := &NotifierGroup{}
	for _, notifier := range notifiers {
		if notifier != nil {
			group.notifiers = append(group.notifiers, notifier)
		}
	}
	if len(group.notifiers) == 0 {
		return nil
	}
	return group
}

func (g *NotifierGroup) Update(state State) {
	if g == nil {
		return
	}
	for _, notifier := range g.notifiers {
		notifier.Update(state)
	}
}

func (g *NotifierGroup) Seeked(position time.Duration) {
	if g == nil {
		return
	}
	for _, notifier := range g.notifiers {
		notifier.Seeked(position)
	}
}
