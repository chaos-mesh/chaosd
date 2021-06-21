package main

import (
	"fmt"
	"github.com/pingcap/errors"
	"github.com/spf13/cobra"
	"log"
	"net/http"
)

func PortOccupyTool(port string) error {
	var Port string
	var rootCmd = &cobra.Command {
		Use:   "http server start",
		Short: "http server start",
		Run: func(cmd *cobra.Command, args []string) {
			startHttp(Port)
		},
	}

	rootCmd.Flags().StringVarP(&Port, "port", "p", port, "port to occupy")
	if err := rootCmd.Execute(); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func generateHttpPort(port string) string {
	s := fmt.Sprintf(":%s", port)
	return s
}

func startHttp(porttoOccupy string) {

	s := generateHttpPort(porttoOccupy)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello cmd!")
	})
	if err := http.ListenAndServe(s, nil); err != nil {
		log.Println("ListenAndServe: ", err)
	}
}


func main() {
	PortOccupyTool("")
}