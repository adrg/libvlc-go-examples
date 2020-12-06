package main

import (
	"fmt"
	"log"

	vlc "github.com/adrg/libvlc-go/v2"
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
	if err := media.Parse(); err != nil {
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
		case vlc.MediaTrackText:
			subtitle := track.Subtitle
			fmt.Println("Type: subtitle track")
			fmt.Println("Encoding:", subtitle.Encoding)
		}
		fmt.Println("---")
	}
}
