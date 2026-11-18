package records

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func open(t *testing.T) (*Book, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "records.json")
	b, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	return b, path
}

func names(es []Entry) string {
	var s []string
	for _, e := range es {
		s = append(s, e.Name)
	}
	return strings.Join(s, ",")
}

func TestATableKeepsTheFiveBestResults(t *testing.T) {
	b, _ := open(t)
	for i, score := range []int{30, 10, 50, 20, 40, 60} {
		if _, err := b.Add("snake", Entry{Name: string(rune('a' + i)), Score: score}); err != nil {
			t.Fatal(err)
		}
	}
	if got := names(b.Top("snake")); got != "f,c,e,a,d" {
		t.Fatalf("top = %s, want the five best, best first", got)
	}
	if b.Qualifies("snake", Entry{Score: 20}) || !b.Qualifies("snake", Entry{Score: 21}) {
		t.Fatalf("a full table takes only results that beat its last one")
	}
	if rank, _ := b.Add("snake", Entry{Name: "g", Score: 5}); rank != 0 || len(b.Top("snake")) != Size {
		t.Fatalf("a result off the table is not added, rank %d", rank)
	}
	if len(b.Top("runner")) != 0 {
		t.Fatalf("each game has its own table")
	}
}

func TestAddReportsTheRankAndTiesKeepTheOlderFirst(t *testing.T) {
	b, _ := open(t)
	b.Add("snake", Entry{Name: "old", Score: 10})
	rank, _ := b.Add("snake", Entry{Name: "new", Score: 10})
	if rank != 2 || names(b.Top("snake")) != "old,new" {
		t.Fatalf("rank %d, top %s: an equal score goes after the older one", rank, names(b.Top("snake")))
	}
	if rank, _ := b.Add("snake", Entry{Name: "top", Score: 11}); rank != 1 {
		t.Fatalf("rank %d, want 1", rank)
	}
}

func TestLessTimeBreaksATie(t *testing.T) {
	b, _ := open(t)
	b.Add("sokoban", Entry{Name: "slow", Score: 3, Seconds: 300})
	b.Add("sokoban", Entry{Name: "fast", Score: 3, Seconds: 200})
	b.Add("sokoban", Entry{Name: "more", Score: 4, Seconds: 900})
	if got := names(b.Top("sokoban")); got != "more,fast,slow" {
		t.Fatalf("top = %s, want more score first, then less time", got)
	}
}

func TestNothingIsNotARecord(t *testing.T) {
	b, _ := open(t)
	if b.Qualifies("snake", Entry{Score: 0}) {
		t.Fatalf("a score of 0 never enters the table")
	}
}

func TestNamesAreTrimmedCutAndNeverEmpty(t *testing.T) {
	b, _ := open(t)
	b.Add("snake", Entry{Name: "   ", Score: 3})
	b.Add("snake", Entry{Name: "  a very long player name ", Score: 2})
	b.Add("snake", Entry{Name: "tab\there", Score: 1})
	top := b.Top("snake")
	if top[0].Name != DefaultName || top[1].Name != "a very long " || len([]rune(top[1].Name)) != MaxName || top[2].Name != "tabhere" {
		t.Fatalf("names %q", names(top))
	}
}

func TestRecordsSurviveReopening(t *testing.T) {
	b, path := open(t)
	b.Add("snake", Entry{Name: "ana", Score: 7})
	b.Add("sokoban", Entry{Name: "bo", Score: 2, Seconds: 61})
	again, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if top := again.Top("sokoban"); len(top) != 1 || top[0] != (Entry{Name: "bo", Score: 2, Seconds: 61}) || again.Top("snake")[0].Name != "ana" {
		t.Fatalf("reopened records differ: %+v %+v", again.Top("snake"), top)
	}
	data, _ := os.ReadFile(path)
	if !strings.Contains(string(data), `"version": 1`) {
		t.Fatalf("the file carries its format version:\n%s", data)
	}
}

func TestTopIsACopy(t *testing.T) {
	b, _ := open(t)
	b.Add("snake", Entry{Name: "ana", Score: 7})
	b.Top("snake")[0].Name = "x"
	if b.Top("snake")[0].Name != "ana" {
		t.Fatalf("Top must not expose the table")
	}
}

func TestAMissingDirectoryIsCreated(t *testing.T) {
	path := filepath.Join(t.TempDir(), "a", "b", "records.json")
	b, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := b.Add("snake", Entry{Name: "ana", Score: 1}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}

func TestAnUnreadableFileIsKeptAsideAndAFreshOneStarts(t *testing.T) {
	for _, content := range []string{"{not json", `{"version": 99, "games": {}}`, `{"version": 1, "games": {"snake": [{"name": "x", "score": "a"}]}}`} {
		path := filepath.Join(t.TempDir(), "records.json")
		os.WriteFile(path, []byte(content), 0o600)
		b, err := Open(path)
		if err != nil {
			t.Fatal(err)
		}
		if len(b.Top("snake")) != 0 {
			t.Fatalf("a fresh book starts empty")
		}
		if bad, _ := os.ReadFile(path + ".bad"); string(bad) != content {
			t.Fatalf("the unreadable file must be kept as .bad, got %q", bad)
		}
		if _, err := b.Add("snake", Entry{Name: "ana", Score: 1}); err != nil {
			t.Fatal(err)
		}
	}
}

func TestAFailedSaveReportsAndKeepsTheOldFile(t *testing.T) {
	b, path := open(t)
	b.Add("snake", Entry{Name: "ana", Score: 1})
	before, _ := os.ReadFile(path)
	dir := filepath.Dir(path)
	os.Chmod(dir, 0o500)
	defer os.Chmod(dir, 0o700)
	if _, err := b.Add("snake", Entry{Name: "bo", Score: 2}); err == nil {
		t.Fatalf("a save that cannot write must report an error")
	}
	if after, _ := os.ReadFile(path); string(after) != string(before) {
		t.Fatalf("a failed save must leave the old file whole")
	}
}

func TestAnInMemoryBookNeverWrites(t *testing.T) {
	b, err := Open("")
	if err != nil {
		t.Fatal(err)
	}
	if rank, err := b.Add("snake", Entry{Name: "ana", Score: 1}); rank != 1 || err != nil {
		t.Fatalf("rank %d err %v", rank, err)
	}
}

func TestABoardIsOneGamesTable(t *testing.T) {
	b, _ := open(t)
	board := b.Board("snake")
	board.Add(Entry{Name: "ana", Score: 3})
	if !board.Qualifies(Entry{Score: 1}) || board.Top()[0].Name != "ana" || b.Top("snake")[0].Name != "ana" {
		t.Fatalf("a board reads and writes its game's table")
	}
}
