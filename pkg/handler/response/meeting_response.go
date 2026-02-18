package response

/*import (
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/data/entity"
)

type MeetingResponse struct {
	ID                         string `json:"id"`
	Name                       string `json:"name"`
	Description                string `json:"description,omitempty"`
	StartTime                  string `json:"start_time,omitempty"`
	DurationInMinutes          int    `json:"duration_in_minutes,omitempty"`
	MeetingType                string `json:"meeting_type,omitempty"`
	MaxParticipants            int    `json:"max_participants,omitempty"`
	MaxInteractiveParticipants int    `json:"max_interactive_participants,omitempty"`
	CreatedAt                  string `json:"created_at,omitempty"`
	UpdatedAt                  string `json:"updated_at,omitempty"`
	CreatedBy                  string `json:"created_by,omitempty"`
	UpdatedBy                  string `json:"updated_by,omitempty"`
	TenantID                   string `json:"tenant_id,omitempty"`
}

func (mr *MeetingResponse) FromEntity(e *entity.MeetingEntity) {
	mr.ID = e.ID
	mr.Name = e.Name
	mr.Description = e.Description
	mr.StartTime = e.StartTime.UTC().String()
	mr.DurationInMinutes = e.DurationInMinutes
	mr.MeetingType = e.MeetingType
	mr.MaxParticipants = e.MaxParticipants
	mr.MaxInteractiveParticipants = e.MaxInteractiveParticipants
	mr.CreatedAt = e.BaseEntity.CreatedAt.UTC().String()
	mr.UpdatedAt = e.BaseEntity.UpdatedAt.UTC().String()
	mr.CreatedBy = e.BaseEntity.CreatedBy
	mr.UpdatedBy = e.BaseEntity.UpdatedBy
	mr.TenantID = e.BaseEntity.TenantID
}*/

// GetActiveContainerOfMeetingResponse represents the response containing active container information
type GetActiveContainerOfMeetingResponse struct {
	MeetingId   string `json:"meeting_id" validate:"required" example:"meeting-123"`     // Meeting identifier
	ContainerId string `json:"container_id" validate:"required" example:"container-abc"` // Container identifier where the meeting is hosted
}
