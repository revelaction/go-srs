package srs

import (
	"time"

	"github.com/revelaction/go-srs/db"
	"github.com/revelaction/go-srs/review"
	"github.com/revelaction/go-srs/uid"
)

// Srs encapsulates the db handler and the UID generator.
//
// Deck and card ids are provided by this package
type Srs struct {
	Db  db.Handler
	UID uid.UID
}

func New(db db.Handler, id uid.UID) *Srs {
	return &Srs{
		Db:  db,
		UID: id,
	}
}

// Update update the cards in Review, persist then in th db and returns the
// updated Cards due time.
func (h *Srs) Update(r review.Review) (due review.Due, err error) {

	err = r.Validate()
	if err != nil {
		return due, err
	}

	//1) insert for no previous DeckId with the created boxId
	if r.DeckId == "" {
		// is external
		deckId := h.UID.Create()

		due, err = h.Db.Insert(r, deckId)
		if err != nil {
			return due, err
		}

		return due, nil
	}

	// 2) we have DeckId and all cards are new. Look up and insert after last
	// index
	if r.AllNewCards() {
		due, err = h.Db.Insert(r, "")
		if err != nil {
			return review.Due{}, err
		}

		return due, nil

	}

	//3) review has DeckId and cardIds: update for existent
	due, err = h.Db.Update(r)
	if err != nil {
		return review.Due{}, err
	}

	return due, nil
}

// Due returns all card ids that are due to be reviewed at time t
func (h *Srs) Due(deckId string, t time.Time) (due review.Due, err error) {

	due, err = h.Db.Due(deckId, t)
	if err != nil {
		return due, err
	}

	return due, nil
}
