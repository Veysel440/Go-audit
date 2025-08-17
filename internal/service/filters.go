package service

import "time"

type ListFilter struct {
	ActorID, ActorType, Action, ResourceID, ResourceType string
	Since, Until                                         *time.Time
	Limit, Offset                                        int64
}
