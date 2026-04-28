package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func usage() {
	fmt.Fprintln(flag.CommandLine.Output(), "Usage:")
	fmt.Fprintln(flag.CommandLine.Output(), "  go2web -u <URL>         make an HTTP request to the specified URL and print the response")
	fmt.Fprintln(flag.CommandLine.Output(), "  go2web -s <search-term> make an HTTP request to search the term and print top 10 results")
	fmt.Fprintln(flag.CommandLine.Output(), "  go2web -h               show this help")
}

func runCLI() {
	url := flag.String("u", "", "make an HTTP request to the specified URL")
	searchTerm := flag.String("s", "", "make an HTTP request to search the term")
	flag.Usage = usage
	flag.Parse()
	if *searchTerm != "" && len(flag.Args()) > 0 {
		*searchTerm = strings.Join(append([]string{*searchTerm}, flag.Args()...), " ")
	}

	if *url == "" && *searchTerm == "" {
		flag.Usage()
		return
	}

	if *url != "" {
		if err := fetchURL(*url); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	if *searchTerm != "" {
		if err := searchWeb(*searchTerm); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
}