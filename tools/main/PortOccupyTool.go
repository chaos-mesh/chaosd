// Copyright 2020 Chaos Mesh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

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