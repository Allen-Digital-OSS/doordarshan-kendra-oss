package handler

/*import (
	"context"
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/common"
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/data/entity"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var (
	meetingCreateRequestJSON = `{
									"name": "Meeting 1",
									"description": "Meeting 1 description",
									"start_time": "2021-01-01T00:00:00Z",
									"duration_in_minutes": 60,
									"max_participants": 10,
									"max_interactive_participants": 5,
									"meeting_type": "standard"
								}`
)

type MockMeetingRepository struct {
	mock.Mock
}

func (m *MockMeetingRepository) Create(ctx context.Context,
	meeting *entity.MeetingEntity) error {
	args := m.Called(ctx, meeting)
	return args.Error(0)
}

func (m *MockMeetingRepository) Update(ctx context.Context,
	meeting *entity.MeetingEntity) error {
	args := m.Called(ctx, meeting)
	return args.Error(0)
}

func (m *MockMeetingRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMeetingRepository) FindByID(ctx context.Context, id string) (*entity.MeetingEntity, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entity.MeetingEntity), args.Error(1)
}

func (m *MockMeetingRepository) List(ctx context.Context, offset uint, limit int) ([]*entity.MeetingEntity, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]*entity.MeetingEntity), args.Error(1)
}

func Test_MeetingHandler_CreateMeeting(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/meetings",
		strings.NewReader(meetingCreateRequestJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockRepository := &MockMeetingRepository{}
	mockRepository.
		On("Create", c.Request().Context(), mock.Anything).
		Return(nil)

	h := NewMeetingHandler(&common.AppConfig{}, mockRepository)

	// Assertions
	if assert.NoError(t, h.CreateMeeting(c)) {
		assert.Equal(t, http.StatusCreated, rec.Code)
	}
}

func Test_MeetingHandler_DeleteMeeting(t *testing.T) {

}

func Test_MeetingHandler_GetMeeting(t *testing.T) {

}

func Test_MeetingHandler_ListMeetings(t *testing.T) {

}

func Test_MeetingHandler_UpdateMeeting(t *testing.T) {

}
*/
