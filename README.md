<p align="center"><img alt="go-srs" src="logo.png"/></p>

[![Go Reference](https://pkg.go.dev/badge/github.com/revelaction/go-srs)](https://pkg.go.dev/github.com/revelaction/go-srs)
[![Test](https://github.com/revelaction/go-srs/actions/workflows/test.yml/badge.svg)](https://github.com/revelaction/go-srs/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/revelaction/go-srs)](https://goreportcard.com/report/github.com/revelaction/go-srs)
[![GitHub Release](https://img.shields.io/github/v/release/revelaction/go-srs?style=flat)]()
[![GitHub Release](https://img.shields.io/badge/built_with-Go-00ADD8.svg?style=flat)]() 

go-srs: Empowering Spaced Repetition Learning in Go

**go-srs** offers a comprehensive Go implementation of the [SuperMemo 2](https://www.supermemo.com/english/ol/sm2.htm) (sm2) algorithm, along with libraries designed to facilitate the creation of [spaced repetition
learning](https://en.wikipedia.org/wiki/Spaced_repetition)-capable programs and servers.

## Key Features: 

- **SuperMemo 2 (sm2) Algorithm Implementation:** go-srs includes a robust implementation of the SuperMemo 2 [(sm2) algorithm](algo/sm2/sm2.go).

- **Local Database Client:** Leveraging [badger](https://github.com/outcaste-io/badger), go-srs comes equipped with a local database client.

- **Unique ID Generator:** The library features a unique ID generator based on [ulid](https://github.com/oklog/ulid).

## Use

First, instantiate the srs struct with the provided algorithm, uid and db implementations:

```go
// Algo
now := time.Now().UTC()
sm2 := sm2.New(now)

// Db
opts := badger.DefaultOptions("./badger")
opts.Logger = nil
bad, err := badger.Open(opts)
defer bad.Close()
db := bdg.New(bad, sm2)

// uid
entropy := ulidPkg.Monotonic(crand.Reader, 0)
uid := ulid.New(entropy)

// instantiate srs struct
hdl := srs.New(db, uid)
```


After that, you can update and save card reviews of a given deck:

```go
r = review.Review{}
r.DeckId = "somedeckid"

r.Items = []review.ReviewItem{
    {CardId: 65, Quality: 1},
    {CardId: 3, Quality: 5},
    {CardId: 1, Quality: 0},
}

_, err = hdl.Update(r)
if err != nil {
    fatal(err)
}
```


And retrieve cards that are due for review

```go
tdue := time.Now().UTC()
dueCards, _ := hdl.Due(res.DeckId, tdue)
```


See `srs_test.go` for more examples.

## Additional implementations

`go-srs` provides interfaces for [db](db/db.go), [algorithm](algo/algo.go) and
[unique id](uid/uid.go) implementations.

