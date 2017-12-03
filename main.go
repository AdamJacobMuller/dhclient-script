package main

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netlink/nl"
)

func main() {
	log.Info("starting")

	interfaceName := os.Getenv("interface")

	link, err := netlink.LinkByName(interfaceName)
	if err != nil {
		log.WithFields(log.Fields{
			"error":     err,
			"interface": interfaceName,
		}).Panic("netlink.LinkByName failed")
	}

	reason := os.Getenv("reason")
	switch reason {
	case "PREINIT":
		err := netlink.LinkSetUp(link)
		if err != nil {
			log.WithFields(log.Fields{
				"error":     err,
				"interface": interfaceName,
			}).Panic("netlink.LinkSetUp failed")
		}
		flush(link, nl.FAMILY_ALL)
	case "BOUND":
		fallthrough
	case "RENEW":
		fallthrough
	case "REBIND":
		fallthrough
	case "REBOOT":
		if os.Getenv("old_ip_address") != "" && os.Getenv("old_ip_address") != os.Getenv("new_ip_address") {
			flush(link, nl.FAMILY_V4)
		}
		if os.Getenv("old_ip_address") == "" ||
			os.Getenv("old_ip_address") != os.Getenv("new_ip_address") ||
			reason == "BOUND" ||
			reason == "REBOOT" {
		}
	}
}

func flush(link netlink.Link, family int) error {
	attrs := link.Attrs()
	if attrs == nil {
		panic("")
	}
	interfaceName := attrs.Name

	addrs, err := netlink.AddrList(link, family)
	if err != nil {
		log.WithFields(log.Fields{
			"error":     err,
			"interface": interfaceName,
		}).Panic("netlink.LinkSetUp failed")
	}
	for _, addr := range addrs {
		err := netlink.AddrDel(link, &addr)
		if err != nil {
			log.WithFields(log.Fields{
				"error":     err,
				"interface": interfaceName,
				"address":   fmt.Sprintf("%s/%s", addr.IP, addr.Mask),
			}).Panic("netlink.AddrDel failed")
		}
	}

	return nil
}

func dl() {
	links, err := netlink.LinkList()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Panic("LinkList failed")
	}
	for _, link := range links {
		fmt.Printf("Link: %+v\n", link)
	}
}
