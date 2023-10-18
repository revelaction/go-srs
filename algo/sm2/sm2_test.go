package sm2

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/revelaction/go-srs/review"
)

func ExampleUpdateNoReviewDue() {

	now := time.Date(2020, time.November, 1, 1, 0, 0, 0, time.UTC)

	sm2 := New(now)

	r := review.ReviewItem{CardId: 1, Quality: review.NoReview}
	item, _ := sm2.Update(nil, r)

	dueItem := sm2.Due(item, now.AddDate(0, 0, 2))

	fmt.Printf("Due Item is %#v\n", dueItem)

	//Output:
	//Due Item is review.DueItem{CardId:1}
}

func ExampleDue() {

	now := time.Date(2020, time.November, 1, 1, 0, 0, 0, time.UTC)

	sm2 := New(now)

	r := review.ReviewItem{CardId: 1, Quality: review.NoReview}
	item, _ := sm2.Update(nil, r)

	r = review.ReviewItem{CardId: 1, Quality: review.CorrectHard}
	item, _ = sm2.Update(item, r)

	fmt.Printf("Item is %s\n", item)

	//Output:
	//Item is {"CardId":1,"Easiness":2.72,"ConsecutiveCorrectAnswers":1,"Due":1604365200}
}

func ExampleAllCorrectHard() {

	now := time.Date(2020, time.November, 1, 0, 0, 0, 0, time.UTC)
	fmt.Printf("Now is '%s'\n", now.Format("2006-01-02 15:04:05"))

	r := review.ReviewItem{CardId: 1, Quality: review.CorrectHard}

	i1 := create(r, now)
	dueTime1 := time.Unix(i1.Due, 0)
	dueStr := dueTime1.Format("2006-01-02 15:04:05")
	fmt.Printf("Easiness:%.2f, ConsecutiveCorrectAnswers:%d, Due:%s\n", i1.Easiness, i1.ConsecutiveCorrectAnswers, dueStr)

	// 1) make a for each with n runs 10, with thesame review
	// 2) Make array of reviews sustomized, for each array, customized
	// always first cerate

	n := 10
	next := i1
	nextTime := dueTime1
	for i := 1; i <= n; i++ {
		next = update(next, r, nextTime)
		nextTime = time.Unix(next.Due, 0)

		dueStr := nextTime.Format("2006-01-02 15:04:05")
		fmt.Printf("Easiness:%.2f, ConsecutiveCorrectAnswers:%d, Due:%s\n", next.Easiness, next.ConsecutiveCorrectAnswers, dueStr)
	}

	//Output:
	//Now is '2020-11-01 00:00:00'
	//Easiness:2.72, ConsecutiveCorrectAnswers:1, Due:2020-11-02 00:00:00
	//Easiness:2.94, ConsecutiveCorrectAnswers:2, Due:2020-11-08 00:00:00
	//Easiness:3.16, ConsecutiveCorrectAnswers:3, Due:2020-11-26 00:00:00
	//Easiness:3.38, ConsecutiveCorrectAnswers:4, Due:2021-01-25 00:00:00
	//Easiness:3.60, ConsecutiveCorrectAnswers:5, Due:2021-09-14 00:00:00
	//Easiness:3.82, ConsecutiveCorrectAnswers:6, Due:2024-06-18 00:00:00
	//Easiness:4.04, ConsecutiveCorrectAnswers:7, Due:2037-10-29 00:00:00
	//Easiness:4.26, ConsecutiveCorrectAnswers:8, Due:2109-04-03 00:00:00
	//Easiness:4.48, ConsecutiveCorrectAnswers:9, Due:2527-07-04 00:00:00
	//Easiness:4.70, ConsecutiveCorrectAnswers:10, Due:5193-02-05 00:00:00
	//Easiness:4.92, ConsecutiveCorrectAnswers:11, Due:23577-07-20 00:00:00

}

