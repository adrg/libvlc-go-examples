package main

import (
	"errors"
	"log"

	vlc "github.com/adrg/libvlc-go/v3"
)

func main() {
	// Initialize libVLC. Additional command line arguments can be passed in
	// to libVLC by specifying them in the Init function.
	if err := vlc.Init("--quiet"); err != nil {
		log.Fatal(err)
	}
	defer vlc.Release()

	// Get discovery service.
	discoverer, err := getDiscoverer()
	if err != nil {
		log.Fatal(err)
	}
	defer discoverer.Release()

	// Get Chromecast renderer.
	renderer, err := getRenderer(discoverer)
	if err != nil {
		log.Fatal(err)
	}

	// Create a new player.
	player, err := vlc.NewPlayer()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		player.Stop()
		player.Release()
	}()

	// Load player media from path.
	media, err := player.LoadMediaFromPath("test.mp4")
	if err != nil {
		log.Fatal(err)
	}
	defer media.Release()

	// Set renderer.
	if err := player.SetRenderer(renderer); err != nil {
		log.Fatal(err)
	}

	// Retrieve player event manager.
	manager, err := player.EventManager()
	if err != nil {
		log.Fatal(err)
	}

	// Register the media end reached event with the event manager.
	quit := make(chan struct{})
	eventCallback := func(event vlc.Event, userData interface{}) {
		close(quit)
	}

	eventID, err := manager.Attach(vlc.MediaPlayerEndReached, eventCallback, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer manager.Detach(eventID)

	// Start media playback.
	if err = player.Play(); err != nil {
		log.Fatal(err)
	}

	<-quit
}

func getDiscoverer() (*vlc.RendererDiscoverer, error) {
	// Get discovery service descriptors.
	descriptors, err := vlc.ListRendererDiscoverers()
	if err != nil {
		return nil, err
	}

	// Search for mDNS discovery service.
	for _, descriptor := range descriptors {
		if descriptor.Name != "microdns_renderer" {
			continue
		}

		discoverer, err := vlc.NewRendererDiscoverer(descriptor.Name)
		if err != nil {
			return nil, err
		}

		return discoverer, nil
	}

	return nil, errors.New("could not find discovery service")
}

func getRenderer(discoverer *vlc.RendererDiscoverer) (*vlc.Renderer, error) {
	// Start renderer discovery.
	var renderer *vlc.Renderer

	stop := make(chan error)
	if err := discoverer.Start(func(event vlc.Event, r *vlc.Renderer) {
		// NOTE: the discovery service cannot be stopped or released from
		// the callback function. Doing so will result in undefined behavior.

		switch event {
		case vlc.RendererDiscovererItemAdded:
			// New renderer (`r`) found.
			rendererType, err := r.Type()
			if err != nil {
				stop <- err
			}
			if rendererType == vlc.RendererChromecast {
				renderer = r
				stop <- nil
			}
		case vlc.RendererDiscovererItemDeleted:
			// The renderer (`r`) is no longer available.
		}
	}); err != nil {
		return nil, err
	}

	if err := <-stop; err != nil {
		return nil, err
	}
	if err := discoverer.Stop(); err != nil {
		return nil, err
	}

	return renderer, nil
}
