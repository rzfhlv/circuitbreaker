package main

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/afex/hystrix-go/hystrix"
)

func main() {
	//hystrix config
	hystrix.ConfigureCommand("command_config", hystrix.CommandConfig{
		Timeout:                30000,
		MaxConcurrentRequests:  300,
		RequestVolumeThreshold: 10,
		SleepWindow:            1000,
		ErrorPercentThreshold:  50,
	})

	http.HandleFunc("/", logger(HandleSubsystem))

	fmt.Println("===== Main system is started =====")
	log.Println("listening on :8282")
	http.ListenAndServe(":8282", nil)
}

// HandleSubsystem send request ke external system
func HandleSubsystem(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	resultCh := make(chan []byte)
	errCh := hystrix.Go("command_config", func() error {
		resp, err := http.Get("http://localhost:9090")
		if err != nil {
			return err
		}

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		resultCh <- b
		return nil
	}, nil)

	select {
	case res := <-resultCh:
		log.Println("Request external sistem berhasil:", string(res))
		w.WriteHeader(http.StatusOK)
	case err := <-errCh:
		log.Println("Request external sistem gagal:", err.Error())
		w.WriteHeader(http.StatusServiceUnavailable)
	}
}

func logger(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.URL.Path, r.Method)
		fn(w, r)
	}
}
