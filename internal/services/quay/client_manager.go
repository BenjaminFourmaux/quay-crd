package quay

import "sync"

type ClientManager struct {
	mu     sync.RWMutex
	client *Client
}

/*
NewClientManager is the constructor of Quay ClientManager
*/
func NewClientManager(client *Client) *ClientManager {
	return &ClientManager{
		client: client,
	}
}

func (cm *ClientManager) Get() *Client {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.client
}

func (cm *ClientManager) Set(client *Client) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.client = client
}

func (cm *ClientManager) Clear() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.client = nil
}
