package hack

import (
	"math/rand"
	"strings"
	"testing"
	"time"
)

// stubBreach ends with the outcome it is given on its first tick.
type stubBreach struct{ outcome Outcome }

func (b *stubBreach) Step(Input)       {}
func (b *stubBreach) Outcome() Outcome { return b.outcome }
func (b *stubBreach) TicksLeft() int   { return 1 }
func (b *stubBreach) Kind() BreachKind { return NoBreach }

// newTestGame is a game whose breaches end at once with outcome.
func newTestGame(outcome Outcome) *Game {
	g := NewGame(Config{Rand: rand.New(rand.NewSource(1))})
	g.newBreach = func(BreachKind) Breach { return &stubBreach{outcome: outcome} }
	return g
}

// run types line and presses enter, then lets a breach, if any, finish.
func run(g *Game, line string) {
	g.Step(Input{Text: []rune(line)})
	g.Step(Input{Enter: true})
	for i := 0; i < 3 && g.Mode() == ModeBreach; i++ {
		g.Step(Input{})
	}
}

// output is the terminal's text after the last command line.
func output(g *Game) string {
	lines := g.Lines()
	last := 0
	for i, l := range lines {
		if l.Kind == LineCommand {
			last = i
		}
	}
	var out []string
	for _, l := range lines[last+1:] {
		out = append(out, l.Text)
	}
	return strings.Join(out, "\n")
}

func TestStartsAtTheGatewayWithAPromptAndTheBrief(t *testing.T) {
	g := newTestGame(Won)
	if g.Node() != "gate" || !g.Owned("gate") || g.Mode() != ModeTerminal || g.Trace() != 0 {
		t.Fatalf("node %q owned %v mode %v trace %d", g.Node(), g.Owned("gate"), g.Mode(), g.Trace())
	}
	if g.Prompt() != "ghost@gate:~$ " {
		t.Fatalf("prompt %q", g.Prompt())
	}
	if !g.TakingText() {
		t.Fatalf("the terminal must take typed text")
	}
	if all := g.Text(); !strings.Contains(all, "help") {
		t.Fatalf("the welcome must point to help:\n%s", all)
	}
}

func TestTypingBackspaceAndEnterEditAndRunTheLine(t *testing.T) {
	g := newTestGame(Won)
	g.Step(Input{Text: []rune("lsx")})
	g.Step(Input{Backspace: true})
	if g.Buffer() != "ls" {
		t.Fatalf("buffer %q, want ls", g.Buffer())
	}
	g.Step(Input{Enter: true})
	if g.Buffer() != "" || !strings.Contains(g.Text(), "ghost@gate:~$ ls") || !strings.Contains(output(g), "readme.txt") {
		t.Fatalf("enter must echo the line and run it:\n%s", g.Text())
	}
}

func TestUpRecallsThePreviousCommand(t *testing.T) {
	g := newTestGame(Won)
	run(g, "scan")
	run(g, "ls")
	g.Step(Input{History: -1})
	g.Step(Input{History: -1})
	if g.Buffer() != "scan" {
		t.Fatalf("buffer %q, want scan after two ups", g.Buffer())
	}
	g.Step(Input{History: 1})
	if g.Buffer() != "ls" {
		t.Fatalf("buffer %q, want ls after a down", g.Buffer())
	}
}

func TestHelpLsCatAndClear(t *testing.T) {
	g := newTestGame(Won)
	run(g, "help")
	for _, cmd := range []string{"ls", "cat", "scan", "connect", "decrypt", "map", "clear"} {
		if !strings.Contains(output(g), cmd) {
			t.Fatalf("help must list %s:\n%s", cmd, output(g))
		}
	}
	run(g, "cat readme.txt")
	if !strings.Contains(output(g), "Proyecto Lantern") {
		t.Fatalf("cat must print the file:\n%s", output(g))
	}
	run(g, "clear")
	if len(g.Lines()) != 0 {
		t.Fatalf("clear must empty the screen, %d lines", len(g.Lines()))
	}
	if g.Trace() != 0 {
		t.Fatalf("good commands must not raise the trace, %d", g.Trace())
	}
}

func TestFailedCommandsRaiseTheTrace(t *testing.T) {
	g := newTestGame(Won)
	run(g, "sudo rm -rf /")
	if !strings.Contains(output(g), "comando no encontrado") || g.Trace() != TraceFailed {
		t.Fatalf("output %q trace %d", output(g), g.Trace())
	}
	run(g, "cat nothing.txt")
	run(g, "connect nowhere")
	if g.Trace() != 3*TraceFailed {
		t.Fatalf("trace %d, want %d after three failed commands", g.Trace(), 3*TraceFailed)
	}
	run(g, "")
	if g.Trace() != 3*TraceFailed {
		t.Fatalf("an empty line must not count")
	}
}

