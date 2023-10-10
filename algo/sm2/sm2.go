// Package sm2 is an implementation of the supermemo algorithm.
// See // https://en.wikipedia.org/wiki/SuperMemo,
// This implmentation includes changes from
// http://www.blueraja.com/blog/477/a-better-spaced-repetition-learning-algorithm-sm2
//
// algorithm SM-2 (wikipedia) is:
//
//     input:  user grade q
//             repetition number n
//             easiness factor EF
//             interval I
//     output: updated values of n, EF, and I
//
//     if q ≥ 3 (correct response) then
//         if n = 0 then
//             I ← 1
//         else if n = 1 then
//             I ← 6
//         else
//             I ← ⌈I × EF⌉
//         end if
//         EF ← EF + (0.1 − (5 − q) × (0.08 + (5 − q) × 0.02)
//         if EF < 1.3 then
//             EF ← 1.3
//         end if
//         increment n
//     else (incorrect response)
//         n ← 0
//         I ← 1
//     end if
//
//     return (n, EF, I)
//
//
// Blueraja and wikipedia differ in some aspects:
//
// 1) EF in wikipedia is changed only with correct responses, in bluraja
// always, independent of q
//
// 2) blueraja increments n and uses it as input for Due days. We follow here the
// wiki version
//
// 3) blueraja calculates EF and uses it as input for Due days. We follow here
// the wiki version
package sm2

import (
	"encoding/json"
	"math"
	"time"

	"github.com/revelaction/go-srs/review"
)

const (
	DefaultEasiness = 2.5
	MinEasiness     = 1.3

	EasinessConst     = -0.8
	EasinessLineal    = 0.28
	EasinessQuadratic = 0.02

	DueDateStartDays   = 6
	IncorrectThreshold = 3.0
)

type Item struct {
	CardId int

	Easiness float64

	ConsecutiveCorrectAnswers int

	// Unix timestamp
	Due int64
}

type Sm2 struct {
	// UTC
	now time.Time
}

func New(now time.Time) *Sm2 {
	return &Sm2{now: now}
}

// Update takes a serialized representation of a Item, deserializes it and
// calculates a modified version according to the review
//
// The serialized version allows to avoid exposure of the internal algo Item
// details
func (s *Sm2) Update(oldItem []byte, r review.ReviewItem) ([]byte, error) {

	newItem := Item{}
	// if empty -> new no ned to decode
	if nil != oldItem {
		decodedItem, err := decode(oldItem)
		if err != nil {
			return nil, err
		}

		newItem = update(decodedItem, r, s.now)
	} else {
		newItem = create(r, s.now)
	}

	encodedItem, err := encode(newItem)
	if err != nil {
		return nil, err
	}

	return encodedItem, nil
}

// Due determines if the serialized Item item is overdue.
func (s *Sm2) Due(item []byte, t time.Time) (d review.DueItem) {
	dec, err := decode(item)
	if err != nil {
		return d
	}
	if dec.Due < t.Unix() {
		//gives unix time stamp in utc decItem.Due}
		return review.DueItem{CardId: dec.CardId}
	}

	return d
}

// create returns an Item after after processing the review
func create(r review.ReviewItem, now time.Time) Item {

	n := Item{}
	n.CardId = r.CardId

	if r.Quality == review.NoReview {
		n.ConsecutiveCorrectAnswers = 0
		n.Easiness = DefaultEasiness
	} else {
		// thsi is the first review for a new card
		n.Easiness = easiness(DefaultEasiness, quality(r.Quality))
		n.ConsecutiveCorrectAnswers = 1
	}

	n.Due = now.AddDate(0, 0, 1).Unix()
	return n

}

// easiness calculates the easiness factor.
func easiness(old float64, q float64) float64 {
	v := old + EasinessConst + (EasinessLineal * q) + (EasinessQuadratic * math.Pow(q, 2))
	if v < MinEasiness {
		return MinEasiness
	}

	return v
}

// quality traslates the user review self-evaluation to sm2 own metric
// sm2 uses 0, 5, which corresponds to 1,6 from the review.
func quality(q review.Quality) float64 {
	return float64(q - 1)
}

// update updates the internal sm2 parameters.
func update(old Item, r review.ReviewItem, now time.Time) Item {

	n := Item{}
	n.CardId = r.CardId

	// Easiness
	// days: bluraja seems wrong with Easiness from new instead of old.
	// wikipedia is correct here (old )
	n.Easiness = easiness(old.Easiness, quality(r.Quality))

	// Due
	// bluraja seems wrong with ConsecutiveCorrectAnswers from new instead of
	// old. wikipedia is correct here (increase days after)
	if r.Quality >= review.CorrectHard {
		days := float64(DueDateStartDays) * math.Pow(old.Easiness, float64(old.ConsecutiveCorrectAnswers-1))
		n.Due = now.AddDate(0, 0, int(math.Round(days))).Unix()
	} else {
		n.Due = now.AddDate(0, 0, 1).Unix()
	}

	// ConsecutiveCorrectAnswers
	if r.Quality >= review.CorrectHard {
		n.ConsecutiveCorrectAnswers = old.ConsecutiveCorrectAnswers + 1
	} else {
		n.ConsecutiveCorrectAnswers = 0
	}

	return n
}

// deserialize
func decode(encodedItem []byte) (Item, error) {
	res := Item{}
	errJson := json.Unmarshal(encodedItem, &res)
	if errJson != nil {
		return res, errJson
	}

	return res, nil
}

// serialize
func encode(item Item) ([]byte, error) {
	b, err := json.Marshal(item)
	if err != nil {
		return nil, err
	}

	return b, nil
}
