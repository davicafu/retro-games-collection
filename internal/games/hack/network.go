package hack

import "github.com/davidbarreiro/cli-games-collection/internal/engine/grid"

// BreachKind is the arcade challenge that guards a node.
type BreachKind int

const (
	NoBreach BreachKind = iota
	BreachICE
	BreachSync
	BreachGates
)

// File is a file on a node. An encrypted file shows its Body only after
// decrypt with Key.
type File struct {
	Name      string
	Body      []string
	Encrypted bool
	Key       string
	// Final marks the file whose reading wins the game.
	Final bool
}

// Node is a machine on the target network. Pos is where the map draws it.
type Node struct {
	ID       string
	Title    string
	Pos      grid.Point
	Links    []string
	Breach   BreachKind
	Password string
	// NeedsKey is the decrypted file the node asks for before a breach.
	NeedsKey string
	Files    []File
}

// network is the target: Meridian Grid, a fictional power utility, and
// what its Project Lantern hides.
var network = []Node{
	{
		ID: "gate", Title: "contractor gateway", Pos: grid.Point{X: 4, Y: 8},
		Links: []string{"mail", "dev"},
		Files: []File{
			{Name: "readme.txt", Body: []string{
				"JOB 0x2F // anonymous client, paid in advance",
				"Target: Meridian Grid, the city's power utility.",
				"Find out what 'Project Lantern' is. Read it from the core server.",
				"Start with `scan` to see the machines this one reaches.",
			}},
			{Name: "notes.txt", Body: []string{
				"- the mail relay leaks everything; staff never delete anything",
				"- the dev box is watched by ICE sentries: move fast",
				"- every failed command raises the trace. at 100% they have you.",
			}},
		},
	},
	{
		ID: "mail", Title: "mail relay", Pos: grid.Point{X: 24, Y: 3},
		Links: []string{"gate", "vault"}, Breach: BreachSync,
		Files: []File{
			{Name: "inbox_0412.txt", Body: []string{
				"From: CTO   To: ops",
				"Rotated the vault password. It is our mascot's name,",
				"all lowercase. Do NOT write it down anywhere this time.",
			}},
			{Name: "newsletter.txt", Body: []string{
				"MERIDIAN WEEKLY",
				"Say hello to Pixel, the office cat and our new official mascot!",
				"Pixel has already unplugged two servers. We love Pixel.",
			}},
		},
	},
	{
		ID: "dev", Title: "dev workstation", Pos: grid.Point{X: 24, Y: 13},
		Links: []string{"gate", "vault"}, Breach: BreachICE,
		Files: []File{
			{Name: "todo.txt", Body: []string{
				"[x] ship Lantern build 7 to core",
				"[ ] stop encrypting vault backups with 'ember', it is in every script",
			}},
			{Name: "lantern.log", Body: []string{
				"lantern: build 7 deployed to core",
				"lantern: district relays answering: 41 of 41",
				"lantern: core admin key moved to the vault (encrypted)",
			}},
		},
	},
	{
		ID: "vault", Title: "backup vault", Pos: grid.Point{X: 44, Y: 8},
		Links: []string{"mail", "dev", "core"}, Password: "pixel",
		Files: []File{
			{Name: "core.key.enc", Encrypted: true, Key: "ember", Body: []string{
				"CORE ADMIN KEY: LANTERN-7-OVERRIDE",
				"The core server accepts this key. Use `connect core`.",
			}},
			{Name: "backup.log", Body: []string{
				"nightly backup ok // cipher: house standard (see dev todo)",
			}},
		},
	},
	{
		ID: "core", Title: "core server", Pos: grid.Point{X: 58, Y: 8},
		Links: []string{"vault"}, Breach: BreachGates, NeedsKey: "core.key.enc",
		Files: []File{
			{Name: "lantern.txt", Final: true, Body: []string{
				"PROJECT LANTERN",
				"A switch in all 41 district relays. One command from this server",
				"turns any district dark, and the meters keep billing.",
				"Your client was the city ombudsman. The file is already on its way.",
				"Lights stay on. Nice work, ghost.",
			}},
		},
	},
}
