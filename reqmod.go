package main

import (
	"code.google.com/p/go-icap"
	"fmt"
	"io/ioutil"
	"log"
	"time"
)

// Request-modification functions.

var ISTag = fmt.Sprintf("Redwood%d", time.Now())

func handleRequest(w icap.ResponseWriter, req *icap.Request) {
	h := w.Header()
	h.Set("ISTag", ISTag)
	h.Set("Service", "Redwood content filter")

	switch req.Method {
	case "OPTIONS":
		h.Set("Methods", "REQMOD")
		h.Set("Transfer-Preview", "*")
		h.Set("Preview", "0")
		w.WriteHeader(200, nil, false)

	case "REQMOD":
		if req.Request.Host == "203.0.113.1" {
			icap.ServeLocally(w, req)
			return
		}

		c := context{
			icapRequest: req,
			request:     req.Request,
		}

		c.scanURL()

		if c.action == BLOCK {
			c.showBlockPage(w)
			logChan <- &c
			return
		}

		if changeQuery(c.request.URL) {
			content, err := ioutil.ReadAll(req.Request.Body)
			if err != nil {
				log.Println(err)
			}
			w.WriteHeader(200, req.Request, len(content) > 0)
			if len(content) > 0 {
				w.Write(content)
			}
		} else {
			w.WriteHeader(204, nil, false)
		}

		logChan <- &c

	default:
		w.WriteHeader(405, nil, false)
	}
}

// scanURL calculates scores and an action based on the request's URL.
func (c *context) scanURL() {
	c.tally = URLRules.MatchingRules(c.URL())
	c.calculate(c.user())
}
