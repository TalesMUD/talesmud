package bus

import (
	bus "github.com/asaskevich/EventBus"
)

// Global bus for all components
var Global = bus.New()
