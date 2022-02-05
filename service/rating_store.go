package service

import (
	"sync"
)

type RatingStore interface {
	Rate(laptopId string, score float64) *Rating
}

type Rating struct {
	count int
	sum   float64
}

type InMemoryRatingStore struct {
	mutex   sync.RWMutex
	ratings map[string]*Rating
}

func NewInMemoryRatingStore() *InMemoryRatingStore {
	return &InMemoryRatingStore{
		ratings: make(map[string]*Rating),
	}
}

func (store *InMemoryRatingStore) Rate(laptopId string, score float64) *Rating {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	rating := store.ratings[laptopId]
	if rating == nil {
		rating = &Rating{
			count: 1,
			sum:   score,
		}
		store.ratings[laptopId] = rating
	} else {
		rating.count++
		rating.sum += score
	}

	return rating
}
