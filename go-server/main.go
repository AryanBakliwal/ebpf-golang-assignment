package main

import (
	"fmt"
	"net/http"
)

func handler1(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello from port 4040!")
}

func handler2(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello from port 5050!")
}

func main() {
	// Server on port 4040
	mux1 := http.NewServeMux()
	mux1.HandleFunc("/", handler1)
	server1 := &http.Server{
		Addr:    ":4040",
		Handler: mux1,
	}

	// Server on port 5050
	mux2 := http.NewServeMux()
	mux2.HandleFunc("/", handler2)
	server2 := &http.Server{
		Addr:    ":5050",
		Handler: mux2,
	}

	go func() {
		fmt.Println("Starting server on port 4040")
		if err := server1.ListenAndServe(); err != nil {
			fmt.Println("Server 4040 error:", err)
		}
	}()

	go func() {
		fmt.Println("Starting server on port 5050")
		if err := server2.ListenAndServe(); err != nil {
			fmt.Println("Server 5050 error:", err)
		}
	}()

	select {}
}
