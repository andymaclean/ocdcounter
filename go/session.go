package main

import "fmt"

// data type for the session.  Contains session specific parameters
type APISession struct {
	userId        UUID
	userIdString  *string
	groupId       UUID
	groupIdString *string
	userEmail     *string
}

func Create_APISession(dbo DataOperator, req Request) (APISession, error) {
	if req.RequestContext.Authorizer == nil {
		return APISession{}, fmt.Errorf("username is not in JWT claims")
	}

	email, hasemail := req.RequestContext.Authorizer.JWT.Claims["cognito:username"]

	if !hasemail {
		return APISession{}, fmt.Errorf("username is not in JWT claims")
	}

	uuid, uerror := dbo.LookupUserUUID(&email)

	if uerror != nil {
		return APISession{}, uerror
	}

	group, hasgrp := req.PathParameters["group"]

	if hasgrp {
		if groupId, gerr := ToUUID(group); gerr != nil {
			return APISession{}, gerr
		} else {
			return APISession{
				userId:    uuid,
				groupId:   groupId,
				userEmail: &email,
			}, nil
		}
	}

	return APISession{
		userId:    uuid,
		userEmail: &email,
	}, nil
}

func (s APISession) GetUserId() *UUID {
	return &s.userId
}

func (s *APISession) GetUserIdString() *string {
	if s.userIdString == nil {
		st := s.userId.String()
		s.userIdString = &st
	}
	return s.userIdString
}

func (s APISession) GetGroupId() *UUID {
	return &s.groupId
}

func (s *APISession) GetGroupIdString() *string {
	if s.groupIdString == nil {
		st := s.groupId.String()
		s.groupIdString = &st
	}
	return s.groupIdString
}
