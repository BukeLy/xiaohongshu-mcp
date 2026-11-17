package main

import (
	"testing"
	"time"

	"github.com/xpzouying/headless_browser"
)

func noBrowser() *headless_browser.Browser { return nil }

func TestQueueBrowserWaitsForRelease(t *testing.T) {
	first := queueBrowser(noBrowser)

	got := make(chan *queuedBrowser)
	go func() { got <- queueBrowser(noBrowser) }()

	select {
	case <-got:
		t.Fatal("第二个浏览器在第一个关闭前就启动了")
	case <-time.After(100 * time.Millisecond):
	}

	first.release()
	select {
	case second := <-got:
		second.release()
	case <-time.After(time.Second):
		t.Fatal("第一个关闭后，第二个仍在排队")
	}
}

func TestQueueBrowserReleasesWhenLaunchPanics(t *testing.T) {
	func() {
		defer func() { _ = recover() }()
		queueBrowser(func() *headless_browser.Browser { panic("launch failed") })
	}()

	done := make(chan struct{})
	go func() {
		queueBrowser(noBrowser).release()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("启动 panic 后名额没有归还，后面的请求一直排队")
	}
}
