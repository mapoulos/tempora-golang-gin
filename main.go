package main

import (
	"github.com/gin-gonic/gin"
)

type Meditation struct {
	ID     string `json:"_id"`
	UserId string `json:"_userId"`
	URL    string `json:"audioUrl"`
	Name   string `json:"name"`
}

type CreateMeditationInput struct {
	URL  string `json:"audioUrl" binding:"required"`
	Name string `json:"name" binding:"required"`
}

type Env struct {
	store MeditationStore
}

type Abser interface {
	Abs() float64
}

type MyFloat struct {
	num float64
}

func (m MyFloat) Abs() float64 {
	if m.num > 0 {
		return float64(m.num)
	}
	return float64(-m.num)
}

func main() {
	r := gin.Default()

	env := Env{}
	env.store = NewDynamoMeditationStore("tempora-local", true)

	r.GET("/meditations", env.ListMeditationsForUserHandler)

	r.GET("/meditations/:id", env.GetMeditationForUser)

	r.PATCH("/meditations/:id", env.UpdateMeditationForUserHandler)

	r.PUT("/meditations/:id", env.UpdateMeditationForUserHandler)

	r.DELETE("/meditations/:id", env.DeleteMeditationForUserHandler)

	r.POST("/meditations", env.CreateMeditationForUserHandler)

	r.Run()
}
