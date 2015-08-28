package mock

type (
	Channel struct{}
)

func Ack(tag uint64, multiple bool) error {
	return nil
}

func Nack(tag uint64, multiple bool, requeue bool) error {
	return nil
}

func Reject(tag uint64, requeue bool) error {
	return nil
}
