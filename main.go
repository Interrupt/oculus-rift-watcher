package main

import (
	"context"
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
}

func main() {
	watcher := OculusWatcher{
		WatchUrl: "https://www.oculus.com/rift-s/",
		WaitTime: 3 * time.Second,
	}

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	for {
		found := watcher.CheckForStock(ctx)
		if found {
			log.Println("Oculus Rift S found in stock!")
			PlayAlertSound()
			continue
		}

		time.Sleep(watcher.WaitTime)
	}
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
	}

	log.Printf("Not in stock yet. Button text: %s", res)

	if res == "Notify Me" {
		return false
	}

	return true
}

func PlayAlertSound() {
	f, err := os.Open("alert.mp3")
	if err != nil {
		log.Fatal(err)
	}

	streamer, format, err := mp3.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	defer streamer.Close()

	done := make(chan bool)
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	<-done
}
