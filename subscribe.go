package main

type subscriptions map[string][]chan []byte

func (s subscriptions) subscribeUser(
	name string,
) (int, chan []byte) {

	messages := make(chan []byte)
	s[name] = append(s[name], messages)

	return len(s[name]) - 1, messages
}

func (s subscriptions) getActiveSubscribers(
	name string,
) int {
	active := 0

	for _, messages := range s[name] {
		if messages != nil {
			active++
		}
	}

	return active
}

func (s subscriptions) publishMessage(
	name string, message []byte,
) {
	for _, messages := range s[name] {
		if messages != nil {
			messages <- message
		}
	}
}

func (s subscriptions) setInactive(
	name string, idx int,
) {
	s[name][idx] = nil
}

func (s subscriptions) unsubscribeUser(
	name string,
) {
	delete(s, name)
}