func TestScanShowsLinkedNodesAndTheirState(t *testing.T) {
	g := newTestGame(Won)
	run(g, "scan")
	out := output(g)
	for _, want := range []string{"mail", "servidor de correo", "dev", "ICE", "SYNC"} {
		if !strings.Contains(out, want) {
			t.Fatalf("scan missing %q:\n%s", want, out)
		}
	}
}

func TestConnectBreachesThenOwnsTheNode(t *testing.T) {
	g := newTestGame(Won)
	g.Step(Input{Text: []rune("connect mail")})
	g.Step(Input{Enter: true})
	if g.Mode() != ModeBreach || g.TakingText() {
		t.Fatalf("connecting to a guarded node must start its breach, mode %v", g.Mode())
	}
	g.Step(Input{})
	if g.Mode() != ModeTerminal || g.Node() != "mail" || !g.Owned("mail") {
		t.Fatalf("a won breach must own the node and move there: mode %v node %q", g.Mode(), g.Node())
	}
	run(g, "connect gate")
	if g.Node() != "gate" || g.Mode() != ModeTerminal {
		t.Fatalf("an owned node needs no breach, at %q", g.Node())
	}
	run(g, "connect core")
	if g.Node() != "gate" || g.Trace() != TraceFailed {
		t.Fatalf("a node no owned node links to must be out of reach, at %q trace %d", g.Node(), g.Trace())
	}
}

func TestALostBreachRaisesTheTraceAndKeepsYouOut(t *testing.T) {
	g := newTestGame(Lost)
	run(g, "connect dev")
	if g.Node() != "gate" || g.Owned("dev") || g.Trace() != difficulties[0].BreachTrace || g.Mode() != ModeTerminal {
		t.Fatalf("node %q owned %v trace %d mode %v", g.Node(), g.Owned("dev"), g.Trace(), g.Mode())
	}
}

func TestAPasswordGuardsTheVaultAndIsMasked(t *testing.T) {
	g := newTestGame(Won)
	run(g, "connect mail")
	run(g, "connect vault")
	if !strings.Contains(g.Prompt(), "contraseña") {
		t.Fatalf("prompt %q, want a password prompt", g.Prompt())
	}
	run(g, "kitty")
	if g.Node() != "mail" || g.Trace() != TracePassword || strings.Contains(g.Text(), "kitty") {
		t.Fatalf("a wrong password: node %q trace %d, and it must not show:\n%s", g.Node(), g.Trace(), g.Text())
	}
	run(g, "connect vault")
	run(g, "pixel")
	if g.Node() != "vault" || !g.Owned("vault") || strings.Contains(g.Text(), "pixel") {
		t.Fatalf("the right password must open the vault: node %q", g.Node())
	}
}

func TestDecryptRevealsAFileWithTheRightKey(t *testing.T) {
	g := newTestGame(Won)
	run(g, "connect mail")
	run(g, "connect vault")
	run(g, "pixel")
	run(g, "cat core.key.enc")
	if strings.Contains(output(g), "LANTERN-7") {
		t.Fatalf("an encrypted file must not show its contents")
	}
	run(g, "decrypt core.key.enc spark")
	if g.Trace() != TracePassword {
		t.Fatalf("a wrong key must raise the trace by %d, got %d", TracePassword, g.Trace())
	}
	run(g, "decrypt core.key.enc ember")
	if !strings.Contains(output(g), "LANTERN-7") {
		t.Fatalf("the right key must show the file:\n%s", output(g))
	}
	run(g, "cat core.key.enc")
	if !strings.Contains(output(g), "LANTERN-7") {
		t.Fatalf("a decrypted file stays readable")
	}
}

func TestTheCoreNeedsTheKeyThenReadingLanternWins(t *testing.T) {
	g := newTestGame(Won)
	run(g, "connect mail")
	run(g, "connect vault")
	run(g, "pixel")
	run(g, "connect core")
	if g.Node() != "vault" || g.Mode() != ModeTerminal || g.Trace() != TraceFailed {
		t.Fatalf("the core must refuse without the key: node %q trace %d", g.Node(), g.Trace())
	}
	run(g, "decrypt core.key.enc ember")
	run(g, "connect core")
	if g.Node() != "core" {
		t.Fatalf("with the key the core's breach must open it, at %q", g.Node())
	}
	run(g, "cat lantern.txt")
	if g.Over() || !strings.Contains(output(g), "NIVEL 1 SUPERADO") {
		t.Fatalf("reading the final file must clear the level: over %v\n%s", g.Over(), output(g))
	}
}

