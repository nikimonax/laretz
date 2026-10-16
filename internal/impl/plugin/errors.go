package plugin

import (
	"errors"
	"fmt"
)

var (
	ErrPluginAlreadyRegistered = errors.New("plugin already registered")
	ErrPluginNotFound          = errors.New("plugin not found")
)

func NewErrPluginAlreadyRegistered(pluginID PluginID) error {
	return fmt.Errorf("%w: %s", ErrPluginAlreadyRegistered, pluginID)
}

func NewErrPluginNotFound(pluginID PluginID) error {
	return fmt.Errorf("%w: %s", ErrPluginNotFound, pluginID)
}
