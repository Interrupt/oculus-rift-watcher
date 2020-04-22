package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	//"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

type OculusWatcher struct {
	WatchUrl string
	WaitTime time.Duration

	ChromeContext context.Context
}

func main() {
	watcher := OculusWatcher{
		WatchUrl: "https://www.oculus.com/rift-s/",
		WaitTime: 3 * time.Second,
	}

	log.Printf("Starting in stock watcher for %s", watcher.WatchUrl)

	// Start the initial context
	ctx, cancel := chromedp.NewContext(context.Background())

	// Loop forever!
	for {
		err := watcher.CheckStockWithTimeout(ctx)

		if err != nil {
			log.Printf("Error: %+v", err)
			cancel()

			// Start a new context, something bad happened with the old one
			ctx, cancel = chromedp.NewContext(context.Background())
		}

		time.Sleep(watcher.WaitTime)
	}
}

func (w *OculusWatcher) CheckStockWithTimeout(ctx context.Context) error {
	watcherChannel := make(chan bool, 1)

	go func() {
		found := w.CheckForStock(ctx)
		watcherChannel <- found
	}()

	select {
	case found := <-watcherChannel:
		if found {
			log.Printf(" - Go now: %s", w.WatchUrl)
			PlayAlertSound()
		}
	case <-time.After(10 * time.Second):
		return fmt.Errorf("Timed out waiting for results.")
	}

	return nil
}

func (w *OculusWatcher) CheckForStock(ctx context.Context) bool {
	var res string

	tasks := chromedp.Tasks{
		chromedp.Navigate(w.WatchUrl),
		chromedp.WaitVisible(`#oculus-rift-s`),
		chromedp.Evaluate(`document.getElementsByClassName("hero__buy-btn-container")[0].innerText;`, &res),
	}

	// run task list
	err := chromedp.Run(ctx, tasks)
	if err != nil {
		log.Println(err)
		return false
	}

	if res == "Notify Me" {
		log.Printf("Not in stock. Button text: %s", res)
		return false
	}

	// Woohoo! In stock hopefully
	log.Printf("Oculus Rift S found in stock! Button Text: %s", res)
	return true
}

func PlayAlertSound() {
	f, err := os.Open("alert.mp3")
	if err != nil {
		log.Printf("Error: Could not open alert sound.")
		return
	}
	defer f.Close()

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		log.Printf("Error: Could not decode alert mp3")
		return
	}
	defer streamer.Close()

	done := make(chan bool)
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	<-done
}
