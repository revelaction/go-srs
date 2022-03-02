package badger_test

import (
	badger "github.com/dgraph-io/badger/v2"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/revelaction/go-srs/algo/sm2"
	"github.com/revelaction/go-srs/db"
	bdg "github.com/revelaction/go-srs/db/badger"
	"github.com/revelaction/go-srs/review"
)

func TestInsertNewDeckId(t *testing.T) {

	// Init badger
	dir, err := ioutil.TempDir(".", "badger")
	if err != nil {
		t.Errorf("got unexpected error %s", err)
	}

	defer os.RemoveAll(dir)

	opts := badger.DefaultOptions(dir)
	opts.Logger = nil

	bad, err := badger.Open(opts)

	if err != nil {
		t.Errorf("got unexpected error %s", err)
	}

	defer bad.Close()

	// Init Algo
	now := time.Date(2020, time.November, 1, 0, 0, 0, 0, time.UTC)
	sm2 := sm2.New(now)

	// Init Db
	dbh := bdg.New(bad, sm2)

	// Empty review without DeckId
	r := review.Review{}
	r.Items = []review.ReviewItem{
		{Quality: 4},
		{Quality: 3},
	}

	wantNewDeckId := "hi"

	due, err := dbh.Insert(r, wantNewDeckId)
	if err != nil {
		t.Errorf("got unexpected error %s", err)
	}

	if due.DeckId != wantNewDeckId {
		t.Errorf("\ngot %#v\nwant %#v", due.DeckId, wantNewDeckId)
	}

	wantLen := 2
	if len(due.Items) != wantLen {
		t.Errorf("\nCheking len:\ngot %d\nwant %d", len(due.Items), wantLen)
	}

	// Check card ids
	// New cards start with 1.
	for idx, item := range due.Items {
		wantId := idx + 1
		if item.CardId != wantId {
			t.Errorf("\ngot cardId %d\nwant cardId %d", item.CardId, wantId)
		}
	}
}

// TestInsertNotExistingDeckId tests that trying to add cards to a non existent
// DeckId produces error
func TestInsertNotExistingDeckId(t *testing.T) {

	// Init badger
	dir, err := ioutil.TempDir(".", "badger")
	if err != nil {
		t.Errorf("got unexpected error %s", err)
	}

	defer os.RemoveAll(dir)

	opts := badger.DefaultOptions(dir)
	opts.Logger = nil

	bad, err := badger.Open(opts)

	if err != nil {
		t.Errorf("got unexpected error %s", err)
	}

	defer bad.Close()

	// Algo
	now := time.Date(2020, time.November, 1, 0, 0, 0, 0, time.UTC)
	sm2 := sm2.New(now)

	// Db
	dbh := bdg.New(bad, sm2)

	// Empty review without DeckId
	r := review.Review{}
	r.DeckId = "existingDeckId"
	r.Items = []review.ReviewItem{
		{Quality: 4},
		{Quality: 3},
	}

	// Call Insert method with empty string to signal external DeckId
	_, err = dbh.Insert(r, "")

	if err != nil {
		if err != db.ErrDeckIdNotExists {
			t.Errorf("\ngot error %s\nwant ErrDeckIdNotExists", err)
		}
	}
}

func TestInsertNewAndReinsertWith(t *testing.T) {

	// Init badger
	dir, err := ioutil.TempDir(".", "badger")
	if err != nil {
		t.Errorf("got unexpected error %s", err)
	}

	defer os.RemoveAll(dir)

	opts := badger.DefaultOptions(dir)
	opts.Logger = nil

	bad, err := badger.Open(opts)

	if err != nil {
		t.Errorf("got unexpected error %s", err)
	}

	defer bad.Close()

	// Algo
	now := time.Date(2020, time.November, 1, 0, 0, 0, 0, time.UTC)
	sm2 := sm2.New(now)

	// Db
	dbh := bdg.New(bad, sm2)

	// Empty review without DeckId
	r := review.Review{}
	r.Items = []review.ReviewItem{
		{Quality: 4},
		{Quality: 3},
	}

	wantNewDeckId := "hi" // this is generated in the Srs object

	// Call Insert method with
	dueNew, err := dbh.Insert(r, wantNewDeckId)
	if err != nil {
		t.Errorf("got unexpected error %s", err)
	}

	wantLenNew := 2
	if len(dueNew.Items) != wantLenNew {
		t.Errorf("\nCheking len:\ngot %d\nwant %d", len(dueNew.Items), wantLenNew)
	}

    // we have now saved 2 card ids (1,2) in DeckId "hi". They are new so they
    // have Due time 1 day later as now.  check before, we have no Due cards to
    // review
	newNowBefore := now.Add(time.Hour * 24)
	dueBefore, err := dbh.Due(wantNewDeckId, newNowBefore)
	if err != nil {
		t.Errorf("got unexpected error %s", err)
	}

	wantLenBefore := 0
	if len(dueBefore.Items) != wantLenBefore {
		t.Errorf("\nChecking len:\ngot %d\nwant %d", len(dueBefore.Items), wantLenBefore)
	}

	// check after just one second, cards are now overdue
	newNowAfter := newNowBefore.Add(time.Second * 1)
	dueAfter, err := dbh.Due(wantNewDeckId, newNowAfter)
	if err != nil {
		t.Errorf("got unexpected error %s", err)
	}

	wantLenAfter := 2
	if len(dueAfter.Items) != wantLenAfter {
		t.Errorf("\nCheking len:\ngot %d\nwant %d", len(dueAfter.Items), wantLenAfter)
	}

	// Insert now two new cards two cards (3,4)
	r.DeckId = wantNewDeckId
	r.Items = []review.ReviewItem{
		{Quality: 3},
		{Quality: 4},
	}

	dueAdditional, err := dbh.Insert(r, "")
	if err != nil {
		t.Errorf("got unexpected error %s", err)
	}

	wantLenAdditional := 2
	if len(dueAdditional.Items) != wantLenAdditional {
		t.Errorf("\nCheking len:\ngot %d\nwant %d", len(dueAdditional.Items), wantLenAdditional)
	}

	//In one day we should have 4 overdue cards
	newNowFor4 := newNowBefore.Add(time.Hour * 24).Add(time.Second * 1)
	dueFor4, err := dbh.Due(wantNewDeckId, newNowFor4)
	if err != nil {
		t.Errorf("got unexpected error %s", err)
	}

	wantLenFor4 := 4
	if len(dueFor4.Items) != wantLenFor4 {
		t.Errorf("\nCheking len:\ngot %d\nwant %d", len(dueFor4.Items), wantLenFor4)
	}
}

