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
	// Requires is the machine, a firewall, that must be yours before this
	// one opens.
	Requires string
	// Honeypot marks a trap: it looks open, and connecting raises the
	// trace by TraceHoneypot.
	Honeypot bool
	// Watched marks a machine the level's IDS watches: standing on it
	// raises the trace until the IDS is yours.
	Watched bool
	// Lockout is how many wrong passwords lock the machine for good; then
	// only a LockBreach opens it. 0 never locks.
	Lockout    int
	LockBreach BreachKind
	Files      []File
}

// Hint is a vague and an explicit hint for a stage of a level, the stage
// reached once When holds (always, when nil).
type Hint struct {
	When            func(g *Game) bool
	Vague, Explicit string
}

// Level is a network to break into, with its story and hints. Its final
// file, once read, finishes it.
type Level struct {
	Name string
	// IDS is the machine that, once yours, stops the watching.
	IDS   string
	Nodes []Node
	Hints []Hint
}

func owns(ids ...string) func(*Game) bool {
	return func(g *Game) bool {
		for _, id := range ids {
			if g.Owned(id) {
				return true
			}
		}
		return false
	}
}

func decrypted(node, file string) func(*Game) bool {
	return func(g *Game) bool { return g.decrypted[node+"/"+file] }
}

// levels are the three targets in order, each harder in its breaches and
// its logic. The story is in Spanish (L8); commands, machine names and
// common file names stay in English.
var levels = [...]Level{
	{
		// Level 1, the tutorial: Meridian Grid, a fictional power utility,
		// and what its Project Lantern hides.
		Name: "MERIDIAN GRID // pasarela de contrata",
		Hints: []Hint{
			{Vague: "Lee readme.txt y usa `scan` para ver a qué máquinas llegas.",
				Explicit: "Escribe `connect mail` o `connect dev`: ganar su brecha te da la máquina."},
			{When: owns("mail", "dev"),
				Vague:    "La contraseña del vault está en los correos de mail: es el nombre de la mascota.",
				Explicit: "La mascota se llama Pixel: `connect vault` y escribe pixel."},
			{When: owns("vault"),
				Vague:    "core.key.enc está cifrado; la clave sale en las notas de dev.",
				Explicit: "Desde el vault: `decrypt core.key.enc ember`."},
			{When: decrypted("vault", "core.key.enc"),
				Vague:    "Con la clave de admin descifrada, el servidor central ya te deja intentarlo.",
				Explicit: "`connect core` y supera la brecha de puertas lógicas."},
			{When: owns("core"),
				Vague:    "El secreto está en un fichero del servidor central.",
				Explicit: "`cat lantern.txt`"},
		},
		Nodes: []Node{
			{
				ID: "gate", Title: "pasarela de contrata", Pos: grid.Point{X: 4, Y: 8},
				Links: []string{"mail", "dev"},
				Files: []File{
					{Name: "readme.txt", Body: []string{
						"ENCARGO 0x2F // cliente anónimo, pagado por adelantado",
						"Objetivo: Meridian Grid, la compañía eléctrica de la ciudad.",
						"Averigua qué es el 'Proyecto Lantern'. Léelo en el servidor central.",
						"Empieza con `scan` para ver a qué máquinas llega esta.",
					}},
					{Name: "notas.txt", Body: []string{
						"- el servidor de correo lo filtra todo; nadie borra nada",
						"- la estación de dev la vigilan centinelas ICE: muévete rápido",
						"- cada comando fallido sube el rastreo. al 100% te pillan.",
						"- atascado? `hint` da una pista, pero resta puntos.",
					}},
				},
			},
			{
				ID: "mail", Title: "servidor de correo", Pos: grid.Point{X: 24, Y: 3},
				Links: []string{"gate", "vault"}, Breach: BreachSync,
				Files: []File{
					{Name: "correo_0412.txt", Body: []string{
						"De: CTO   Para: sistemas",
						"He cambiado la contraseña del vault. Es el nombre de la mascota,",
						"todo en minúsculas. Esta vez NO la apuntéis en ningún sitio.",
					}},
					{Name: "boletin.txt", Body: []string{
						"BOLETÍN MERIDIAN",
						"¡Os presentamos a Pixel, la gata de la oficina y nueva mascota oficial!",
						"Pixel ya ha desenchufado dos servidores. Queremos mucho a Pixel.",
					}},
				},
			},
			{
				ID: "dev", Title: "estación de desarrollo", Pos: grid.Point{X: 24, Y: 13},
				Links: []string{"gate", "vault"}, Breach: BreachICE,
				Files: []File{
					{Name: "todo.txt", Body: []string{
						"[x] desplegar la build 7 de Lantern en el core",
						"[ ] dejar de cifrar las copias del vault con 'ember', sale en todos los scripts",
					}},
					{Name: "lantern.log", Body: []string{
						"lantern: build 7 desplegada en el core",
						"lantern: relés de distrito respondiendo: 41 de 41",
						"lantern: clave de admin del core movida al vault (cifrada)",
					}},
				},
			},
			{
				ID: "vault", Title: "almacén de copias", Pos: grid.Point{X: 42, Y: 8},
				Links: []string{"mail", "dev", "core"}, Password: "pixel",
				Files: []File{
					{Name: "core.key.enc", Encrypted: true, Key: "ember", Body: []string{
						"CLAVE DE ADMIN DEL CORE: LANTERN-7-OVERRIDE",
						"El servidor central acepta esta clave. Usa `connect core`.",
					}},
					{Name: "backup.log", Body: []string{
						"copia nocturna ok // cifrado: el estándar de la casa (ver todo de dev)",
					}},
				},
			},
			{
				ID: "core", Title: "servidor central", Pos: grid.Point{X: 56, Y: 8},
				Links: []string{"vault"}, Breach: BreachGates, NeedsKey: "core.key.enc",
				Files: []File{
					{Name: "lantern.txt", Final: true, Body: []string{
						"PROYECTO LANTERN",
						"Un interruptor en los 41 relés de distrito. Una orden desde este servidor",
						"deja a oscuras cualquier distrito, y los contadores siguen facturando.",
						"Tu cliente era la defensora del pueblo. El fichero ya va de camino.",
						"Las luces seguirán encendidas. Buen trabajo, ghost.",
					}},
				},
			},
		},
	},
	{
		// Level 2: Helix Biomedical and the trial it buried. Adds machines
		// behind a firewall and a password built from two clues.
		Name: "HELIX BIOMEDICAL // pasarela de proveedor",
		Hints: []Hint{
			{Vague: "Empieza con `scan`: web y rrhh son tus entradas.",
				Explicit: "`connect web` o `connect rrhh`, y supera su brecha."},
			{When: owns("web", "rrhh"),
				Vague:    "El lab está detrás del firewall: hazte con fw antes.",
				Explicit: "`connect fw` y supera su brecha de puertas."},
			{When: owns("fw"),
				Vague:    "La contraseña del lab sigue la política de rrhh: lee sus ficheros.",
				Explicit: "Mascota y año de Marta: `connect lab` y escribe nimbus1987."},
			{When: owns("lab"),
				Vague:    "Los resultados reales están cifrados en datos; Marta explica la clave.",
				Explicit: "`connect datos` y después `decrypt hx12_real.enc 21xh`."},
			{When: owns("datos"),
				Vague:    "Descifra los resultados reales del ensayo.",
				Explicit: "`decrypt hx12_real.enc 21xh`"},
		},
		Nodes: []Node{
			{
				ID: "gate", Title: "pasarela de proveedor", Pos: grid.Point{X: 2, Y: 8},
				Links: []string{"web", "rrhh"},
				Files: []File{
					{Name: "readme.txt", Body: []string{
						"ENCARGO 0x31 // una periodista de sanidad",
						"Objetivo: Helix Biomedical, farmacéutica.",
						"Dicen que el ensayo HX-12 salió limpio. Ella cree que no.",
						"Trae los resultados reales. Están en algún sitio tras su firewall.",
					}},
					{Name: "notas.txt", Body: []string{
						"- esta vez las brechas son más duras y el rastreo no perdona",
						"- algunas máquinas no abren hasta que tomas su firewall",
					}},
				},
			},
			{
				ID: "web", Title: "servidor web", Pos: grid.Point{X: 18, Y: 3},
				Links: []string{"gate", "fw", "lab"}, Breach: BreachICE,
				Files: []File{
					{Name: "robots.txt", Body: []string{
						"User-agent: *",
						"Disallow: /ensayos/hx12   # nadie de fuera debe ver esto",
					}},
					{Name: "changelog.txt", Body: []string{
						"v4.2: el lab ya solo acepta conexiones a través de fw",
					}},
				},
			},
			{
				ID: "rrhh", Title: "recursos humanos", Pos: grid.Point{X: 18, Y: 13},
				Links: []string{"gate", "fw", "lab"}, Breach: BreachSync,
				Files: []File{
					{Name: "politica.txt", Body: []string{
						"POLÍTICA DE CONTRASEÑAS (sí, otra vez)",
						"La del personal: nombre de tu mascota y tu año de nacimiento,",
						"todo junto y en minúsculas. Fácil de recordar, ¿verdad?",
					}},
					{Name: "ficha_marta.txt", Body: []string{
						"Marta Ruiz, jefa de ensayos clínicos. Acceso: lab.",
						"Nacida en 1987. Contacto de emergencia: su gato Nimbus (no contesta).",
					}},
				},
			},
			{
				ID: "fw", Title: "firewall perimetral", Pos: grid.Point{X: 34, Y: 8},
				Links: []string{"web", "rrhh", "lab"}, Breach: BreachGates,
				Files: []File{
					{Name: "reglas.log", Body: []string{
						"allow lab   <- fw   (solo con fw comprometido... digo, autorizado)",
						"allow datos <- lab",
					}},
				},
			},
			{
				ID: "lab", Title: "laboratorio de ensayos", Pos: grid.Point{X: 50, Y: 3},
				Links: []string{"web", "rrhh", "fw", "datos"}, Password: "nimbus1987", Requires: "fw",
				Files: []File{
					{Name: "correo_marta.txt", Body: []string{
						"De: Marta   Para: dirección",
						"He guardado los resultados reales de HX-12 en datos, cifrados.",
						"La clave es el código del ensayo escrito al revés, en minúsculas.",
						"Los publicados son los 'ajustados'. Yo no pienso firmarlos.",
					}},
					{Name: "hx12_publico.txt", Body: []string{
						"HX-12: efectos adversos graves: 0. Eficacia: 94%.",
					}},
				},
			},
			{
				ID: "datos", Title: "almacén de datos", Pos: grid.Point{X: 54, Y: 13},
				Links: []string{"lab"}, Breach: BreachICE,
				Files: []File{
					{Name: "hx12_real.enc", Encrypted: true, Key: "21xh", Final: true, Body: []string{
						"HX-12 // RESULTADOS SIN AJUSTAR",
						"Efectos adversos graves: 31 de 400 pacientes. Eficacia: 52%.",
						"Decisión de dirección: retirar los 31 casos 'por errores de registro'.",
						"El informe ya está en la bandeja de la periodista.",
					}},
					{Name: "backup.log", Body: []string{
						"copia ok // cifrado: clave definida por la jefa de ensayos",
					}},
				},
			},
		},
	},
	{
		// Level 3: Kestrel Analytics and its watchlist. Adds a honeypot and
		// an admin key split in two fragments across the network.
		Name: "KESTREL ANALYTICS // pasarela de mantenimiento",
		Hints: []Hint{
			{Vague: "Usa `scan` y desconfía de lo que parezca demasiado fácil.",
				Explicit: "`connect vpn` o `connect cam`. No entres en admin: es un honeypot."},
			{When: owns("vpn", "cam"),
				Vague:    "La clave del núcleo está partida en dos fragmentos: uno en nas, otro en backup.",
				Explicit: "`connect nas` y lee fragmento_1.txt y rotacion.txt."},
			{When: owns("nas"),
				Vague:    "backup pide contraseña: la pista está en los registros de cam.",
				Explicit: "El nombre de la empresa al revés: `connect backup` y escribe lertsek."},
			{When: owns("backup"),
				Vague:    "Junta los dos fragmentos y descifra la clave en fw.",
				Explicit: "`connect fw` y después `decrypt nucleo.key.enc orion42`."},
			{When: decrypted("fw", "nucleo.key.enc"),
				Vague:    "Con la clave descifrada, el núcleo ya te deja intentarlo.",
				Explicit: "`connect core` y supera su brecha de puertas."},
			{When: owns("core"),
				Vague:    "La lista está en un fichero del núcleo.",
				Explicit: "`cat kestrel.txt`"},
		},
		Nodes: []Node{
			{
				ID: "gate", Title: "pasarela de mantenimiento", Pos: grid.Point{X: 2, Y: 8},
				Links: []string{"vpn", "cam"},
				Files: []File{
					{Name: "readme.txt", Body: []string{
						"ENCARGO 0x40 // sin nombre, solo una dirección de entrega",
						"Objetivo: Kestrel Analytics, 'consultora de riesgos'.",
						"Venden a quien pague listas de gente que vigilar.",
						"Saca su lista del núcleo. Es el último trabajo, ghost.",
					}},
					{Name: "notas.txt", Body: []string{
						"- Kestrel pone trampas: no todo lo que está abierto es una puerta",
						"- la clave del núcleo la rotan y la parten en trozos",
					}},
				},
			},
			{
				ID: "vpn", Title: "concentrador vpn", Pos: grid.Point{X: 16, Y: 3},
				Links: []string{"gate", "admin", "nas"}, Breach: BreachSync,
				Files: []File{
					{Name: "aviso_seguridad.txt", Body: []string{
						"A todo el equipo: 'admin' es un honeypot de seguridad.",
						"Quien se conecte queda marcado. Para trabajar, usad nas.",
					}},
				},
			},
			{
				ID: "admin", Title: "panel de administración", Pos: grid.Point{X: 30, Y: 1},
				Links: []string{"vpn"}, Honeypot: true,
			},
			{
				ID: "cam", Title: "red de cámaras", Pos: grid.Point{X: 16, Y: 13},
				Links: []string{"gate", "nas", "backup"}, Breach: BreachICE,
				Files: []File{
					{Name: "cam07.log", Body: []string{
						"cam07: objetivo 118 entra en la asociación de vecinos, 19:02",
						"cam07: objetivo 118 sale con dos personas no fichadas, 20:47",
					}},
					{Name: "mantenimiento.txt", Body: []string{
						"La contraseña de backup es el nombre de la empresa al revés.",
						"Nadie lo adivinará nunca. (firmado: el de sistemas)",
					}},
				},
			},
			{
				ID: "nas", Title: "almacén compartido", Pos: grid.Point{X: 32, Y: 8},
				Links: []string{"vpn", "cam", "backup", "fw"}, Breach: BreachGates,
				Files: []File{
					{Name: "fragmento_1.txt", Body: []string{
						"clave del núcleo, fragmento 1 de 2: orion",
					}},
					{Name: "rotacion.txt", Body: []string{
						"Rotación de la clave del núcleo: fragmento 1 seguido del 2, sin espacios.",
						"La anterior (orion41) está revocada: usarla dispara una alerta.",
					}},
				},
			},
			{
				ID: "backup", Title: "copias de seguridad", Pos: grid.Point{X: 32, Y: 14},
				Links: []string{"cam", "nas"}, Password: "lertsek", Requires: "nas",
				Files: []File{
					{Name: "fragmento_2.txt", Body: []string{
						"clave del núcleo, fragmento 2 de 2: 42",
					}},
				},
			},
			{
				ID: "fw", Title: "firewall interno", Pos: grid.Point{X: 44, Y: 8},
				Links: []string{"nas", "core"}, Breach: BreachICE,
				Files: []File{
					{Name: "nucleo.key.enc", Encrypted: true, Key: "orion42", Body: []string{
						"CLAVE DE ADMIN DEL NÚCLEO: KESTREL-ROOT-9",
						"El núcleo acepta esta clave. Usa `connect core`.",
					}},
				},
			},
			{
				ID: "core", Title: "núcleo", Pos: grid.Point{X: 56, Y: 8},
				Links: []string{"fw"}, Breach: BreachGates, NeedsKey: "nucleo.key.enc",
				Files: []File{
					{Name: "kestrel.txt", Final: true, Body: []string{
						"KESTREL // LISTA DE SEGUIMIENTO",
						"1.204 personas: periodistas, vecinos, sindicalistas, una jueza.",
						"Clientes: tres empresas y un ministerio que dice no saber nada.",
						"La dirección de entrega era la de un juzgado. Ya la tienen.",
						"Se acabó. Buen trabajo, ghost.",
					}},
				},
			},
		},
	},
	{
		// Level 4: Cormoran Shipping and the scrap that is not scrap. Adds
		// active watching: an IDS raises the trace on watched machines.
		Name: "CORMORAN SHIPPING // pasarela del puerto",
		IDS:  "ids",
		Hints: []Hint{
			{Vague: "Hay un IDS: en las máquinas vigiladas el rastreo sube solo. Lee notas.txt.",
				Explicit: "`connect puerto` o `connect correo`, y después ve a por ids cuanto antes."},
			{When: owns("puerto", "correo"),
				Vague:    "Toma ids: mientras no sea tuya, cada segundo en una máquina vigilada cuenta.",
				Explicit: "`connect ids` y supera su brecha de puertas."},
			{When: owns("ids"),
				Vague:    "La contraseña de aduanas sale de la norma del correo del jefe y del registro de las grúas.",
				Explicit: "Barco y muelle: `connect aduanas` y escribe albatros7."},
			{When: owns("aduanas"),
				Vague:    "El manifiesto real está cifrado en flota; aduanas dice con qué.",
				Explicit: "`connect flota` y después `decrypt manifiesto_real.enc 4471`."},
			{When: owns("flota"),
				Vague:    "Descifra el manifiesto real.",
				Explicit: "`decrypt manifiesto_real.enc 4471`"},
		},
		Nodes: []Node{
			{
				ID: "gate", Title: "pasarela del puerto", Pos: grid.Point{X: 2, Y: 8},
				Links: []string{"puerto", "correo"},
				Files: []File{
					{Name: "readme.txt", Body: []string{
						"ENCARGO 0x4C // un estibador del sindicato",
						"Objetivo: Cormoran Shipping, naviera.",
						"El 'Albatros' descarga de noche contenedores de 'chatarra' que nadie toca.",
						"Trae el manifiesto real. Ellos tienen otro para la prensa.",
					}},
					{Name: "notas.txt", Body: []string{
						"- tienen un IDS: en una máquina vigilada el rastreo sube solo, segundo a segundo",
						"- si la máquina ids es tuya, deja de vigilar. ve a por ella pronto",
					}},
				},
			},
			{
				ID: "puerto", Title: "grúas del puerto", Pos: grid.Point{X: 16, Y: 3},
				Links: []string{"gate", "ids", "aduanas"}, Breach: BreachICE, Watched: true,
				Files: []File{
					{Name: "grua12.log", Body: []string{
						"03:10 atraca el Albatros en el muelle 7, luces apagadas",
						"03:25 descarga de 40 contenedores, sin inspección",
					}},
				},
			},
			{
				ID: "correo", Title: "correo corporativo", Pos: grid.Point{X: 16, Y: 13},
				Links: []string{"gate", "ids", "aduanas"}, Breach: BreachSync, Watched: true,
				Files: []File{
					{Name: "correo_jefe.txt", Body: []string{
						"De: jefe de operaciones   Para: aduanas",
						"Nueva contraseña de aduanas: el último barco que atracó de noche",
						"y su muelle, todo junto y en minúsculas. Borrad este correo.",
					}},
				},
			},
			{
				ID: "ids", Title: "detector de intrusiones", Pos: grid.Point{X: 30, Y: 8},
				Links: []string{"puerto", "correo", "aduanas", "flota"}, Breach: BreachGates,
				Files: []File{
					{Name: "ids.conf", Body: []string{
						"vigila: puerto, correo, aduanas, flota",
						"alerta: +1% de rastreo por segundo de sesión no autorizada",
					}},
				},
			},
			{
				ID: "aduanas", Title: "sistema de aduanas", Pos: grid.Point{X: 44, Y: 3},
				Links: []string{"puerto", "correo", "ids", "flota"}, Password: "albatros7", Watched: true,
				Files: []File{
					{Name: "manifiesto.txt", Body: []string{
						"Albatros: 40 contenedores de chatarra. Contenedor de control: CTN-4471.",
						"El manifiesto real va cifrado con el número del contenedor de control,",
						"solo las cifras.",
					}},
				},
			},
			{
				ID: "flota", Title: "gestión de flota", Pos: grid.Point{X: 44, Y: 13},
				Links: []string{"aduanas", "ids"}, Breach: BreachICE, Watched: true, Requires: "aduanas",
				Files: []File{
					{Name: "manifiesto_real.enc", Encrypted: true, Key: "4471", Final: true, Body: []string{
						"ALBATROS // MANIFIESTO REAL",
						"40 contenedores de residuos industriales con metales pesados.",
						"Destino declarado: chatarrería. Destino real: vertedero sin licencia.",
						"Propietario último de Cormoran: Halcyon Group. Otra vez ese nombre.",
					}},
				},
			},
		},
	},
	{
		// Level 5: Halcyon Group, the holding behind every target. Adds a
		// lockout: two wrong passwords lock the key machine for good.
		Name: "HALCYON GROUP // pasarela de la sede",
		IDS:  "auditoria",
		Hints: []Hint{
			{Vague: "Halcyon vigila y bloquea cuentas: lee notas.txt antes de nada.",
				Explicit: "`connect prensa` o `connect legal`. No toques soporte: es un honeypot."},
			{When: owns("prensa", "legal"),
				Vague:    "Toma auditoria para que dejen de vigilarte, y mira el archivo.",
				Explicit: "`connect auditoria`, luego `connect archivo` y lee candidatas.txt."},
			{When: owns("archivo"),
				Vague:    "Descarta candidatas con lo que dicen prensa y legal. Solo tienes dos intentos.",
				Explicit: "Ni el mar ni el nombre de la empresa: `connect presidencia` y escribe aurora1959."},
			{When: owns("presidencia"),
				Vague:    "La nota de la presidenta explica la clave del núcleo.",
				Explicit: "`connect core`, supera su ICE y después `decrypt halcyon.enc lhk`."},
			{When: owns("core"),
				Vague:    "Descifra el informe final.",
				Explicit: "`decrypt halcyon.enc lhk`"},
		},
		Nodes: []Node{
			{
				ID: "gate", Title: "pasarela de la sede", Pos: grid.Point{X: 2, Y: 8},
				Links: []string{"prensa", "legal", "soporte"},
				Files: []File{
					{Name: "readme.txt", Body: []string{
						"ENCARGO 0x50 // la misma dirección de entrega que la última vez",
						"Meridian, Helix, Kestrel, Cormoran: todas son de Halcyon Group.",
						"Entra en su núcleo y saca lo que las une. Es el final, ghost.",
					}},
					{Name: "notas.txt", Body: []string{
						"- su IDS se llama auditoria: hasta que sea tuyo, el rastreo sube solo",
						"- presidencia bloquea la cuenta al segundo fallo: piensa antes de escribir",
						"- soporte parece abierto. Demasiado abierto.",
					}},
				},
			},
			{
				ID: "soporte", Title: "soporte técnico", Pos: grid.Point{X: 30, Y: 1},
				Links: []string{"gate", "prensa"}, Honeypot: true,
			},
			{
				ID: "prensa", Title: "gabinete de prensa", Pos: grid.Point{X: 16, Y: 3},
				Links: []string{"gate", "soporte", "archivo"}, Breach: BreachSync, Watched: true,
				Files: []File{
					{Name: "entrevista.txt", Body: []string{
						"Elena Varo, presidenta: 'Le tengo pánico al mar.",
						"No lo quiero cerca ni en un barco ni en una contraseña.'",
					}},
				},
			},
			{
				ID: "legal", Title: "asesoría legal", Pos: grid.Point{X: 16, Y: 13},
				Links: []string{"gate", "auditoria", "archivo"}, Breach: BreachICE, Watched: true,
				Files: []File{
					{Name: "circular.txt", Body: []string{
						"Desde 2021 queda prohibido usar el nombre de la empresa en las contraseñas.",
						"Afecta también a la presidencia. Sobre todo a la presidencia.",
					}},
				},
			},
			{
				ID: "auditoria", Title: "auditoría interna", Pos: grid.Point{X: 30, Y: 13},
				Links: []string{"legal", "archivo"}, Breach: BreachGates,
				Files: []File{
					{Name: "auditoria.conf", Body: []string{
						"vigila: prensa, legal, archivo, presidencia",
						"bloqueo de cuentas: 2 fallos",
					}},
				},
			},
			{
				ID: "archivo", Title: "archivo de dirección", Pos: grid.Point{X: 32, Y: 7},
				Links: []string{"prensa", "legal", "auditoria", "presidencia"}, Breach: BreachGates, Watched: true,
				Files: []File{
					{Name: "candidatas.txt", Body: []string{
						"Contraseñas que ha usado la presidenta en los últimos años:",
						"marea88, halcyon2020, aurora1959. Solo una sigue activa.",
					}},
				},
			},
			{
				ID: "presidencia", Title: "despacho de presidencia", Pos: grid.Point{X: 50, Y: 3},
				Links: []string{"archivo", "core"}, Password: "aurora1959", Lockout: 2, LockBreach: BreachICE, Watched: true,
				Files: []File{
					{Name: "nota_presidenta.txt", Body: []string{
						"La clave del núcleo: las iniciales de los tres proyectos que nadie",
						"debía ver, en el orden en que los lanzamos y en minúsculas.",
						"Lantern, el ensayo HX-12 y Kestrel. No la apuntéis.",
					}},
				},
			},
			{
				ID: "core", Title: "núcleo de Halcyon", Pos: grid.Point{X: 50, Y: 12},
				Links: []string{"presidencia"}, Breach: BreachICE,
				Files: []File{
					{Name: "halcyon.enc", Encrypted: true, Key: "lhk", Final: true, Body: []string{
						"HALCYON GROUP // CONSEJO, ACTA RESERVADA",
						"Un mismo consejo decidió apagar barrios, esconder un ensayo,",
						"vigilar a quien protestaba y enterrar residuos en el puerto.",
						"El acta lleva cinco firmas. Ya va camino de un juzgado y de tres redacciones.",
						"Se acabó de verdad. Buen trabajo, ghost.",
					}},
				},
			},
		},
	},
}
