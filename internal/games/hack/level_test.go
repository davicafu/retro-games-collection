package hack

import (
	"math/rand"
	"strings"
	"testing"
)

// at puts g on level lvl, owning ids and standing on the last one.
func at(g *Game, lvl int, ids ...string) {
	g.loadLevel(lvl)
	for _, id := range ids {
		g.owned[id] = true
		g.node = id
	}
}

func TestReadingALevelsSecretOpensTheNextOne(t *testing.T) {
	g := newTestGame(Won)
	at(g, 0, "core")
	run(g, "hint")
	run(g, "cat lantern.txt")
	if g.Over() || g.Level() != 0 || !strings.Contains(output(g), "NIVEL 1 SUPERADO") {
		t.Fatalf("over %v level %d, want the level cleared and waiting:\n%s", g.Over(), g.Level(), output(g))
	}
	run(g, "ls")
	if g.Level() != 1 || g.Node() != "gate" || g.Owned("core") || g.Trace() != 0 || g.Hints() != 0 {
		t.Fatalf("enter must open level 2 fresh: level %d node %q trace %d hints %d", g.Level(), g.Node(), g.Trace(), g.Hints())
	}
	if text := g.Text(); !strings.Contains(text, "NIVEL 2") || strings.Contains(text, "lantern") {
		t.Fatalf("level 2 must start on a clear screen with its intro:\n%s", text)
	}
	if g.Score() != 2*ScoreStart-HintCost {
		t.Fatalf("score %d, want level 1's %d plus a fresh level", g.Score(), ScoreStart-HintCost)
	}
}

func TestTheLastLevelsSecretWins(t *testing.T) {
	g := newTestGame(Won)
	at(g, len(levels)-1, "core")
	run(g, "decrypt halcyon.enc lhk")
	if !g.Over() || !g.Won() {
		t.Fatalf("the last level's secret must win: over %v won %v", g.Over(), g.Won())
	}
}

func TestDecryptingAnEncryptedSecretClearsTheLevel(t *testing.T) {
	g := newTestGame(Won)
	at(g, 1, "datos")
	run(g, "decrypt hx12_real.enc 21xh")
	if !strings.Contains(output(g), "NIVEL 2 SUPERADO") {
		t.Fatalf("decrypting the secret must clear the level:\n%s", output(g))
	}
}

func TestAFirewallKeepsAMachineClosedUntilItIsYours(t *testing.T) {
	g := newTestGame(Won)
	at(g, 1, "gate", "web")
	run(g, "connect lab")
	if g.Prompt() != "ghost@web:~$ " || g.Trace() != TraceFailed || !strings.Contains(output(g), "fw") {
		t.Fatalf("prompt %q trace %d output %q, want lab closed by fw", g.Prompt(), g.Trace(), output(g))
	}
	if g.Status("lab") != "[bloqueado: fw]" {
		t.Fatalf("status %q", g.Status("lab"))
	}
	run(g, "connect fw")
	run(g, "connect lab")
	if !strings.Contains(g.Prompt(), "contraseña") {
		t.Fatalf("with fw yours, lab must ask its password, prompt %q", g.Prompt())
	}
}

func TestAHoneypotRaisesTheTraceAndStaysClosed(t *testing.T) {
	g := newTestGame(Won)
	at(g, 2, "gate", "vpn")
	if g.Status("admin") != "[abierto]" {
		t.Fatalf("a honeypot must look open, status %q", g.Status("admin"))
	}
	run(g, "connect admin")
	if g.Trace() != TraceHoneypot || g.Owned("admin") || g.Node() != "vpn" || !strings.Contains(output(g), "honeypot") {
		t.Fatalf("trace %d owned %v node %q output %q", g.Trace(), g.Owned("admin"), g.Node(), output(g))
	}
}

func TestALostRunKeepsOnlyTheFinishedLevels(t *testing.T) {
	g := newTestGame(Lost)
	at(g, 0, "core")
	run(g, "cat lantern.txt")
	run(g, "")
	for i := 0; i < 10 && !g.Over(); i++ {
		run(g, "connect web")
	}
	if !g.Over() || g.Won() || g.Score() != ScoreStart || g.Level() != 1 {
		t.Fatalf("over %v won %v score %d level %d, want caught on level 2 with level 1's score", g.Over(), g.Won(), g.Score(), g.Level())
	}
}

