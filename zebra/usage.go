package main

import "fmt"

func printUsage() {
	fmt.Println("Zebra USB tool (gscale-zebra/zebra)")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  zebra list")
	fmt.Println("  zebra status [--device /dev/usb/lp0]")
	fmt.Println("  zebra settings [--device /dev/usb/lp0] [--key print.width --key label.length]")
	fmt.Println("  zebra setvar --device /dev/usb/lp0 --key ezpl.print_width --value 832 [--save=true]")
	fmt.Println("  zebra print-test [--device /dev/usb/lp0] [--message TEXT] [--copies 1] [--dry-run]")
	fmt.Println("  zebra epc-test [--device /dev/usb/lp0] [--epc HEX] [--feed] [--print-human] [--auto-tune=true] [--profile-init=true] [--label-tries 1] [--error-handling none] [--read-power 30] [--write-power 30] [--send]")
	fmt.Println("  zebra calibrate [--device /dev/usb/lp0] [--dry-run] [--save=true]")
	fmt.Println("  zebra self-check [--device /dev/usb/lp0] [--print]")
}
