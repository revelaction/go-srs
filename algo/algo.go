package algo

import (
	"time"

	"github.com/revelaction/go-srs/review"
)

// Algo interface represents a srs algorithm
type Algo interface {

	// Update calculates the new algo parameters for the cards present in the
	// review r, based on the old parameters and the review quality of each
	// card.
	//
	// old is a []bytes serialization of the internal algo parameters.
	Update(old []byte, r review.ReviewItem) ([]byte, error)

	// Due retrieves the Due Items (card ids)  that are overdue for time t (UTC)
	Due(old []byte, t time.Time) review.DueItem
}
