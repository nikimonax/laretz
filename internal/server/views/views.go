package views

// Error is a user-facing failure embedded into a page render.
// The zero value renders nothing.
type Error struct {
	Dialog  string // id of the dialog to keep open; empty renders a page-level alert
	Message string
}