func TestBreachesUseTheirLevelsDifficulty(t *testing.T) {
	g := NewGame(Config{Rand: rand.New(rand.NewSource(1))})
	at(g, 2, "gate")
	run(g, "connect cam")
	ice, ok := g.Breach().(*ice)
	if !ok || len(ice.Sentries()) != difficulties[2].Sentries {
		t.Fatalf("level 3's ICE must have %d sentries", difficulties[2].Sentries)
	}
}

func TestEveryLevelIsWellFormed(t *testing.T) {
	for lvl, l := range levels {
		ids := map[string]Node{}
		for _, n := range l.Nodes {
			ids[n.ID] = n
		}
		finals := 0
		for _, n := range l.Nodes {
			for _, id := range append(append([]string(nil), n.Links...), n.Requires) {
				if _, ok := ids[id]; id != "" && !ok {
					t.Fatalf("level %d: %s names unknown machine %q", lvl+1, n.ID, id)
				}
			}
			for _, f := range n.Files {
				if f.Final {
					finals++
				}
			}
		}
		if finals != 1 || ids["gate"].ID == "" || len(l.Hints) == 0 || l.Hints[0].When != nil {
			t.Fatalf("level %d: %d final files, gate %v, hints %d", lvl+1, finals, ids["gate"].ID != "", len(l.Hints))
		}
	}
}

func TestWatchedMachinesRaiseTheTraceUntilTheIDSIsYours(t *testing.T) {
	g := newTestGame(Won)
	at(g, 3, "gate")
	steps := func(n int) {
		for i := 0; i < n; i++ {
			g.Step(Input{})
		}
	}
	steps(3 * difficulties[3].WatchTicks)
	if g.Trace() != 0 {
		t.Fatalf("the gateway is not watched, trace %d", g.Trace())
	}
	g.owned["puerto"], g.node = true, "puerto"
	if !g.Watched("puerto") || g.Watched("gate") {
		t.Fatalf("puerto is watched and the gateway is not")
	}
	steps(3 * difficulties[3].WatchTicks)
	if g.Trace() != 3*TraceWatched {
		t.Fatalf("trace %d after 3 s on a watched machine, want %d", g.Trace(), 3*TraceWatched)
	}
	g.owned["ids"] = true
	steps(3 * difficulties[3].WatchTicks)
	if g.Trace() != 3*TraceWatched {
		t.Fatalf("with the ids yours the trace must stop rising, %d", g.Trace())
	}
	if g.Watched("puerto") {
		t.Fatalf("with the ids yours no machine is watched")
	}
}

func TestTwoWrongPasswordsLockAMachineThatThenOnlyABreachOpens(t *testing.T) {
	g := newTestGame(Won)
	at(g, 4, "gate", "archivo")
	run(g, "connect presidencia")
	run(g, "marea88")
	if g.Status("presidencia") != "[contraseña]" || g.Trace() != TracePassword {
		t.Fatalf("one wrong try must not lock: status %q trace %d", g.Status("presidencia"), g.Trace())
	}
	run(g, "connect presidencia")
	run(g, "halcyon2020")
	if !strings.Contains(output(g), "bloqueada") || g.Status("presidencia") != "[forzar ICE]" {
		t.Fatalf("two wrong tries must lock it: status %q output %q", g.Status("presidencia"), output(g))
	}
	run(g, "connect presidencia")
	if g.Node() != "presidencia" || strings.Contains(g.Prompt(), "contraseña") {
		t.Fatalf("a locked machine is forced by its breach, at %q prompt %q", g.Node(), g.Prompt())
	}
}

func TestTheRightPasswordOpensBeforeTheLockout(t *testing.T) {
	g := newTestGame(Won)
	at(g, 4, "gate", "archivo")
	run(g, "connect presidencia")
	run(g, "marea88")
	run(g, "connect presidencia")
	run(g, "aurora1959")
	if g.Node() != "presidencia" {
		t.Fatalf("the right password on the second try must open it, at %q", g.Node())
	}
}
