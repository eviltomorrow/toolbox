package main

import (
	"fmt"

	"github.com/coreos/go-iptables/iptables"
)

func contains(list []string, value string) bool {
	for _, val := range list {
		if val == value {
			return true
		}
	}
	return false
}

func main() {
	ipt, err := iptables.New()

	// Saving the list of chains before executing tests
	originaListChain, err := ipt.ListChains("filter")
	if err != nil {
		fmt.Printf("ListChains of Initial failed: %v", err)
	}

	for _, chain := range originaListChain {
		fmt.Println(chain)
	}

	chain := "INPUT"
	err = ipt.ClearChain("filter", chain)
	if err != nil {
		fmt.Printf("ClearChain (of missing) failed: %v", err)
	}
}
