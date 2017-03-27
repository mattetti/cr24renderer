package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"golang.org/x/net/html"

	"bytes"

	"sync"

	"github.com/raff/godet"
)

var flagURL = flag.String("url", "https://splice.com", "url to render")

type HTMLFilter func(htmlStr string) string

func NewScriptRemover(blackList []string) HTMLFilter {
	return func(htmlStr string) string {
		if htmlStr == "" {
			return htmlStr
		}
		doc, err := html.Parse(strings.NewReader(htmlStr))
		if err != nil {
			log.Println("error parsing content when trying to remove scripts", err)
			return htmlStr
		}
		parser := &htmlParser{
			Doc:                doc,
			BlackListedDomains: blackList,
		}
		return parser.Process()
	}
}

func main() {
	flag.Parse()
	// connect to Chrome instance
	remote, err := godet.Connect("localhost:9223", false)

	// disconnect when done
	defer remote.Close()
	wg := sync.WaitGroup{}

	// get browser and protocol version
	// version, err := remote.Version()
	// fmt.Println(version)

	// get list of open tabs
	tabs, err := remote.TabList("")
	for i, tab := range tabs {
		fmt.Printf("tab %d -> %#v\n", i, tab)
	}

	filter := NewScriptRemover(nil)

	// install some callbacks
	remote.CallbackEvent(godet.EventClosed, func(params godet.Params) {
		fmt.Println("RemoteDebugger connection terminated.")
	})

	remote.CallbackEvent("Network.requestWillBeSent", func(params godet.Params) {
		u := params["request"].(map[string]interface{})["url"].(string)
		if len(u) > 42 {
			u = u[:42]
		}
		fmt.Println("requestWillBeSent",
			params["type"],
			params["documentURL"],
			u)
	})

	remote.CallbackEvent("Network.responseReceived", func(params godet.Params) {
		fmt.Println("responseReceived",
			params["type"],
			params["response"].(map[string]interface{})["url"])
	})

	remote.CallbackEvent("Page.loadEventFired", func(params godet.Params) {
		fmt.Printf("Page.loadEventFired - %#v\n", params)
		isReadyResp, err := remote.EvaluateWrap("return window.prerenderReady")
		if err != nil {
			fmt.Println("couldn't get the prerender status")
			return
		}
		isReady, _ := isReadyResp.(bool)
		fmt.Println("prerender ready?", isReady)
		res, err := remote.Evaluate(`document.readyState`)
		fmt.Println("document.readyState", res)
		res, err = remote.Evaluate(`document.documentElement.innerHTML`)
		if err != nil {
			panic(err)
		}
		htmlRes := res.(string)
		// remove script tags
		out := filter(htmlRes)
		filename := "output.html"
		pURL, err := url.Parse(*flagURL)
		if err == nil {
			filename = fmt.Sprintf("%s%s.html", pURL.Host, pURL.RequestURI())
			filename = strings.Replace(filename, "/", "-", -1)
		}
		f, err := os.Create(filename)
		if err != nil {
			panic(err)
		}
		f.WriteString(out)
		f.Close()
		fmt.Println("saved page as", filename)
		wg.Done()
	})

	// remote.CallbackEvent("Log.entryAdded", func(params godet.Params) {
	// 	entry := params["entry"].(map[string]interface{})
	// 	fmt.Println("LOG", entry["type"], entry["level"], entry["text"])
	// })

	// enable event processing
	remote.RuntimeEvents(true)
	remote.NetworkEvents(false)
	remote.PageEvents(true)
	remote.DOMEvents(true)
	remote.LogEvents(true)

	// navigate in existing tab
	if err = remote.ActivateTab(tabs[0]); err != nil {
		panic(err)
	}
	frameID, err := remote.Navigate(*flagURL)
	if err != nil {
		panic(err)
	}
	fmt.Printf("navigated to %s with frame ID %s\n", *flagURL, frameID)
	wg.Add(1)
	wg.Wait()
}

type htmlParser struct {
	// Doc is the root node
	Doc *html.Node
	// BlackListedDomains are the domains for filtering reasons
	BlackListedDomains []string
	// TODO: add a logger
}

func (p *htmlParser) traverseNode(n *html.Node) {
	p.scriptNodeRemover(n)

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		p.traverseNode(c)
	}
}

// Process parses the document and filter it according to the setup filters.
func (p *htmlParser) Process() string {
	if p == nil || p.Doc == nil {
		return ""
	}
	p.traverseNode(p.Doc)
	w := bytes.NewBuffer(nil)
	if err := html.Render(w, p.Doc); err != nil {
		fmt.Println("failed to render the parsed doc", err)
		return ""
	}
	return w.String()
}

func (p *htmlParser) isScriptElement(n *html.Node) bool {
	return n.Type == html.ElementNode && n.Data == "script"
}

func (p *htmlParser) scriptNodeRemover(n *html.Node) bool {
	if p.isScriptElement(n) {
		/*
			for _, attr := range n.Attr {
				if attr.Key == "src" && strings.Contains(attr.Val, ".splice.com") {
					skip = false
					break
				}
			}
		*/
		fmt.Printf("removing %v\n", n.Attr)
		for c := n.FirstChild; c != nil; {
			if c.Parent != n {
				break
			}
			n.RemoveChild(c)
		}
		// clear the attributes
		n.Attr = nil
		return true
	}
	return false
}
