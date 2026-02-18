package myerrors

import "errors"

var (
	ParticipantNotFound    = errors.New("participant not found")
	RouterNotFound         = errors.New("router not found")
	MoreThanOneRouterFound = errors.New("more than one router found")
	MeetingNotFound        = errors.New("meeting not found")
	InValidCapacity        = errors.New("invalid capacity")
	NoFreeContainer        = errors.New("no free container")
	UnhandledCase          = errors.New("unhandled case")
)
