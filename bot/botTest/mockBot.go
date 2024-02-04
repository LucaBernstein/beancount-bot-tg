package botTest

import (
	"time"

	"github.com/LucaBernstein/beancount-bot-tg/v2/helpers"
	tb "gopkg.in/telebot.v3"
)

type MockBot struct {
	LastSentWhat    interface{}
	AllLastSentWhat []interface{}
}

func (b *MockBot) Start()                                                                       {}
func (b *MockBot) Handle(endpoint interface{}, handler tb.HandlerFunc, mw ...tb.MiddlewareFunc) {}
func (b *MockBot) Send(to tb.Recipient, what interface{}, options ...interface{}) (*tb.Message, error) {
	b.LastSentWhat = what
	b.AllLastSentWhat = append(b.AllLastSentWhat, what)
	return nil, nil
}
func (b *MockBot) Respond(c *tb.Callback, resp ...*tb.CallbackResponse) error {
	return nil
}
func (b *MockBot) Me() *tb.User {
	return &tb.User{Username: "Test bot"}
}
func (b *MockBot) SendSilent(logFn func(level helpers.Level, m *tb.Message, format string, v ...interface{}), to tb.Recipient, what interface{}, options ...interface{}) (*tb.Message, error) {
	return b.Send(to, what, options...)
}
func (b *MockBot) Reset() {
	b.AllLastSentWhat = nil
}

type MockContext struct {
	M *tb.Message
}

func (c *MockContext) Message() *tb.Message {
	return c.M
}
func (c *MockContext) Bot() *tb.Bot                                            { return nil }
func (c *MockContext) Update() tb.Update                                       { return tb.Update{} }
func (c *MockContext) Callback() *tb.Callback                                  { return nil }
func (c *MockContext) Query() *tb.Query                                        { return nil }
func (c *MockContext) InlineResult() *tb.InlineResult                          { return nil }
func (c *MockContext) ShippingQuery() *tb.ShippingQuery                        { return nil }
func (c *MockContext) PreCheckoutQuery() *tb.PreCheckoutQuery                  { return nil }
func (c *MockContext) Poll() *tb.Poll                                          { return nil }
func (c *MockContext) PollAnswer() *tb.PollAnswer                              { return nil }
func (c *MockContext) ChatMember() *tb.ChatMemberUpdate                        { return nil }
func (c *MockContext) ChatJoinRequest() *tb.ChatJoinRequest                    { return nil }
func (c *MockContext) Migration() (int64, int64)                               { return 0, 0 }
func (c *MockContext) Sender() *tb.User                                        { return nil }
func (c *MockContext) Chat() *tb.Chat                                          { return nil }
func (c *MockContext) Recipient() tb.Recipient                                 { return nil }
func (c *MockContext) Text() string                                            { return "" }
func (c *MockContext) Data() string                                            { return "" }
func (c *MockContext) Args() []string                                          { return nil }
func (c *MockContext) Send(what interface{}, opts ...interface{}) error        { return nil }
func (c *MockContext) SendAlbum(a tb.Album, opts ...interface{}) error         { return nil }
func (c *MockContext) Reply(what interface{}, opts ...interface{}) error       { return nil }
func (c *MockContext) Forward(msg tb.Editable, opts ...interface{}) error      { return nil }
func (c *MockContext) ForwardTo(to tb.Recipient, opts ...interface{}) error    { return nil }
func (c *MockContext) Edit(what interface{}, opts ...interface{}) error        { return nil }
func (c *MockContext) EditCaption(caption string, opts ...interface{}) error   { return nil }
func (c *MockContext) EditOrSend(what interface{}, opts ...interface{}) error  { return nil }
func (c *MockContext) EditOrReply(what interface{}, opts ...interface{}) error { return nil }
func (c *MockContext) Delete() error                                           { return nil }
func (c *MockContext) DeleteAfter(d time.Duration) *time.Timer                 { return nil }
func (c *MockContext) Notify(action tb.ChatAction) error                       { return nil }
func (c *MockContext) Ship(what ...interface{}) error                          { return nil }
func (c *MockContext) Accept(errorMessage ...string) error                     { return nil }
func (c *MockContext) Answer(resp *tb.QueryResponse) error                     { return nil }
func (c *MockContext) Respond(resp ...*tb.CallbackResponse) error              { return nil }
func (c *MockContext) Get(key string) interface{}                              { return nil }
func (c *MockContext) Set(key string, val interface{})                         {}
func (c *MockContext) Entities() tb.Entities                                   { return nil }
func (c *MockContext) Topic() *tb.Topic                                        { return nil }

// Test type matching
var _ tb.Context = &MockContext{}
