package review_test

import (
	"testing"

    "github.com/revelaction/go-srs/review"
)

func TestValidateValid(t *testing.T) {

	r := review.Review{}
	r.DeckId = "hi"
	r.Items = []review.ReviewItem{
		{Quality: 4},
		{Quality: 3},
		{Quality: 0},
	}

	err := r.Validate()

	if err != nil {
		t.Errorf("\ngot error %s\nwant nil", err)
	}

}

func TestValidateInvalidQuality(t *testing.T) {

	// without deckId
	r := review.Review{}
	r.Items = []review.ReviewItem{
		{Quality: 4},
		{Quality: 3},
		{Quality: 8},
	}

	err := r.Validate()

	if err != review.ErrInvalidQuality {
		t.Errorf("\ngot error %s\nwant ErrDeckIdNotExists", err)
	}
}

func TestValidateCardIdWithoutDeckId(t *testing.T) {

	r := review.Review{}
	r.Items = []review.ReviewItem{
		{CardId: 10, Quality: 4},
	}

	err := r.Validate()

	if err != review.ErrCardIdWithoutDeckId {
		t.Errorf("\ngot error %s\nwant ErrCardIdWithoutDeckId", err)
	}

}

func TestValidateCardsWithIdAndWithout(t *testing.T) {

	r := review.Review{}
	r.DeckId = "hi"
	r.Items = []review.ReviewItem{
		{CardId: 10, Quality: 4},
		{Quality: 5},
	}

	err := r.Validate()

	if err != review.ErrMixedCardId {
		t.Errorf("\ngot error %s\nwant ErrMixedCardId", err)
	}

}

func TestValidateCardIdGreaterThanMax(t *testing.T) {

	r := review.Review{}
	r.DeckId = "hi"

	r.Items = []review.ReviewItem{
		{CardId: 3000000, Quality: 4},
	}

	err := r.Validate()

	if err != review.ErrInvalidCardId {
		t.Errorf("\ngot error %s\nwant ErrInvalidCardId", err)
	}

}

func TestAllNewCards(t *testing.T) {

	r := review.Review{}
	r.DeckId = "hi"
	r.Items = []review.ReviewItem{
		{Quality: 4},
		{Quality: 3},
		{Quality: 8},
	}

	if !r.AllNewCards() {
		t.Errorf("expected all new cards")
	}
}