func ExampleAllIncorrectEasy() {

	now := time.Date(2020, time.November, 1, 0, 0, 0, 0, time.UTC)
	fmt.Printf("Now is '%s'\n", now.Format("2006-01-02 15:04:05"))

	r := review.ReviewItem{CardId: 1, Quality: review.IncorrectEasy}

	i1 := create(r, now)
	dueTime1 := time.Unix(i1.Due, 0)
	dueStr := dueTime1.Format("2006-01-02 15:04:05")
	fmt.Printf("Easiness:%.2f, ConsecutiveCorrectAnswers:%d, Due:%s\n", i1.Easiness, i1.ConsecutiveCorrectAnswers, dueStr)

	// 1) make a for each with n runs 10, with thesame review
	// 2) Make array of reviews sustomized, for each array, customized
	// always first cerate

	n := 5
	next := i1
	nextTime := dueTime1
	for i := 1; i <= n; i++ {
		next = update(next, r, nextTime)
		nextTime = time.Unix(next.Due, 0)

		dueStr := nextTime.Format("2006-01-02 15:04:05")
		fmt.Printf("Easiness:%.2f, ConsecutiveCorrectAnswers:%d, Due:%s\n", next.Easiness, next.ConsecutiveCorrectAnswers, dueStr)
	}

	//Output:
	//Now is '2020-11-01 00:00:00'
	//Easiness:2.34, ConsecutiveCorrectAnswers:1, Due:2020-11-02 00:00:00
	//Easiness:2.18, ConsecutiveCorrectAnswers:0, Due:2020-11-03 00:00:00
	//Easiness:2.02, ConsecutiveCorrectAnswers:0, Due:2020-11-04 00:00:00
	//Easiness:1.86, ConsecutiveCorrectAnswers:0, Due:2020-11-05 00:00:00
	//Easiness:1.70, ConsecutiveCorrectAnswers:0, Due:2020-11-06 00:00:00
	//Easiness:1.54, ConsecutiveCorrectAnswers:0, Due:2020-11-07 00:00:00
}

func ExampleAllCorrectHardIncorrectEasyAlternate() {

	now := time.Date(2020, time.November, 1, 0, 0, 0, 0, time.UTC)
	fmt.Printf("Now is '%s'\n", now.Format("2006-01-02 15:04:05"))

	reviews := []review.ReviewItem{
		{CardId: 1, Quality: review.CorrectHard},
		{CardId: 1, Quality: review.IncorrectEasy},
		{CardId: 1, Quality: review.CorrectHard},
		{CardId: 1, Quality: review.IncorrectEasy},
		{CardId: 1, Quality: review.CorrectHard},
		{CardId: 1, Quality: review.IncorrectEasy},
	}

	r := review.ReviewItem{CardId: 1, Quality: review.NoReview}
	startItem := create(r, now)
	startDueTime := time.Unix(startItem.Due, 0)
	dueStr := startDueTime.Format("2006-01-02 15:04:05")
	fmt.Printf("Easiness:%.2f, ConsecutiveCorrectAnswers:%d, Due:%s\n", startItem.Easiness, startItem.ConsecutiveCorrectAnswers, dueStr)

	next := startItem
	nextTime := startDueTime

	for _, r := range reviews {
		next = update(next, r, nextTime)
		nextTime = time.Unix(next.Due, 0)
		dueStr := nextTime.Format("2006-01-02 15:04:05")
		fmt.Printf("Easiness:%.2f, ConsecutiveCorrectAnswers:%d, Due:%s\n", next.Easiness, next.ConsecutiveCorrectAnswers, dueStr)
	}

	//Output:
	//Now is '2020-11-01 00:00:00'
	//Easiness:2.50, ConsecutiveCorrectAnswers:0, Due:2020-11-02 00:00:00
	//Easiness:2.72, ConsecutiveCorrectAnswers:1, Due:2020-11-04 00:00:00
	//Easiness:2.56, ConsecutiveCorrectAnswers:0, Due:2020-11-05 00:00:00
	//Easiness:2.78, ConsecutiveCorrectAnswers:1, Due:2020-11-07 00:00:00
	//Easiness:2.62, ConsecutiveCorrectAnswers:0, Due:2020-11-08 00:00:00
	//Easiness:2.84, ConsecutiveCorrectAnswers:1, Due:2020-11-10 00:00:00
	//Easiness:2.68, ConsecutiveCorrectAnswers:0, Due:2020-11-11 00:00:00
}

