package standard

import (
	"errors"
	"fmt"
	. "github.com/xxpxxxxp/goactor"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"sort"
	"sync"
)

type Parameter struct {
	Name  string
	Value string
}

type HttpRequest struct {
	Url     string
	Method  string
	Body    io.Reader
	Headers []*Parameter
}

func (h *HttpRequest) connect(client *http.Client) ([]byte, error) {
	req, _ := http.NewRequest(h.Method, h.Url, h.Body)

	if h.Headers != nil {
		for _, header := range h.Headers {
			req.Header.Set(header.Name, header.Value)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

type HttpResponse struct {
	Body  []byte
	Error error
}

type HttpActor struct {
	Client *http.Client
}

func NewDefaultHttpActor() *HttpActor {
	return &HttpActor{Client: &http.Client{}}
}

func (h *HttpActor) OnPlugin(system *ActorSystem) {}

func (h *HttpActor) Receive(system *ActorSystem, eventType EventType, event interface{}) interface{} {
	type IndexedResponse struct {
		Index   int
		Respose *HttpResponse
	}

	httpAsync := func(index int, client *http.Client, request *HttpRequest, wg *sync.WaitGroup, response chan<- *IndexedResponse) {
		rst, err := request.connect(client)
		response <- &IndexedResponse{index, &HttpResponse{rst, err}}
		wg.Done()
	}

	switch request := event.(type) {
	case []*HttpRequest:
		switch eventType {
		case EVENT_REQUEST:
			// caller doesn't care about result
			for _, r := range request {
				go r.connect(h.Client)
			}
			return nil
		case EVENT_REQUIRE:
			if h.Client.Timeout == 0 {
				// the client must have timeout set, or there's chance to block forever
				responses := make([]*HttpResponse, len(request))
				t := &HttpResponse{nil, errors.New("http client must have timeout set")}
				for i := range responses {
					responses[i] = t
				}
				return responses
			}

			respChan := make(chan *IndexedResponse, len(request))
			var wg sync.WaitGroup
			wg.Add(len(request))
			for i, r := range request {
				go httpAsync(i, h.Client, r, &wg, respChan)
			}
			wg.Wait()

			// get all responses & sort by index
			indexedResponses := make([]*IndexedResponse, len(request))
			for i := range indexedResponses {
				indexedResponses[i] = <-respChan
			}

			sort.Slice(indexedResponses, func(i, j int) bool { return indexedResponses[i].Index < indexedResponses[j].Index })

			responses := make([]*HttpResponse, len(request))
			for i := range indexedResponses {
				responses[i] = indexedResponses[i].Respose
			}
			return responses
		}
	case *HttpRequest:
		switch eventType {
		case EVENT_REQUEST:
			// caller doesn't care about result
			go request.connect(h.Client)
			return nil
		case EVENT_REQUIRE:
			rst, err := request.connect(h.Client)
			return &HttpResponse{rst, err}
		}
	default:
		return errors.New(fmt.Sprintf("unsupported event type \"%s\" for HttpActor", reflect.TypeOf(request).Name()))
	}
	return nil
}

func (h *HttpActor) OnPullout(system *ActorSystem) {}
