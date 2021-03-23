package main

import (
	"log"
	"os"

	vlc "github.com/adrg/libvlc-go/v3"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

const appID = "com.github.libvlc-go.gtk3-media-discovery-example"

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

func clearListBox(l *gtk.ListBox) {
	l.GetChildren().Foreach(func(row interface{}) {
		widget, ok := row.(*gtk.Widget)
		if !ok {
			return
		}
		l.Remove(widget)
	})
}

func activateListBoxRows(l *gtk.ListBox) {
	l.GetChildren().Foreach(func(row interface{}) {
		widget, ok := row.(*gtk.Widget)
		if !ok {
			return
		}
		widget.SetSensitive(true)
	})
}

func main() {
	// Initialize libVLC module.
	err := vlc.Init("--quiet", "--no-xlib")
	assertErr(err)

	player, err := vlc.NewListPlayer()
	assertErr(err)

	var (
		serviceIdx         int = -1
		service            *vlc.MediaDiscoverer
		serviceDescriptors []*vlc.MediaDiscovererDescriptor
	)

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

		// Get discovery service category combo box.
		categoryComboBox, ok := builderGetObject(builder, "categoryComboBox").(*gtk.ComboBoxText)
		assertConv(ok)

		// Get discovery services list box.
		servicesListBox, ok := builderGetObject(builder, "servicesListBox").(*gtk.ListBox)
		assertConv(ok)

		// Get start button.
		startButton, ok := builderGetObject(builder, "startButton").(*gtk.Button)
		assertConv(ok)
		startButton.SetSensitive(false)

		// Get media list box.
		mediaListBox, ok := builderGetObject(builder, "mediaListBox").(*gtk.ListBox)
		assertConv(ok)

		// Get play button.
		playButton, ok := builderGetObject(builder, "playButton").(*gtk.Button)
		assertConv(ok)
		playButton.SetSensitive(false)

		// Get pause button.
		pauseButton, ok := builderGetObject(builder, "pauseButton").(*gtk.Button)
		assertConv(ok)
		pauseButton.SetSensitive(false)

		// Add builder signal handlers.
		signals := map[string]interface{}{
			"onServiceCategoryChange": func() {
				// Deactivate buttons.
				playButton.SetSensitive(false)
				pauseButton.SetSensitive(false)

				// Clear discovery service list.
				clearListBox(servicesListBox)

				// Clear media list.
				clearListBox(mediaListBox)

				// Stop player.
				player.Stop()

				// Release current service, if any.
				if service != nil {
					assertErr(service.Release())
					service = nil
					serviceIdx = -1
				}

				// Get selected discovery service category.
				category := categoryComboBox.GetActive() - 1
				if category < 0 {
					startButton.SetSensitive(false)
					return
				}

				// Get discovery service descriptors.
				serviceDescriptors, err = vlc.ListMediaDiscoverers(vlc.MediaDiscoveryCategory(category))
				assertErr(err)

				// Add discovery services to the list.
				for _, serviceDescriptor := range serviceDescriptors {
					nameLabel, err := gtk.LabelNew(serviceDescriptor.Name)
					assertErr(err)
					nameLabel.SetHAlign(gtk.ALIGN_START)
					nameLabel.SetMarginStart(5)
					longNameLabel, err := gtk.LabelNew(serviceDescriptor.LongName)
					assertErr(err)
					longNameLabel.SetMarginStart(20)
					longNameLabel.SetHAlign(gtk.ALIGN_START)

					rowBox, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
					assertErr(err)
					rowBox.SetMarginTop(5)
					rowBox.SetMarginBottom(5)
					rowBox.Add(nameLabel)
					rowBox.Add(longNameLabel)

					row, err := gtk.ListBoxRowNew()
					assertErr(err)

					row.Add(rowBox)
					servicesListBox.Add(row)
					servicesListBox.ShowAll()
				}
			},
			"onServiceListRowSelected": func() {
				row := servicesListBox.GetSelectedRow()
				if row == nil {
					startButton.SetSensitive(false)
					return
				}

				idx := row.GetIndex()
				startButton.SetSensitive(idx >= 0 && idx != serviceIdx)
			},
			"onServiceStart": func() {
				row := servicesListBox.GetSelectedRow()
				if row == nil {
					return
				}

				idx := row.GetIndex()
				if idx < 0 || idx > len(serviceDescriptors)-1 || idx == serviceIdx {
					return
				}

				// Deactivate buttons.
				playButton.SetSensitive(false)
				pauseButton.SetSensitive(false)

				// Stop player.
				player.Stop()

				// Release current service, if any.
				if service != nil {
					assertErr(service.Release())
					service = nil
					serviceIdx = -1
				}

				// Clear media list.
				clearListBox(mediaListBox)

				// Create media discovery service.
				service, err = vlc.NewMediaDiscoverer(serviceDescriptors[idx].Name)
				if err != nil {
					log.Printf("ERROR: %v\n", err)
					return
				}
				serviceIdx = idx

				activateListBoxRows(servicesListBox)
				row.SetSensitive(false)
				startButton.SetSensitive(false)

				// Start media discovery service.
				if err = service.Start(func(event vlc.Event, media *vlc.Media, index int) {
					switch event {
					case vlc.MediaListItemAdded:
						location, err := media.Location()
						assertErr(err)

						nameLabel, err := gtk.LabelNew(location)
						assertErr(err)
						nameLabel.SetHAlign(gtk.ALIGN_START)
						nameLabel.SetMarginStart(5)

						rowBox, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
						assertErr(err)
						rowBox.SetMarginTop(5)
						rowBox.SetMarginBottom(5)
						rowBox.Add(nameLabel)

						row, err := gtk.ListBoxRowNew()
						assertErr(err)

						row.Add(rowBox)
						mediaListBox.Insert(row, index)
						mediaListBox.ShowAll()
					case vlc.MediaListItemDeleted:
						row := mediaListBox.GetRowAtIndex(index)
						if row != nil {
							mediaListBox.Remove(row)
							mediaListBox.ShowAll()
						}
					}
				}); err != nil {
					log.Printf("ERROR: %v\n", err)
					return
				}

				mediaList, err := service.MediaList()
				assertErr(err)
				player.SetMediaList(mediaList)
			},
			"onMediaListRowSelected": func() {
				row := mediaListBox.GetSelectedRow()
				if row == nil {
					playButton.SetSensitive(false)
					return
				}

				idx := row.GetIndex()
				playButton.SetSensitive(idx >= 0)
			},
			"onMediaPlay": func() {
				row := mediaListBox.GetSelectedRow()
				if row == nil {
					return
				}

				idx := row.GetIndex()
				if idx < 0 {
					return
				}

				err = player.PlayAtIndex(uint(idx))
				assertErr(err)

				activateListBoxRows(mediaListBox)
				row.SetSensitive(false)
				playButton.SetSensitive(false)
				pauseButton.SetSensitive(true)
				pauseButton.SetLabel("Pause")
			},
			"onMediaStop": func() {
				isPlaying := player.IsPlaying()
				player.SetPause(isPlaying)

				label := "Pause"
				if isPlaying {
					label = "Resume"
				}
				pauseButton.SetLabel(label)
			},
		}
		builder.ConnectSignals(signals)

		appWin.ShowAll()
		app.AddWindow(appWin)
	})

	// Cleanup on exit.
	app.Connect("shutdown", func() {
		// Release media discovery service.
		if service != nil {
			service.Release()
		}

		// Release player.
		player.Stop()
		player.Release()

		// Release VLC module.
		vlc.Release()
	})

	// Launch the application.
	os.Exit(app.Run(os.Args))
}
