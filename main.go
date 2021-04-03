package main

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

type Meditation struct {
	ID     string `json:"_id"`
	UserId string `json:"_userId"`
	URL    string `json:"audioUrl"`
	Name   string `json:"name"`
}

type CreateMeditationInput struct {
	URL    string `json:"audioUrl" binding:"required"`
	UserId string `json:"_user_id"`
	Name   string `json:"name" binding:"required"`
}

var meditations []Meditation

func SaveMeditation(meditation Meditation) {
	slice := append(meditations, meditation)
	meditations = slice
}

func ListMeditations() []Meditation {
	return meditations
}

func GetMeditation(id string) (Meditation, error) {
	for _, m := range meditations {
		if m.ID == id {
			return m, nil
		}
	}
	emptyMeditation := Meditation{
		Name: "",
		URL:  "",
		ID:   "",
	}
	return emptyMeditation, errors.New("No meditation with id " + id + " was found")
}

func DeleteMeditation(id string) error {
	idxToDelete := -1
	for i, m := range meditations {
		if m.ID == id {
			idxToDelete = i
		}
	}

	if idxToDelete > -1 {
		finalIdx := len(meditations) - 1
		meditations[idxToDelete], meditations[finalIdx] = meditations[finalIdx], meditations[idxToDelete]
		meditations = meditations[:finalIdx]
		return nil
	}

	return errors.New("No meditation with id " + id + " was found")
}

func main() {
	r := gin.Default()

	meditations = []Meditation{}

	r.GET("/meditations", func(c *gin.Context) {
		c.JSON(http.StatusOK, meditations)
	})

	r.GET("/meditations/:id", func(c *gin.Context) {
		id := c.Param("id")
		m, err := GetMeditation(id)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, m)
	})

	r.DELETE("/meditations/:id", func(c *gin.Context) {
		id := c.Param("id")

		if err := DeleteMeditation(id); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{})
	})

	r.POST("/meditations", func(c *gin.Context) {
		var input CreateMeditationInput

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}

		u4 := uuid.NewV4()

		newMeditation := Meditation{
			ID:   u4.String(),
			URL:  input.URL,
			Name: input.Name,
		}

		SaveMeditation(newMeditation)

		c.JSON(http.StatusCreated, newMeditation)
	})

	r.Run()
}
