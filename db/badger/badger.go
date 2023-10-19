// Package badger implements the srs/db.Handler interface with a badger backend
package badger

import (
	"fmt"
	badger "github.com/outcaste-io/badger/v3"
	"strconv"
	"strings"
	"time"

	"github.com/revelaction/go-srs/algo"
	"github.com/revelaction/go-srs/db"
	"github.com/revelaction/go-srs/review"
)

// Handler is a badger client.
//
// It accepts an Algo to allow for atomic operations.
// (for example Update must read from the db, decode, compute new values
// according to the review, and write back to db)
type Handler struct {
	Db   *badger.DB
	Algo algo.Algo
}

func New(db *badger.DB, algo algo.Algo) *Handler {
	return &Handler{
		Db:   db,
		Algo: algo,
	}
}

// Insert runs the Algo on a all cards of a Review, and saves the result in the
// db.  It returns a slice of review Due structs containing the Due Date for
// each of the card ids. There are two cases:
//
// 1) New cards for existing DeckId: this requires a lookup of the last CardId in the db
// 2) New cards for new DeckId created in this session, which do no require
// lookup, index starts at 0.
//
// Insert is atomic
func (h *Handler) Insert(r review.Review, deckId string) (res review.Due, err error) {
	txn := h.Db.NewTransaction(true)
	defer txn.Discard()

	// New deck
	if deckId != "" {
		r.DeckId = deckId
		res, err = h.insertAfter(txn, 0, r)
		if err != nil {
			return res, err
		}

		if err := txn.Commit(); err != nil {
			return res, err
		}

		return res, nil
	}

	// Existent deckId
	// Look up the last card id and insert after
	var maxCardIdInDb int

	maxCardIdInDb, err = findMaxCardId(txn, r.DeckId)
	if err != nil {
		return res, err
	}

	res, err = h.insertAfter(txn, maxCardIdInDb, r)
	if err != nil {
		return res, err
	}

	// Commit the transaction and check for error.
	if err := txn.Commit(); err != nil {
		return res, err
	}

	return res, nil
}

// Update looks up in the db the (must) existing cards in the review r, run the
// srs algo on them, and saves the updated result in the db.  It returns a
// slice of reviews containing the Due Date for each of the card ids.
//
// The function is atomic
func (h *Handler) Update(r review.Review) (due review.Due, err error) {

	txn := h.Db.NewTransaction(true)
	defer txn.Discard()

	due, err = h.updateBetween(txn, r)
	if err != nil {
		return due, err
	}

	if err := txn.Commit(); err != nil {
		return due, err
	}

	return due, nil
}

// Due returs th Due cards for the time t
func (h *Handler) Due(deckId string, t time.Time) (due review.Due, err error) {

	due.DeckId = deckId

	// create transaction
	txn := h.Db.NewTransaction(true)
	defer txn.Discard()

	// iterate for the prefix
	opts := badger.DefaultIteratorOptions
	it := txn.NewIterator(opts)
	defer it.Close()

	prefixDeckId := []byte(deckId)

	for it.Seek(prefixDeckId); it.ValidForPrefix(prefixDeckId); it.Next() {
		item := it.Item()
		err := item.Value(func(v []byte) error {

			// This func with val would only be called if item.Value encounters no error.
			dueItem := h.Algo.Due(v, t)
			if dueItem.CardId > 0 {
				due.Items = append(due.Items, dueItem)
			}

			return nil
		})

		if err != nil {
			return due, err
		}
	}

	return due, nil
}

func findMaxCardId(txn *badger.Txn, deckId string) (int, error) {

	opts := badger.DefaultIteratorOptions
	opts.Reverse = true
	opts.PrefetchValues = false
	it := txn.NewIterator(opts)
	defer it.Close()

	var valCopy []byte
	prefixDeckId := []byte(deckId)

	// Reverse seek a prefix
	//
	// https://discuss.dgraph.io/t/reverse-seek-prefix-not-working-as-expected/8635/4
	// https://github.com/dgraph-io/badger/issues/436
	// https://github.com/dgraph-io/badger/issues/347
	// https://dgraph.io/docs/badger/faq/
	prefixSeek := []byte(deckId + "1000000")

	it.Seek(prefixSeek)
	if !it.ValidForPrefix(prefixDeckId) {
		return 0, db.ErrDeckIdNotExists
	}

	valCopy = it.Item().KeyCopy(nil)

	num, err := numberFromPaddedKey(valCopy)
	if err != nil {
		return 0, err
	}

	return num, nil
}

// Insert after max the cards that are new
// cardId start at 1 (no 0) for new cards
func (h *Handler) insertAfter(txn *badger.Txn, max int, r review.Review) (res review.Due, err error) {

	res = review.Due{}
	res.DeckId = r.DeckId

	for idx, ri := range r.Items {

		cardId := max + idx + 1

		// cardId must be injected, to be properly encoded in the Algo native struct
		ri.CardId = cardId
		b, err := h.Algo.Update(nil, ri)
		if err != nil {
			return res, err
		}

		key := buildKey(r.DeckId, cardId)

		if err := txn.Set(key, b); err != nil {
			return res, err
		}

		// add new Card Id to response
		res.Items = append(res.Items, review.DueItem{CardId: cardId})
	}

	return res, nil
}

// Update cards not new. All must exists
func (h *Handler) updateBetween(txn *badger.Txn, r review.Review) (due review.Due, err error) {

	due = review.Due{}
	due.DeckId = r.DeckId

	for _, ri := range r.Items {

		key := buildKey(r.DeckId, ri.CardId)
		v, err := txn.Get(key)
		// ErrKeyNotFound
		if err != nil {
			return due, err
		}

		valCopy, err := v.ValueCopy(nil)
		if err != nil {
			return due, err
		}

		b, err := h.Algo.Update(valCopy, ri)
		if err != nil {
			return due, err
		}

		if err := txn.Set(key, b); err != nil {
			// An ErrTxnTooBig will be reported in case the number of pending
			// writes/deletes in the transaction exceeds a certain limit. In
			// that case, it is best to commit the transaction and start a new
			// transaction immediately
			return due, err
		}

		// add new Card Id to response
		due.Items = append(due.Items, review.DueItem{CardId: ri.CardId})
	}

	return due, nil
}

func buildKey(boxId string, cardId int) []byte {
	return []byte(boxId + fmt.Sprintf("%06d", cardId))
}

func numberFromPaddedKey(key []byte) (int, error) {
	// get the last TODO len
	// remove the leading 0
	keyStr := string(key)
	numStr := strings.TrimLeft(keyStr[len(keyStr)-6:], "0")

	// convert to int
	s, err := strconv.Atoi(numStr)
	if err != nil {
		return 0, err
	}

	return s, nil
}
