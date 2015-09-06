package server

// VHost is a fake AMQP virtual host
type VHost struct {
	name      string
	exchanges map[string]Exchange
	queues    map[string]Queue
}

// NewVHost create a new fake AMQP Virtual Host
func NewVHost(name string) VHost {
	vh := VHost{
		name:   name,
		queues: make(map[string]Queue),
	}

	vh.createDefaultExchanges()
	return vh
}

func (vhost *VHost) createDefaultExchanges() {
	exchs := make(map[string]Exchange)
	exchs["amqp.topic"] = &TopicExchange{}
	vhost.exchanges = exchs
}
