package server

import (
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/handler"
	"github.com/labstack/echo/v4"
)

// SetupRoutes sets up routes for the server.
func (s *Server) SetupRoutes() {
	s.e.GET("/", func(c echo.Context) error {
		return c.String(200, "Doordarshan Kendra")
	})
	s.e.GET("/health", handler.HealthProbe)

	s.e.POST("/v1/createMeeting", s.meetingV1handler.CreateMeeting)
	s.e.POST("/v1/joinMeeting", s.meetingV1handler.JoinMeeting)
	s.e.POST("/v1/getProducersOfMeeting", s.meetingV1handler.GetProducersOfMeeting)
	s.e.POST("/v1/connectProducerTransport", s.meetingV1handler.ConnectProducerTransport)
	s.e.POST("/v1/createProducer", s.meetingV1handler.CreateProducer)
	s.e.POST("/v1/connectConsumerTransport", s.meetingV1handler.ConnectConsumerTransport)
	s.e.POST("/v1/recreateProducerTransport", s.meetingV1handler.RecreateProducerTransport)
	s.e.POST("/v1/recreateConsumerTransport", s.meetingV1handler.RecreateConsumerTransport)
	//s.e.POST("/v1/recreateTransports", s.meetingV1handler.RecreateTransports)
	s.e.POST("/v1/restartProducerIce", s.meetingV1handler.RestartProducerIce)
	s.e.POST("/v1/restartConsumerIce", s.meetingV1handler.RestartConsumerIce)
	s.e.POST("/v1/restartIce", s.meetingV1handler.RestartIce)
	s.e.POST("/v1/createConsumer", s.meetingV1handler.CreateConsumer)
	s.e.POST("/v1/resumeConsumer", s.meetingV1handler.ResumeConsumer)
	s.e.POST("/v1/resumeProducer", s.meetingV1handler.ResumeProducer)
	s.e.POST("/v1/pauseProducer", s.meetingV1handler.PauseProducer)
	s.e.POST("/v1/pauseConsumer", s.meetingV1handler.PauseConsumer)
	s.e.POST("/v1/closeProducer", s.meetingV1handler.CloseProducer)
	// s.e.POST("/v1/closeAllProducers", s.meetingV1handler.CloseAllProducers)
	/*s.e.POST("/v1/closeAllConsumersForProducer", s.meetingV1handler.CloseAllConsumersForProducer)*/
	s.e.POST("/v1/closeConsumer", s.meetingV1handler.CloseConsumer)
	s.e.POST("/v1/leaveMeeting", s.meetingV1handler.LeaveMeeting)
	s.e.POST("/v1/endMeeting", s.meetingV1handler.EndMeeting)
	s.e.POST("/v1/getRTPCapabilities", s.meetingV1handler.GetRTPCapabilities)
	s.e.POST("/v1/preMeetingDetails", s.meetingV1handler.PreMeetingDetails)
	s.e.POST("/v1/activeContainer", s.meetingV1handler.GetActiveContainerOfMeeting)

}
