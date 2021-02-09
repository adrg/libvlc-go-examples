package main

import (
	"log"

	vlc "github.com/adrg/libvlc-go/v2"
)

func main() {
	// Initialize libVLC. Additional command line arguments can be passed in
	// to libVLC by specifying them in the Init function.
	if err := vlc.Init("--no-video", "--quiet"); err != nil {
		log.Fatal(err)
	}
	defer vlc.Release()

	// Create a new list player.
	lp, err := vlc.NewListPlayer()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		lp.Stop()
		lp.Release()
	}()

	// Create a new media list.
	list, err := vlc.NewMediaList()
	if err != nil {
		log.Fatal(err)
	}
	defer list.Release()

	// Add media to list.
	media, err := vlc.NewMediaFromPath("localpath/test.mp3")
	if err != nil {
		log.Fatal(err)
	}
	if err = media.Parse(); err != nil {
		log.Fatal(err)
	}

	err = list.AddMedia(media)
	if err != nil {
		log.Fatal(err)
	}

	// Set player media list.
	if err = lp.SetMediaList(list); err != nil {
		log.Fatal(err)
	}

	// Retrieve player event manager.
	manager, err := lp.EventManager()
	if err != nil {
		log.Fatal(err)
	}

	// Create event handler.
	quit := make(chan struct{})
	eventCallback := func(event vlc.Event, userData interface{}) {
		switch event {
		case vlc.MediaListPlayerPlayed:
			log.Println("Player end reached")
			close(quit)
		case vlc.MediaListPlayerNextItemSet:
			// Retrieve underlying player.
			p, err := lp.Player()
			if err != nil {
				log.Println(err)
				break
			}

			// Retrieve currently playing media.
			media, err := p.Media()
			if err != nil {
				log.Println(err)
				break
			}

			// Get media location.
			location, err := media.Location()
			if err != nil {
				log.Println(err)
				break
			}
			log.Println("Media location:", location)

			// Get media title and artist metadata.
			title, err := media.Meta(vlc.MediaTitle)
			if err != nil {
				log.Println(err)
				break
			}

			artist, err := media.Meta(vlc.MediaArtist)
			if err != nil {
				log.Println(err)
				break
			}

			log.Println("Media title:", title)
			log.Println("Media artist:", artist)
		}
	}

	// Register events with the event manager.
	events := []vlc.Event{
		vlc.MediaListPlayerPlayed,
		vlc.MediaListPlayerNextItemSet,
	}

	var eventIDs []vlc.EventID
	for _, event := range events {
		eventID, err := manager.Attach(event, eventCallback, nil)
		if err != nil {
			log.Fatal(err)
		}

		eventIDs = append(eventIDs, eventID)
	}

	// De-register attached events.
	defer manager.Detach(eventIDs...)

	// Start playing the media list.
	if err = lp.Play(); err != nil {
		log.Fatal(err)
	}

	<-quit
}
