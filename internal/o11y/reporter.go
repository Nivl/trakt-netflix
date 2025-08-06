package o11y

type Reporter interface {
	SendMessage(msg string) error
}
