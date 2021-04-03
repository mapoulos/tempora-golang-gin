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
	store *MemoryMeditationStore
}

func main() {
	r := gin.Default()

	env := Env{}
	env.store = NewMemoryMeditationStore()

	r.GET("/meditations", env.ListMeditationsForUserHandler)

	r.GET("/meditations/:id", env.GetMeditationForUser)

	r.DELETE("/meditations/:id", env.DeleteMeditationForUserHandler)

	r.POST("/meditations", env.CreateMeditationForUserHandler)

	r.Run()
}
