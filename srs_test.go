package srs_test

import (
	ulidPkg "github.com/oklog/ulid/v2"
	badger "github.com/outcaste-io/badger/v3"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/revelaction/go-srs"
	"github.com/revelaction/go-srs/algo/sm2"
	bdg "github.com/revelaction/go-srs/db/badger"
	"github.com/revelaction/go-srs/review"
	"github.com/revelaction/go-srs/uid/ulid"
)

func TestInsertNewDeckIdAndDue(t *testing.T) {

	// Algo
	now := time.Date(2020, time.November, 1, 0, 0, 0, 0, time.UTC)
	sm2 := sm2.New(now)

	// Badger dir random
	dir, err := ioutil.TempDir(".", "badger")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(dir)

	opts := badger.DefaultOptions(dir)
	opts.Logger = nil
	bad, _ := badger.Open(opts)
	defer bad.Close()

	db := bdg.New(bad, sm2)

	// fixed  entropy value
	ti := time.Unix(1000000, 0)
	entropy := ulidPkg.Monotonic(rand.New(rand.NewSource(ti.UnixNano())), 0)

	uid := ulid.New(entropy)

	hdl := srs.New(db, uid)

	// Empty review without UID
	r := review.Review{}
	r.Items = []review.ReviewItem{
		{Quality: 4},
		{Quality: 3},
	}

	res, err := hdl.Update(r)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("☑  Created DeckId is  : %s", res.DeckId)

	// the new cards are made for tomorrow, check for two days
	tdue := time.Now().UTC().AddDate(0, 0, 2)
	due, err := hdl.Due(res.DeckId, tdue)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("☑  Response is : %#v", due)
	// 2 new cards, 2 inserted, 2 returned for next day
	wantLen := 2
	if len(due.Items) != wantLen {
		t.Errorf("\nCheking len:\ngot %d\nwant %d", len(due.Items), wantLen)
	}

	// Check cardId start with 1
	for idx, item := range due.Items {
		wantId := idx + 1
		if item.CardId != wantId {
			t.Errorf("\ngot cardId %d\nwant cardId %d", item.CardId, wantId)
		}
	}
}

func TestInsertWithDeckIdAndDue(t *testing.T) {

	// Algo
	now := time.Date(2020, time.November, 1, 0, 0, 0, 0, time.UTC)
	sm2 := sm2.New(now)

	// Badger dir random
	dir, err := ioutil.TempDir(".", "badger")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(dir) // clean up

	opts := badger.DefaultOptions(dir)
	opts.Logger = nil
	bad, _ := badger.Open(opts)
	defer bad.Close()

	db := bdg.New(bad, sm2)

	// fixed  entropy value
	ti := time.Unix(1000000, 0)
	entropy := ulidPkg.Monotonic(rand.New(rand.NewSource(ti.UnixNano())), 0)

	uid := ulid.New(entropy)
	//t.Logf("☑ uid is: %#v", uid.Create())

	hdl := srs.New(db, uid)

	// Empty review without UID
	r1 := review.Review{}
	r1.Items = []review.ReviewItem{
		{Quality: review.NoReview},
		{Quality: review.NoReview},
	}

	res, err := hdl.Update(r1)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("☑  review 1: Created DeckId is  : %s", res.DeckId)

	// Insert 3 additional cards in the created DeckId
	// Empty review without UID
	r2 := review.Review{}
	r2.DeckId = res.DeckId
	r2.Items = []review.ReviewItem{
		{Quality: review.NoReview},
		{Quality: review.NoReview},
		{Quality: review.NoReview},
	}

	res, err = hdl.Update(r2)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("☑  review 2: Created DeckId is  : %s", res.DeckId)

	// the new cards are made for tomorrow, check for 1 day 1 second, when it
	// becomes overdue
	tdue := time.Now().UTC().AddDate(0, 0, 1).Add(time.Second * 1)
	dueAfterReview2, err := hdl.Due(res.DeckId, tdue)
	if err != nil {
		t.Fatal(err)
	}

	wantLen := 5
	if len(dueAfterReview2.Items) != wantLen {
		t.Errorf("\nCheking len:\ngot %d\nwant %d", len(dueAfterReview2.Items), wantLen)
	}

	// Check cardId start with 1 -> 5
	for idx, item := range dueAfterReview2.Items {
		wantId := idx + 1
		if item.CardId != wantId {
			t.Errorf("\ngot cardId %d\nwant cardId %d", item.CardId, wantId)
		}
	}
}

func TestInsertUpdateDue(t *testing.T) {

	// Algo
	now := time.Date(2020, time.November, 1, 0, 0, 0, 0, time.UTC)
	sm2 := sm2.New(now)

	dir, err := ioutil.TempDir(".", "badger")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(dir) // clean up

	opts := badger.DefaultOptions(dir)
	opts.Logger = nil
	bad, _ := badger.Open(opts)
	defer bad.Close()

	db := bdg.New(bad, sm2)

	// fixed  entropy value
	ti := time.Unix(1000000, 0)
	entropy := ulidPkg.Monotonic(rand.New(rand.NewSource(ti.UnixNano())), 0)

	uid := ulid.New(entropy)

	hdl := srs.New(db, uid)

	// Empty review without UID
	r := review.Review{}
	r.Items = []review.ReviewItem{
		{Quality: 3},
		{Quality: 4},
	}

	res, err := hdl.Update(r)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("☑  Created new DeckId is  : %s", res.DeckId)
	t.Logf("☑  First Saved in db with Response is : %#v", res)

	// New review with existing DeckId
	r = review.Review{}
	r.DeckId = res.DeckId

	// Fill new review with retrieved
	for _, item := range res.Items {
		r.Items = append(r.Items, review.ReviewItem{CardId: item.CardId, Quality: 3}) // minimal
	}

	t.Logf("☑  created next review for same DeckId %s: %#v", r.DeckId, r)

	res, err = hdl.Update(r)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("☑  Second Saved in db with response : %#v", res)

	// the new cards are made for tomorrow, check for 30 days
	tdue := time.Now().UTC().AddDate(0, 0, 30)
	due, err := hdl.Due(res.DeckId, tdue)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("☑  Response is : %#v", due)
	// 2 new cards, 2 inserted, 2 returned for next day
	wantLen := 2
	if len(due.Items) != wantLen {
		t.Errorf("\nCheking len:\ngot %d\nwant %d", len(due.Items), wantLen)
	}

	// Check cardId start with 1
	for idx, item := range due.Items {
		wantId := idx + 1
		if item.CardId != wantId {
			t.Errorf("\ngot cardId %d\nwant cardId %d", item.CardId, wantId)
		}
	}

}
