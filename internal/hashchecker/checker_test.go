package hashchecker_test

// import (
// 	"Naverno/internal/hashchecker"
// 	"Naverno/internal/piece"
// 	"Naverno/internal/storage/discardstorage"
// 	"crypto/sha1"
// 	"testing"
// 	"time"
// )
//
// func TestCheck(t *testing.T) {
// 	s := discardstorage.New()
// 	p := piece.NewPiece(4, 10, 30, [20]byte{0})
// 	p.Hash = sha1.Sum(make([]byte, p.Size))
// 	w := hashchecker.New(s, p)
// 	res := make(chan *hashchecker.HashChecker)
//
// 	go w.Run(res)
//
// 	testTime := time.NewTimer(time.Second * 2)
// 	select {
// 	case result := <-res:
// 		if result.Err != nil {
// 			t.Fatalf("unexpected error -> %v", result.Err)
// 		}
// 		if !result.Matches {
// 			t.Errorf("the hash doesn't match")
// 		}
// 	case <-testTime.C:
// 		t.Fatal("excedeed test time limit")
// 	}
//
// }