func TestEasiness(t *testing.T) {
	tests := []struct {
		quality review.Quality
		want    float64
	}{
		{quality: review.IncorrectBlackout, want: 1.7},
		{quality: review.IncorrectFamiliar, want: 2},
		{quality: review.IncorrectEasy, want: 2.34},
		{quality: review.CorrectHard, want: 2.72},
		{quality: review.CorrectEffort, want: 3.14},
		{quality: review.CorrectEasy, want: 3.6},
	}

	for _, tc := range tests {

		es := easiness(DefaultEasiness, quality(tc.quality))
		if es != tc.want {
			t.Errorf("\ngot %#v\nwant %#v", es, tc.want)
		}
	}
}

func TestEasinessAllBlackout(t *testing.T) {

	want := MinEasiness

	es := DefaultEasiness
	for i := 1; i <= 5; i++ {
		es = easiness(es, quality(review.IncorrectBlackout))
		t.Logf("%d Easiness: %f", i, es)
	}

	if es != want {
		t.Errorf("\ngot %#v\nwant %#v", es, want)
	}
}

func TestEasinessAllCorrectHard(t *testing.T) {

	want := 4.7

	es := DefaultEasiness
	for i := 1; i <= 10; i++ {
		es = easiness(es, quality(review.CorrectHard))
		t.Logf("%d Easiness: %f", i, es)
	}

	if !floatEqual(es, want) {
		t.Errorf("\ngot %#v\nwant %#v", es, want)
	}
}

func TestNewCardWithoutReviewQuality(t *testing.T) {

	r := review.ReviewItem{CardId: 1, Quality: review.NoReview}

	now := time.Date(2020, time.November, 1, 1, 0, 0, 0, time.UTC)

	wantStartItem := Item{
		CardId:                    1,
		Easiness:                  DefaultEasiness,
		ConsecutiveCorrectAnswers: 0,
		Due:                       now.AddDate(0, 0, 1).Unix(),
	}

	haveItem := create(r, now)
	if haveItem != wantStartItem {
		t.Errorf("\ngot %#v\nwant %#v", haveItem, wantStartItem)
	}

	//t.Logf("☑  Created DeckId is  : %s", sm2.())
}

func TestNewCardWithReviewQualityBlackout(t *testing.T) {

	r := review.ReviewItem{CardId: 1, Quality: review.IncorrectBlackout}

	now := time.Date(2020, time.November, 1, 1, 0, 0, 0, time.UTC)

	wantStartItem := Item{
		CardId:                    1,
		Easiness:                  1.7,
		ConsecutiveCorrectAnswers: 1,
		Due:                       now.AddDate(0, 0, 1).Unix(),
	}

	haveItem := create(r, now)
	if haveItem != wantStartItem {
		t.Errorf("\ngot %#v\nwant %#v", haveItem, wantStartItem)
	}
}

func TestUpdateDecodeError(t *testing.T) {

	now := time.Date(2020, time.November, 1, 1, 0, 0, 0, time.UTC)

	sm2 := New(now)

	r := review.ReviewItem{CardId: 1, Quality: review.NoReview}
	b := []byte{'g', 'o', 'l', 'a', 'n', 'g'}
	_, err := sm2.Update(b, r)

	// We expect a error because no existing DeckId
	var e *json.SyntaxError
	if !errors.As(err, &e) {
		t.Errorf("\ngot error %s\nwant SyntaxError", err)
	}
}

// comparing floats
// https://floating-point-gui.de/errors/comparison/
// https://gist.github.com/cevaris/bc331cbe970b03816c6b
func floatEqual(a, b float64) bool {
	ab := float64(math.Float64bits(a))
	bb := float64(math.Float64bits(b))

	percent := math.Abs(ab-bb) / ab
	precision := 0.000000001 // nano
	if percent > precision {
		return false
	}

	return true
}
