package review

import (
	"errors"
)

var (
	ErrInvalidCardId       = errors.New("card id is greater than max")
	ErrMixedCardId         = errors.New("the review contains mixed card ids")
	ErrCardIdWithoutDeckId = errors.New("card id given but no deck id")
	ErrInvalidQuality      = errors.New("invalid quality")
)

// Max num cards per DeckId
const MaxCardId = 100000

const (
	//  0: no review
	NoReview Quality = iota

	//  1: "Total blackout", complete failure to recall the information.
	IncorrectBlackout

	//  2: Incorrect response, but upon seeing the correct answer it felt familiar.
	IncorrectFamiliar

	//  3: Incorrect response, but upon seeing the correct answer it seemed easy to remember.
	IncorrectEasy

	//  4: Correct response, but required significant difficulty to recall.
	CorrectHard

	//  5: Correct response, after some hesitation.
	CorrectEffort

	//  6: Correct response with perfect recall.
	CorrectEasy
)

// Quality specifies a grade (from 0 to 5) indicating a user self-evaluation of
// the quality of their response, with each grade having the following meaning:
//
// This generic review quality is mostly taken from supermemo
// Each algo implemented should traslate this review quality to the own quality evaluation
type Quality int

func (q Quality) Validate() error {
	if q > CorrectEasy || q < NoReview {
		return ErrInvalidQuality
	}

	return nil
}

// Review contains the ReviewItem for a DeckId
// A DeckId is some external id that identify a collection of Cards belonging
// to some User.
type Review struct {
	DeckId string
	Items  []ReviewItem
}

// ReviewItem contains the Quality evaluated for the CardId
// A CardId is a unique id (for a given DeckId). If not externally provided,
// go-srs can provide one (currently based on ulid)
type ReviewItem struct {
	CardId  int
	Quality Quality
}

// Due contains all DueItems (CardId) that need to be reviewed.
type Due struct {
	DeckId string
	Items  []DueItem
}

type DueItem struct {
	CardId int
	//Due time.Time
}

// Validate the Review
//
// - No card id if no Deck id
// - if there is a deck id, all cards ids = 0, or all not 0
// - Card Id should not be greater than MaxCardId
// - Quality validation
func (r *Review) Validate() error {

	oneCardIdZero := false
	oneCardIdNonZero := false

	// CardId
	for _, item := range r.Items {

		// card Id should not exist if no Deck Id
		if item.CardId > 0 {
			if r.DeckId == "" {
				return ErrCardIdWithoutDeckId
			}
		}

		// All cards ids or none
		if r.DeckId != "" {
			if item.CardId > 0 {
				oneCardIdNonZero = true
			} else {
				oneCardIdZero = true
			}
		}

		if item.CardId >= MaxCardId {
			return ErrInvalidCardId
		}

		err := item.Quality.Validate()
		if err != nil {
			return err
		}
	}

	// All cards ids or none
	if r.DeckId != "" {
		if oneCardIdZero == oneCardIdNonZero {
			return ErrMixedCardId
		}
	}

	return nil
}

func (r *Review) AllNewCards() bool {
	return r.Items[0].CardId == 0
}
