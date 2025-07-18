package store

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tesseral-labs/tesseral/internal/scim/authn"
	"github.com/tesseral-labs/tesseral/internal/scim/store/queries"
	"github.com/tesseral-labs/tesseral/internal/uuidv7"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type logAuditEventParams struct {
	EventName      string
	EventDetails   proto.Message
	ResourceType   queries.AuditLogEventResourceType
	ResourceID     *uuid.UUID
	OrganizationID *uuid.UUID
}

func (s *Store) logAuditEvent(ctx context.Context, q *queries.Queries, req logAuditEventParams) (queries.AuditLogEvent, error) {
	// Generate the UUIDv7 based on the event time.
	eventTime := time.Now()
	eventID := uuidv7.NewWithTime(eventTime)

	eventDetailsBytes, err := protojson.Marshal(req.EventDetails)
	if err != nil {
		return queries.AuditLogEvent{}, fmt.Errorf("failed to marshal event details: %w", err)
	}

	qEventParams := queries.CreateAuditLogEventParams{
		ID:                eventID,
		ProjectID:         authn.ProjectID(ctx),
		OrganizationID:    req.OrganizationID,
		ResourceType:      refOrNil(req.ResourceType),
		ResourceID:        req.ResourceID,
		EventName:         req.EventName,
		EventTime:         &eventTime,
		EventDetails:      eventDetailsBytes,
		ActorScimApiKeyID: refOrNil(authn.SCIMAPIKeyID(ctx)),
	}

	qEvent, err := q.CreateAuditLogEvent(ctx, qEventParams)
	if err != nil {
		return queries.AuditLogEvent{}, err
	}

	return qEvent, nil
}
