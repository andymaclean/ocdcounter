package main

import (
	googleUUID "github.com/google/uuid"
)

type UUID struct {
	uuid googleUUID.UUID
}

func (u UUID) String() string {
	return u.uuid.String()
}

func (u *UUID) MarshalJSON() ([]byte, error) {
	return u.uuid.MarshalBinary()
}

func (u *UUID) UnmarshalJSON(b []byte) error {
	var gerr error
	u.uuid, gerr = googleUUID.ParseBytes(b)
	return gerr
}

func MakeUUID() UUID {
	return UUID{
		uuid: googleUUID.Must(googleUUID.NewRandom()),
	}
}

func ToUUID(uuid string) (UUID, error) {
	u, err := googleUUID.Parse(uuid)

	if err != nil {
		return UUID{}, err
	}

	return UUID{
		uuid: u,
	}, nil
}
