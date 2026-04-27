package main

import (
	"flag"
	"fmt"
	"os"
)

func usage() {
	fmt.Fprintln(flag.CommandLine.Output(), "Usage:")
	fmt.Fprintln(flag.CommandLine.Output(), "  go2web -u <URL>         make an HTTP request to the specified URL and print the response")
	fmt.Fprintln(flag.CommandLine.Output(), "  go2web -s <search-term> make an HTTP request to search the term and print top 10 results")
	fmt.Fprintln(flag.CommandLine.Output(), "  go2web -h               show this help")
}

func main() {
	url := flag.String("u", "", "make an HTTP request to the specified URL")
	searchTerm := flag.String("s", "", "make an HTTP request to search the term")
	flag.Usage = usage
	flag.Parse()

	if *url == "" && *searchTerm == "" {
		flag.Usage()
		return
	}

	if *url != "" {
		fmt.Fprintln(os.Stderr, "-u is not implemented yet")
		os.Exit(1)
	}

	if *searchTerm != "" {
		fmt.Fprintln(os.Stderr, "-s is not implemented yet")
		os.Exit(1)
	}
}