func TestAFullTraceEndsTheRound(t *testing.T) {
	g := newTestGame(Lost)
	for i := 0; i < 10 && !g.Over(); i++ {
		run(g, "connect dev")
	}
	if !g.Over() || g.Won() || g.Trace() != 100 {
		t.Fatalf("over %v won %v trace %d, want caught at 100", g.Over(), g.Won(), g.Trace())
	}
	before := len(g.Lines())
	run(g, "ls")
	if len(g.Lines()) != before {
		t.Fatalf("a finished round must not advance")
	}
}

func TestTheMapCommandOpensTheMap(t *testing.T) {
	g := newTestGame(Won)
	run(g, "map")
	if g.Mode() != ModeMap || g.TakingText() || g.Selected() != "gate" {
		t.Fatalf("mode %v selected %q", g.Mode(), g.Selected())
	}
}

func TestElapsedTimeIsCounted(t *testing.T) {
	g := newTestGame(Won)
	for i := 0; i < 60; i++ {
		g.Step(Input{})
	}
	if g.Elapsed().Round(time.Millisecond) != time.Second {
		t.Fatalf("elapsed %v after 60 ticks, want 1s", g.Elapsed())
	}
}

func TestTheMapMovesBetweenNodesAndConnects(t *testing.T) {
	g := newTestGame(Won)
	run(g, "map")
	g.Step(Input{Move: right})
	if g.Selected() != "mail" {
		t.Fatalf("right from the gateway selects %q, want mail", g.Selected())
	}
	g.Step(Input{Move: down})
	if g.Selected() != "dev" {
		t.Fatalf("down from mail selects %q, want dev", g.Selected())
	}
	g.Step(Input{Enter: true})
	for i := 0; i < 3 && g.Mode() == ModeBreach; i++ {
		g.Step(Input{})
	}
	if g.Node() != "dev" || g.Mode() != ModeTerminal {
		t.Fatalf("enter on the map must connect: node %q mode %v", g.Node(), g.Mode())
	}
	run(g, "map")
	g.Step(Input{Backspace: true})
	if g.Mode() != ModeTerminal {
		t.Fatalf("backspace must leave the map")
	}
}

func TestAMachineOutOfReachAsksNoPassword(t *testing.T) {
	g := newTestGame(Won)
	run(g, "connect vault")
	if strings.Contains(g.Prompt(), "contraseña") || g.Trace() != TraceFailed || !strings.Contains(output(g), "sin ruta") {
		t.Fatalf("prompt %q trace %d output %q, want no route from the gateway", g.Prompt(), g.Trace(), output(g))
	}
}

func TestHintGivesTheNextStepAndCostsPoints(t *testing.T) {
	g := newTestGame(Won)
	run(g, "hint")
	first := output(g)
	if !strings.Contains(first, "scan") || g.Hints() != 1 || g.Score() != ScoreStart-HintCost {
		t.Fatalf("hint %q hints %d score %d", first, g.Hints(), g.Score())
	}
	run(g, "hint")
	second := output(g)
	if second == first || !strings.Contains(second, "connect") || g.Hints() != 2 {
		t.Fatalf("a second hint must be more explicit: %q, hints %d", second, g.Hints())
	}
	run(g, "hint")
	if !strings.Contains(output(g), "connect") || g.Hints() != 2 || g.Trace() != 0 {
		t.Fatalf("with no hint left, hint repeats the last one for free: hints %d trace %d", g.Hints(), g.Trace())
	}
}

func TestHintsFollowProgress(t *testing.T) {
	g := newTestGame(Won)
	explicit := func() string {
		run(g, "hint")
		run(g, "hint")
		return output(g)
	}
	run(g, "connect mail")
	if out := explicit(); !strings.Contains(out, "pixel") {
		t.Fatalf("after mail the hint is about the vault password: %q", out)
	}
	run(g, "connect vault")
	run(g, "pixel")
	if out := explicit(); !strings.Contains(out, "decrypt core.key.enc ember") {
		t.Fatalf("in the vault the hint is about the key: %q", out)
	}
	run(g, "decrypt core.key.enc ember")
	if out := explicit(); !strings.Contains(out, "connect core") {
		t.Fatalf("with the key the hint points to the core: %q", out)
	}
	run(g, "connect core")
	if out := explicit(); !strings.Contains(out, "cat lantern.txt") {
		t.Fatalf("on the core the hint points to the secret: %q", out)
	}
	if g.Hints() != 8 {
		t.Fatalf("hints %d, want 8", g.Hints())
	}
}

func TestScoreFallsWithTraceAndHints(t *testing.T) {
	g := newTestGame(Won)
	run(g, "nope")
	run(g, "hint")
	if want := ScoreStart - TraceCost*TraceFailed - HintCost; g.Score() != want {
		t.Fatalf("score %d, want %d", g.Score(), want)
	}
	g.trace, g.hints = 100, 10
	if g.Score() != 0 {
		t.Fatalf("score %d, want it never below 0", g.Score())
	}
}
