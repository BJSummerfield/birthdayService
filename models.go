package main

import "github.com/google/uuid"

type Birthday struct {
	ID       uuid.UUID `json:"id" bson:"id"`
	Birthday string    `json:"birthday" bson:"birthday"`
}