func TestInsertAndUdpate(t *testing.T) {

	dir, err := ioutil.TempDir(".", "badger")
	if err != nil {
		t.Errorf("got unexpected error %s", err)
	}

	defer os.RemoveAll(dir)

	opts := badger.DefaultOptions(dir)
	opts.Logger = nil

	bad, err := badger.Open(opts)

	if err != nil {
		t.Errorf("got unexpected error %s", err)
	}

	defer bad.Close()

	// Algo
	now := time.Date(2020, time.November, 1, 0, 0, 0, 0, time.UTC)
	algo := sm2.New(now)

	// Db
	dbh := bdg.New(bad, algo)

	// Empty review without DeckId
	r := review.Review{}
	r.Items = []review.ReviewItem{
		{Quality: 4},
		{Quality: 3},
	}

	wantNewDeckId := "hi" 

	// Call Insert method with
	dueNew, err := dbh.Insert(r, wantNewDeckId)
	if err != nil {
		t.Errorf("got unexpected error %s", err)
	}

	wantLenNew := 2
	if len(dueNew.Items) != wantLenNew {
		t.Errorf("\nCheking len:\ngot %d\nwant %d", len(dueNew.Items), wantLenNew)
	}

    // we have now saved 2 card ids (1,2) in DeckId "hi". They are new so they
    // have Due time 1 day later as now.  We check 1 day and 1 second later
	newNowAfter := now.Add(time.Hour * 24).Add(time.Second * 1)
	dueAfter, err := dbh.Due(wantNewDeckId, newNowAfter)
	if err != nil {
		t.Errorf("got unexpected error %s", err)
	}

	wantLenAfter := 2
	if len(dueAfter.Items) != wantLenAfter {
		t.Errorf("\nCheking len:\ngot %d\nwant %d", len(dueAfter.Items), wantLenAfter)
	}

    // Let's move forward the time. That means a new now in the algo, we build
    // again the badger wrapper:
	// Algo
	algo = sm2.New(newNowAfter)
	// Db
	dbh = bdg.New(bad, algo)

	// Update now the two new cards two cards (3,4) with a bad review, meaning
	// they will be scheduled for 24 hours later again.
	r.DeckId = wantNewDeckId
	// Cards have now Id
	r.Items = []review.ReviewItem{
		{CardId: 1, Quality: 2},
		{CardId: 2, Quality: 2},
	}

	dueUpdate, err := dbh.Update(r)
	if err != nil {
		t.Errorf("got unexpected error %s", err)
	}

	wantLenUpdate := 2
	if len(dueUpdate.Items) != wantLenUpdate {
		t.Errorf("\nCheking len:\ngot %d\nwant %d", len(dueUpdate.Items), wantLenUpdate)
	}

	// In one day we should have the two cards overdue, not before
	newUpdateBefore := newNowAfter.Add(time.Hour * 24)
	dueUpdateBefore, err := dbh.Due(wantNewDeckId, newUpdateBefore)
	if err != nil {
		t.Errorf("got unexpected error %s", err)
	}

	wantLenUpdateBefore := 0
	if len(dueUpdateBefore.Items) != wantLenUpdateBefore {
		t.Errorf("\nChecking len:\ngot %d\nwant %d", len(dueUpdateBefore.Items), wantLenUpdateBefore)
	}

	// check after just one second, cards are now overdue
	newUpdateAfter := newUpdateBefore.Add(time.Second * 1)
	dueUpdateAfter, err := dbh.Due(wantNewDeckId, newUpdateAfter)
	if err != nil {
		t.Errorf("got unexpected error %s", err)
	}

	wantLenUpdateAfter := 2
	if len(dueUpdateAfter.Items) != wantLenUpdateAfter {
		t.Errorf("\nCheking len:\ngot %d\nwant %d", len(dueUpdateAfter.Items), wantLenUpdateAfter)
	}
}
