package main

import "go.mongodb.org/mongo-driver/bson/primitive"

type Birthday struct {
    ID       primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
    Birthday string             `json:"birthday" bson:"birthday"`
}
