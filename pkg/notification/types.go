package notification

type Channel string

const (
	ChannelEmail Channel = "email"
)

type Notification struct {
	To       string
	Subject  string
	HTMLBody string
	TextBody string
}
