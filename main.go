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
	URL  string `json:"audioUrl" binding:"required"`
	Name string `json:"name" binding:"required"`
}

var meditationsMap map[string][]Meditation

func SaveMeditation(meditation Meditation) {
	userSlice, ok := meditationsMap[meditation.UserId]

	if !ok {
		userSlice = []Meditation{meditation}
	} else {
		userSlice = append(userSlice, meditation)
	}
	meditationsMap[meditation.UserId] = userSlice
}

func ListMeditations(userId string) []Meditation {
	return meditationsMap[userId]
}

func GetMeditation(userId string, id string) (Meditation, error) {
	emptyMeditation := Meditation{
		Name: "",
		URL:  "",
		ID:   "",
	}
	meditations, ok := meditationsMap[userId]
	if !ok {
		return emptyMeditation, errors.New("No user with id " + userId + " was found")
	}
	for _, m := range meditations {
		if m.ID == id {
			return m, nil
		}
	}

	return emptyMeditation, errors.New("No meditation with id " + id + " was found")
}

func DeleteMeditation(userId string, id string) error {
	meditations, ok := meditationsMap[userId]
	if !ok {
		return errors.New("No user with id " + userId + " was found")
	}
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

		meditationsMap[userId] = meditations

		return nil
	}

	return errors.New("No meditation with id " + id + " was found")
}

func main() {
	r := gin.Default()

	meditationsMap = map[string][]Meditation{}

	r.GET("/meditations", func(c *gin.Context) {
		userId := c.GetHeader("User-Id")

		if userId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No User-Id header present"})
			return
		}

		meditations := ListMeditations(userId)
		if meditations == nil {
			c.JSON(http.StatusOK, [0]Meditation{})
			return
		}
		c.JSON(http.StatusOK, meditations)
	})

	r.GET("/meditations/:id", func(c *gin.Context) {
		id := c.Param("id")
		userId := c.GetHeader("User-Id")

		if userId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No User-Id header present"})
			return
		}

		m, err := GetMeditation(userId, id)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, m)
	})

	r.DELETE("/meditations/:id", func(c *gin.Context) {
		id := c.Param("id")
		userId := c.Request.Header["User-Id"][0]

		if err := DeleteMeditation(userId, id); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{})
	})

	r.POST("/meditations", func(c *gin.Context) {
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

		SaveMeditation(newMeditation)

		c.JSON(http.StatusCreated, newMeditation)
	})

	r.Run()
}
