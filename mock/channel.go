package mock

type (
	Channel struct{}
)

func (ch *Channel) Ack(tag uint64, multiple bool) error {
	return nil
}

func (ch *Channel) Nack(tag uint64, multiple bool, requeue bool) error {
	return nil
}

func (ch *Channel) Reject(tag uint64, requeue bool) error {
	return nil
}

func (ch *Channel) Publish(exc, route string, msg []byte) error {
	return nil
}
