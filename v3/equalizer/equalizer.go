package main

import (
	"fmt"
	"log"
	"time"

	vlc "github.com/adrg/libvlc-go/v3"
)

func main() {
	// Initialize libVLC. Additional command line arguments can be passed in
	// to libVLC by specifying them in the Init function.
	if err := vlc.Init("--no-video", "--quiet"); err != nil {
		log.Fatal(err)
	}
	defer vlc.Release()

	// Get equalizer preset names.
	var (
		eqPresetNames = vlc.EqualizerPresetNames()
		presetIdx     uint
	)

	fmt.Println("Equalizer presets: ")
	for i, eqPresetName := range eqPresetNames {
		if eqPresetName == "Full bass" {
			presetIdx = uint(i)
		}

		fmt.Printf("#%d: %s\n", i, eqPresetName)
	}
	fmt.Println("")

	// NOTE: in order to get a single equalizer preset, use
	// vlc.EqualizerPresetName. Use EqualizerPresetCount to
	// obtain the number of available equalizer presets.

	// Get equalizer band frequencies.
	bandFreqs := vlc.EqualizerBandFrequencies()

	fmt.Println("Equalizer band frequencies: ")
	for i, bandFreq := range bandFreqs {
		fmt.Printf("#%d: %.2f\n", i, bandFreq)
	}
	fmt.Println("")

	// NOTE: in order to get a single band frequency, use
	// vlc.EqualizerBandFrequency. Use EqualizerBandCount
	// to obtain the number of available equalizer bands.

	// Create a new equalizer from a preset.
	// If you want to start from scratch, use vlc.NewEqualizer.
	equalizer, err := vlc.NewEqualizerFromPreset(presetIdx)
	if err != nil {
		log.Fatal(err)
	}
	defer equalizer.Release()

	// Get and set preamplification value.
	preAmp, err := equalizer.PreampValue()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Preamp value: %.2f\n", preAmp)

	if err := equalizer.SetPreampValue(preAmp - 3); err != nil {
		log.Fatal(err)
	}

	// Get and set individidual amplification values for the
	// equalizer frequency bands.
	bandCount := vlc.EqualizerBandCount()
	for i := uint(0); i < bandCount; i++ {
		bandFreq, err := equalizer.AmpValueAtIndex(i)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("#%d (%.2f): %.2f\n", i, bandFreqs[i], bandFreq)

		if err := equalizer.SetAmpValueAtIndex(bandFreq+float64(i), i); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println("")

	// Create a new player.
	player, err := vlc.NewPlayer()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		player.Stop()
		player.Release()
	}()

	// Add a media file from path or from URL.
	// Set player media from path:
	// media, err := player.LoadMediaFromPath("localpath/test.mp4")
	// Set player media from URL:
	media, err := player.LoadMediaFromURL("http://stream-uk1.radioparadise.com/mp3-32")
	if err != nil {
		log.Fatal(err)
	}
	defer media.Release()

	if err := player.SetEqualizer(equalizer); err != nil {
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

	// Start playing the media.
	if err = player.Play(); err != nil {
		log.Fatal(err)
	}

	// Revert to the default equalizer.
	time.Sleep(time.Duration(10) * time.Second)
	if err := player.SetEqualizer(nil); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Reverted to the default equalizer settings")

	<-quit
}
