package o11y

type Reporter interface {
	Report(msg string)
}
