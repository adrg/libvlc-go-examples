package main

import (
	"fmt"
	"log"

	vlc "github.com/adrg/libvlc-go/v3"
)

func main() {
	// Initialize libVLC.
	if err := vlc.Init("--quiet"); err != nil {
		log.Fatal(err)
	}
	defer vlc.Release()

	// Load media from file.
	media, err := vlc.NewMediaFromPath("test.mp4")
	if err != nil {
		log.Fatal(err)
	}

	// Parse media.
	if err := parseMedia(media); err != nil {
		log.Fatal(err)
	}

	// Retrieve media tracks.
	tracks, err := media.Tracks()
	if err != nil {
		log.Fatal(err)
	}

	for i, track := range tracks {
		fmt.Printf("Track #%d\n", i+1)
		fmt.Println("ID:", track.ID)
		fmt.Println("Bit rate:", track.BitRate)
		fmt.Println("Codec:", track.Codec)
		fmt.Println("Original codec:", track.OriginalCodec)
		fmt.Println("Profile:", track.Profile)
		fmt.Println("Level:", track.Level)
		fmt.Println("Language:", track.Language)
		fmt.Println("Description:", track.Description)

		// Retrieve codec description.
		codecDesc, err := track.CodecDescription()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Codec description:", codecDesc)

		switch track.Type {
		case vlc.MediaTrackAudio:
			audio := track.Audio
			fmt.Println("Type: audio track")
			fmt.Println("Audio channels:", audio.Channels)
			fmt.Println("Audio rate:", audio.Rate)
		case vlc.MediaTrackVideo:
			video := track.Video
			fmt.Println("Type: video track")
			fmt.Println("Video width:", video.Width)
			fmt.Println("Video height:", video.Height)
			fmt.Printf("Aspect ratio: %d:%d\n", video.AspectRatioNum, video.AspectRatioDen)
			fmt.Printf("Frame rate: %d/%d\n", video.FrameRateNum, video.FrameRateDen)
			fmt.Println("Video orientation", video.Orientation)
			fmt.Println("Video projection", video.Projection)

			pose := video.Pose
			fmt.Printf("Video viewopoint: %.2f yaw, %.2f pitch, %.2f roll, %.2f FOV\n",
				pose.Yaw, pose.Pitch, pose.Roll, pose.FOV)
		case vlc.MediaTrackText:
			subtitle := track.Subtitle
			fmt.Println("Type: subtitle track")
			fmt.Println("Encoding:", subtitle.Encoding)
		}
		fmt.Println("---")
	}
}

func parseMedia(media *vlc.Media) error {
	// Retrieve media event manager.
	manager, err := media.EventManager()
	if err != nil {
		log.Fatal(err)
	}

	// Create media event handler.
	done := make(chan struct{})
	eventCallback := func(event vlc.Event, userData interface{}) {
		parseStatus, parseErr := media.ParseStatus()
		if err != nil {
			err = parseErr
		} else if parseStatus != vlc.MediaParseDone {
			err = fmt.Errorf("media parse failed with status %d", parseStatus)
		}
		close(done)
	}

	// Register media parsed changed event with the media event manager.
	eventID, err := manager.Attach(vlc.MediaParsedChanged, eventCallback, media)
	if err != nil {
		return err
	}
	defer manager.Detach(eventID)

	// Parse media asynchronously.
	err = media.ParseWithOptions(0, vlc.MediaParseLocal, vlc.MediaParseNetwork)
	if err != nil {
		return err
	}

	<-done
	return err
}
