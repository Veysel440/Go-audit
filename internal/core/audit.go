package core

import "time"

type Audit struct {
	ID           string         `json:"id"            bson:"_id"`
	ActorID      string         `json:"actorId"       bson:"actorId"`
	ActorType    string         `json:"actorType"     bson:"actorType"`
	Action       string         `json:"action"        bson:"action"`
	ResourceID   string         `json:"resourceId"    bson:"resourceId"`
	ResourceType string         `json:"resourceType"  bson:"resourceType"`
	IP           string         `json:"ip,omitempty"  bson:"ip,omitempty"`
	UA           string         `json:"ua,omitempty"  bson:"ua,omitempty"`
	Metadata     map[string]any `json:"metadata,omitempty" bson:"metadata,omitempty"`
	CreatedAt    time.Time      `json:"createdAt"     bson:"createdAt"`

	ChainPrev string `json:"chainPrev,omitempty" bson:"chainPrev,omitempty"`
	ChainHash string `json:"chainHash,omitempty" bson:"chainHash,omitempty"`

	IdemHash string `json:"-" bson:"idemHash,omitempty"`
}

type CreateAudit struct {
	ActorID, ActorType, Action, ResourceID, ResourceType string
	IP, UA                                               string
	Metadata                                             map[string]any
	IdemHash                                             string
}
