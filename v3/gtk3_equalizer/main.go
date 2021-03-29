package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	vlc "github.com/adrg/libvlc-go/v2"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

const appID = "com.github.libvlc-go.gtk3-equalizer-example"

func builderGetObject(builder *gtk.Builder, name string) glib.IObject {
	obj, _ := builder.GetObject(name)
	return obj
}

func assertErr(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func assertConv(ok bool) {
	if !ok {
		log.Panic("invalid widget conversion")
	}
}

func playerReleaseMedia(player *vlc.Player) {
	player.Stop()
	if media, _ := player.Media(); media != nil {
		media.Release()
	}
}

func addFreqScale(label string, container *gtk.Box) *gtk.Scale {
	freqScale, err := gtk.ScaleNewWithRange(gtk.ORIENTATION_VERTICAL, -20, 20, 0.1)
	assertErr(err)
	freqScale.SetValue(0)
	freqScale.SetVExpand(true)
	freqScale.SetHAlign(gtk.ALIGN_CENTER)
	freqScale.SetInverted(true)
	freqScale.SetIncrements(0.1, 0.5)

	freqLabel, err := gtk.LabelNew(label)
	assertErr(err)
	freqLabel.SetHAlign(gtk.ALIGN_CENTER)
	freqLabel.SetMarginTop(10)

	freqBox, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	freqBox.SetVExpand(true)
	freqBox.SetHExpand(true)
	assertErr(err)
	freqBox.Add(freqScale)
	freqBox.Add(freqLabel)
	freqBox.SetMarginTop(10)
	freqBox.SetMarginBottom(10)
	container.Add(freqBox)

	return freqScale
}

func main() {
	// Initialize libVLC module.
	err := vlc.Init("-vvv", "--no-xlib")
	assertErr(err)

	// Create a new player.
	player, err := vlc.NewPlayer()
	assertErr(err)

	// Get equalizer preset names and band frequencies.
	var (
		presetNames = vlc.EqualizerPresetNames()
		bandFreqs   = vlc.EqualizerBandFrequencies()
		equalizer   *vlc.Equalizer
	)

	releaseEqualizer := func() {
		if equalizer != nil {
			err = player.SetEqualizer(nil)
			assertErr(err)
			equalizer.Release()
			equalizer = nil
		}
	}

	// Create new GTK application.
	app, err := gtk.ApplicationNew(appID, glib.APPLICATION_FLAGS_NONE)
	assertErr(err)

	app.Connect("activate", func() {
		// Load application layout.
		builder, err := gtk.BuilderNewFromFile("layout.glade")
		assertErr(err)

		// Get application window.
		appWin, ok := builderGetObject(builder, "appWindow").(*gtk.ApplicationWindow)
		assertConv(ok)

		// Get presets combo box.
		presetsComboBox, ok := builderGetObject(builder, "presetsComboBox").(*gtk.ComboBoxText)
		assertConv(ok)

		// Fill presets combo box.
		for _, presetName := range presetNames {
			presetsComboBox.AppendText(presetName)
		}

		// Get adjustments box.
		adjustmentsBox, ok := builderGetObject(builder, "adjustmentsBox").(*gtk.Box)
		assertConv(ok)
		adjustmentsBox.SetSensitive(false)

		// Fill adjustments box.
		scaleValueChanged := func(scale *gtk.Scale, userData interface{}) {
			if equalizer == nil || scale == nil {
				return
			}

			idx, ok := userData.(int)
			if !ok || idx < -1 || idx >= len(bandFreqs) {
				return
			}

			val := scale.GetValue()
			fmt.Println(idx, val)
			if idx == -1 {
				err = equalizer.SetPreampValue(val)
				assertErr(err)
			} else {
				err = equalizer.SetAmpValueAtIndex(val, uint(idx))
				assertErr(err)
			}
			err = player.SetEqualizer(equalizer)
			assertErr(err)
		}

		preampScale := addFreqScale("Preamp", adjustmentsBox)
		preampScale.Connect("value-changed", scaleValueChanged, -1)

		freqScales := make([]*gtk.Scale, 0, len(bandFreqs))
		for i, bandFreq := range bandFreqs {
			suffix := "Hz"
			if bandFreq >= 1000 {
				suffix = "KHz"
				bandFreq /= 1000
			}

			freqScale := addFreqScale(strconv.FormatFloat(bandFreq, 'f', -1, 64)+" "+suffix, adjustmentsBox)
			freqScales = append(freqScales, freqScale)
			freqScale.Connect("value-changed", scaleValueChanged, i)
		}

		setScaleValues := func() {
			preampVal, err := equalizer.PreampValue()
			assertErr(err)

			preampScale.SetValue(preampVal)
			for i, freqScale := range freqScales {
				freqVal, err := equalizer.AmpValueAtIndex(uint(i))
				assertErr(err)
				freqScale.SetValue(freqVal)
			}
		}

		// Get reset button.
		resetButton, ok := builderGetObject(builder, "resetButton").(*gtk.Button)
		assertConv(ok)
		resetButton.SetSensitive(false)

		// Get media location entry.
		mediaLocationEntry := builderGetObject(builder, "mediaLocationEntry").(*gtk.Entry)
		assertConv(ok)

		// Get play button.
		playButton, ok := builderGetObject(builder, "playButton").(*gtk.Button)
		assertConv(ok)

		// Add builder signal handlers.
		signals := map[string]interface{}{
			"onPresetChanged": func() {
				idx := presetsComboBox.GetActive() - 1
				if idx >= len(presetNames) {
					return
				}
				presetSelected := idx >= 0

				adjustmentsBox.SetSensitive(presetSelected)
				resetButton.SetSensitive(presetSelected)

				// Release previous equalizer.
				releaseEqualizer()

				if !presetSelected {
					preampScale.SetValue(0)
					for _, freqScale := range freqScales {
						freqScale.SetValue(0)
					}
				} else {
					// Create new equalizer from preset.
					equalizer, err = vlc.NewEqualizerFromPreset(uint(idx))
					assertErr(err)
					setScaleValues()
				}

				// Set player equalizer.
				err = player.SetEqualizer(equalizer)
				assertErr(err)
			},
			"onReset": func() {
				idx := presetsComboBox.GetActive() - 1
				if idx < 0 || idx >= len(presetNames) {
					return
				}

				// Release previous equalizer.
				releaseEqualizer()

				// Create new equalizer from preset.
				equalizer, err = vlc.NewEqualizerFromPreset(uint(idx))
				assertErr(err)
				setScaleValues()

				// Set player equalizer.
				err = player.SetEqualizer(equalizer)
				assertErr(err)
			},
			"onChooseFile": func() {
				fileDialog, err := gtk.FileChooserDialogNewWith2Buttons(
					"Choose file...",
					appWin, gtk.FILE_CHOOSER_ACTION_SAVE,
					"Cancel", gtk.RESPONSE_DELETE_EVENT,
					"Save", gtk.RESPONSE_ACCEPT)
				assertErr(err)
				defer fileDialog.Destroy()

				fileFilter, err := gtk.FileFilterNew()
				assertErr(err)
				fileFilter.SetName("Media files")
				fileFilter.AddPattern("*.mp4")
				fileFilter.AddPattern("*.mp3")
				fileDialog.AddFilter(fileFilter)

				if result := fileDialog.Run(); result == gtk.RESPONSE_ACCEPT {
					// Release previous media instance.
					playerReleaseMedia(player)

					location := fileDialog.GetFilename()
					mediaLocationEntry.SetText(location)

					// Set player media and start playback.
					media, err := vlc.NewMediaFromPath(location)
					assertErr(err)
					err = player.SetMedia(media)
					assertErr(err)
					err = player.Play()
					assertErr(err)
					playButton.SetLabel("Pause")
				}
			},
			"onPlay": func() {
				if media, _ := player.Media(); media == nil {
					return
				}

				isPlaying := player.IsPlaying()
				player.SetPause(isPlaying)

				label := "Pause"
				if isPlaying {
					label = "Play"
				}
				playButton.SetLabel(label)
			},
		}
		builder.ConnectSignals(signals)

		appWin.ShowAll()
		app.AddWindow(appWin)
	})

	// Cleanup on exit.
	app.Connect("shutdown", func() {
		releaseEqualizer()
		playerReleaseMedia(player)
		player.Release()
		vlc.Release()
	})

	// Launch the application.
	os.Exit(app.Run(os.Args))
}
