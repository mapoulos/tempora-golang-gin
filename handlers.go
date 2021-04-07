package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

func (e *Env) ListMeditationsForUserHandler(c *gin.Context) {
	userId := c.GetHeader("User-Id")

	if userId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No User-Id header present"})
		return
	}

	meditations, _ := e.store.ListMeditations(userId)
	if meditations == nil {
		c.JSON(http.StatusOK, [0]Meditation{})
		return
	}
	c.JSON(http.StatusOK, meditations)
}

func (e *Env) GetMeditationForUser(c *gin.Context) {
	id := c.Param("id")
	userId := c.GetHeader("User-Id")

	if userId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No User-Id header present"})
		return
	}

	m, err := e.store.GetMeditation(userId, id)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, m)
}

func (e *Env) DeleteMeditationForUserHandler(c *gin.Context) {
	id := c.Param("id")
	userId := c.Request.Header["User-Id"][0]

	if err := e.store.DeleteMeditation(userId, id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func (e *Env) CreateMeditationForUserHandler(c *gin.Context) {
	var input CreateMeditationInput

	userId := c.Request.Header["User-Id"][0]

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	u4 := uuid.NewV4()

	newMeditation := Meditation{
		ID:     u4.String(),
		URL:    input.URL,
		Name:   input.Name,
		UserId: userId,
	}

	e.store.SaveMeditation(newMeditation)

	c.JSON(http.StatusCreated, newMeditation)
}

func (e *Env) UpdateMeditationForUserHandler(c *gin.Context) {
	id := c.Param("id")
	userId := c.GetHeader("User-Id")

	if userId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No User-Id header present"})
		return
	}

	var input CreateMeditationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedMeditation := Meditation{
		ID:     id,
		URL:    input.URL,
		Name:   input.Name,
		UserId: userId,
	}

	err := e.store.UpdateMeditation(updatedMeditation)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
