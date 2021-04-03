package main

import "errors"

type MeditationStore interface {
	SaveMeditation(m Meditation)
	ListMeditations(userId string) []Meditation
	GetMeditation(userId string, id string) Meditation
	DeleteMeditation(userId string, id string) error
	UpdateMeditation(m Meditation)
}

type MemoryMeditationStore struct {
	meditationsMap map[string][]Meditation
}

func NewMemoryMeditationStore() *MemoryMeditationStore {
	store := MemoryMeditationStore{
		meditationsMap: map[string][]Meditation{},
	}
	return &store
}

func (store MemoryMeditationStore) SaveMeditation(meditation Meditation) {
	userSlice, ok := store.meditationsMap[meditation.UserId]

	if !ok {
		userSlice = []Meditation{meditation}
	} else {
		userSlice = append(userSlice, meditation)
	}
	store.meditationsMap[meditation.UserId] = userSlice
}

func (store MemoryMeditationStore) ListMeditations(userId string) []Meditation {
	return store.meditationsMap[userId]
}

func (store MemoryMeditationStore) GetMeditation(userId string, id string) (Meditation, error) {
	emptyMeditation := Meditation{
		Name: "",
		URL:  "",
		ID:   "",
	}
	meditations, ok := store.meditationsMap[userId]
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

func (store MemoryMeditationStore) DeleteMeditation(userId string, id string) error {
	meditations, ok := store.meditationsMap[userId]
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

		store.meditationsMap[userId] = meditations

		return nil
	}

	return errors.New("No meditation with id " + id + " was found")
}
