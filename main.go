package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netlink/nl"
)

func subnet_mask_to_cidr(mask string) (string, error) {
	switch mask {
	case "0.0.0.0":
		return "0", nil
	case "128.0.0.0":
		return "1", nil
	case "192.0.0.0":
		return "2", nil
	case "224.0.0.0":
		return "3", nil
	case "240.0.0.0":
		return "4", nil
	case "248.0.0.0":
		return "5", nil
	case "252.0.0.0":
		return "6", nil
	case "254.0.0.0":
		return "7", nil
	case "255.0.0.0":
		return "8", nil
	case "255.128.0.0":
		return "9", nil
	case "255.192.0.0":
		return "10", nil
	case "255.224.0.0":
		return "11", nil
	case "255.240.0.0":
		return "12", nil
	case "255.248.0.0":
		return "13", nil
	case "255.252.0.0":
		return "14", nil
	case "255.254.0.0":
		return "15", nil
	case "255.255.0.0":
		return "16", nil
	case "255.255.128.0":
		return "17", nil
	case "255.255.192.0":
		return "18", nil
	case "255.255.224.0":
		return "19", nil
	case "255.255.240.0":
		return "20", nil
	case "255.255.248.0":
		return "21", nil
	case "255.255.252.0":
		return "22", nil
	case "255.255.254.0":
		return "23", nil
	case "255.255.255.0":
		return "24", nil
	case "255.255.255.128":
		return "25", nil
	case "255.255.255.192":
		return "26", nil
	case "255.255.255.224":
		return "27", nil
	case "255.255.255.240":
		return "28", nil
	case "255.255.255.248":
		return "29", nil
	case "255.255.255.252":
		return "30", nil
	case "255.255.255.254":
		return "31", nil
	case "255.255.255.255":
		return "32", nil
	default:
		return "", errors.New("invalid subnet mask")
	}
}

func main() {
	log.SetLevel(log.DebugLevel)

	interfaceName := os.Getenv("interface")
	reason := os.Getenv("reason")

	log.WithFields(log.Fields{
		"interface": interfaceName,
		"reason":    reason,
	}).Info("starting")

	link, err := netlink.LinkByName(interfaceName)
	if err != nil {
		log.WithFields(log.Fields{
			"error":     err,
			"interface": interfaceName,
		}).Panic("netlink.LinkByName failed")
	}

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
			update_ip_address(link)

		}
	case "EXPIRE":
		fallthrough
	case "FAIL":
		fallthrough
	case "RELEASE":
		fallthrough
	case "STOP":
		flush(link, nl.FAMILY_V4)
	case "TIMEOUT":
		update_ip_address(link)
		set_mtu(link)
	default:
		log.WithFields(log.Fields{
			"reason": reason,
		}).Error("reason unhandled")
	}
}

func set_mtu(link netlink.Link) {
	interfaceName := linkName(link)
	mtu := os.Getenv("new_interface_mtu")
	if mtu == "" {
		return
	}
	mtuInt, err := strconv.ParseInt(mtu, 10, 64)
	if err != nil {
		log.WithFields(log.Fields{
			"error":     err,
			"interface": interfaceName,
			"mtu":       mtu,
		}).Panic("strconv.ParseInt of mtu failed")
	}
	err = netlink.LinkSetMTU(link, int(mtuInt))
	if err != nil {
		log.WithFields(log.Fields{
			"error":     err,
			"interface": interfaceName,
			"mtu":       mtuInt,
		}).Panic("netlink.LinkSetMTU of mtu failed")
	}
	log.WithFields(log.Fields{
		"interface": interfaceName,
		"mtu":       mtuInt,
	}).Debug("set interface MTU")
}

func update_ip_address(link netlink.Link) {
	interfaceName := linkName(link)
	cidr, err := subnet_mask_to_cidr(os.Getenv("new_subnet_mask"))

	newAddr, err := netlink.ParseAddr(fmt.Sprintf("%s/%s", os.Getenv("new_ip_address"), cidr))
	if err != nil {
		log.WithFields(log.Fields{
			"error":           err,
			"interface":       interfaceName,
			"new-ip-address":  os.Getenv("new_ip_address"),
			"new-subnet-mask": os.Getenv("new_subnet_mask"),
		}).Panic("netlink.ParseAddr failed")
	}

	err = netlink.AddrAdd(link, newAddr)
	if err != nil {
		log.WithFields(log.Fields{
			"error":           err,
			"interface":       interfaceName,
			"new-ip-address":  os.Getenv("new_ip_address"),
			"new-subnet-mask": os.Getenv("new_subnet_mask"),
		}).Panic("netlink.AddrAdd failed")
	}
	log.WithFields(log.Fields{
		"interface":       interfaceName,
		"new-ip-address":  os.Getenv("new_ip_address"),
		"new-subnet-mask": os.Getenv("new_subnet_mask"),
	}).Debug("updated interface IP")
}

func linkName(link netlink.Link) string {
	attrs := link.Attrs()
	if attrs == nil {
		panic("")
	}
	return attrs.Name
}

func flush(link netlink.Link, family int) error {
	interfaceName := linkName(link)

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

	log.WithFields(log.Fields{
		"interface": interfaceName,
		"family":    family,
	}).Debug("flushed interface")

	return nil
}
