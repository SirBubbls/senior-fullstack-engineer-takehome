package broadcast

import (
	"sync"
	"sirbubbls.io/challenge/models"
)

// WeatherUpdateBroadcast manages broadcasting to multiple subscribers
type WeatherUpdateBroadcast struct {
	mu      sync.RWMutex
	clients map[chan models.Weather]struct{}
}

func NewBroadcastHub() *WeatherUpdateBroadcast {
	return &WeatherUpdateBroadcast{
		clients: make(map[chan models.Weather]struct{}),
	}
}
func (b *WeatherUpdateBroadcast) Subscribe() chan models.Weather {
	ch := make(chan models.Weather, 10)
	b.mu.Lock()
	b.clients[ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

func (b *WeatherUpdateBroadcast) Unsubscribe(ch chan models.Weather) {
	b.mu.Lock()
	delete(b.clients, ch)
	close(ch)
	b.mu.Unlock()
}

func (b *WeatherUpdateBroadcast) Broadcast(msg models.Weather) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.clients {
		select {
		case ch <- msg:
		default: // avoid blocking if a client is slow
		}
	}
}
