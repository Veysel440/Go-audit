package service

import "time"

type ListFilter struct {
	ActorID, ActorType, Action, ResourceID, ResourceType string
	Since, Until                                         *time.Time
	Limit, Offset                                        int64
}

// Normalize applies sane defaults and bounds
func (f *ListFilter) Normalize() {
	if f.Limit <= 0 || f.Limit > 1000 {
		f.Limit = 100
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
}
