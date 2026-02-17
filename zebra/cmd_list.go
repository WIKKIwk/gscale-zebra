package main

import (
	"fmt"
	"sort"
)

func runList() error {
	printers, err := FindUSBLPPrinters()
	if err != nil {
		return err
	}
	if len(printers) == 0 {
		fmt.Println("USB printer topilmadi.")
		return nil
	}

	sort.Slice(printers, func(i, j int) bool {
		return printers[i].DevicePath < printers[j].DevicePath
	})

	fmt.Printf("Topilgan printerlar: %d\n", len(printers))
	for i, p := range printers {
		z := "no"
		if p.IsZebra() {
			z = "yes"
		}
		fmt.Printf("%d) %s\n", i+1, p.DevicePath)
		fmt.Printf("   vendor/product: %s:%s\n", p.VendorID, p.ProductID)
		fmt.Printf("   manufacturer : %s\n", safeStr(p.Manufacturer, "-"))
		fmt.Printf("   product      : %s\n", safeStr(p.Product, "-"))
		fmt.Printf("   serial       : %s\n", safeStr(p.Serial, "-"))
		fmt.Printf("   bus/dev      : %s/%s\n", safeStr(p.BusNum, "-"), safeStr(p.DevNum, "-"))
		fmt.Printf("   zebra        : %s\n", z)
	}
	return nil
}
