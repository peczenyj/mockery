package template

// Comment represents a single line of comments exactly as it appears in source.
// This includes the "//" or "/*" strings.
type Comment string

type CommentGroup struct {
	List []Comment
	Text string
}
