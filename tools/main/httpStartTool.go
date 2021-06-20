package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"net/http"
)

func PortOccupyTool(port string) {
	var Port string
	startCmd := &cobra.Command {
		Use:   "start",
		Short: "Start Http port",
		Run: func(cmd *cobra.Command, args []string) {
			startHttp(Port)
		},
	}
	var rootCmd = &cobra.Command{Use: "http server start"}
	startCmd.Flags().StringVarP(&Port, "port", "p", port, "port to occupy")
	rootCmd.AddCommand(startCmd)
	rootCmd.Execute()
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