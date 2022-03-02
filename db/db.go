package db

import (
	"errors"
	"time"

	"github.com/revelaction/go-srs/review"
)

var (

	// ErrDeckIdNotExists is returned when not found DeckId in the Db
	ErrDeckIdNotExists = errors.New("Deck Id does not exists")

	// ErrCardIdNotExists is returned when not found Card Id in the Db
	ErrCardIdNotExists = errors.New("Card Id does not exists.")
)

// Handler interface abstracts the persistence of the updated Cards following a Review.
//
// Implementations should make all methods atomic
// Implementation needs at the least a db backend and a srs algo.
type Handler interface {
	Update(r review.Review) (review.Due, error)
	Insert(r review.Review, boxId string) (review.Due, error)
	Due(deckId string, t time.Time) (review.Due, error)
}
