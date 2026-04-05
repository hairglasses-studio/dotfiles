package main

import (
	"fmt"
	"net"
	"strings"
	"time"

	"go.evanpurkhiser.com/prolink"
)

// Known Pioneer DJ / AlphaTheta MAC address prefixes (OUI)
var pioneerOUIPrefixes = []string{
	"c8:3d:fc", // AlphaTheta Corporation (Pioneer DJ parent company)
	"70:56:81", // Pioneer Corporation
	"00:e0:36", // Pioneer Corporation
	"00:11:a0", // Pioneer Corporation
	"ac:3a:7a", // Pioneer Corporation
}

func main() {
	fmt.Println("========================================")
	fmt.Println("  Pro DJ Link Network Test (XDJ-1000)")
	fmt.Println("========================================")
	fmt.Println("")

	// Step 1: Check UDP port availability
	fmt.Println("[1/4] Checking UDP port availability...")
	checkPorts()

	// Step 2: Check for Pioneer devices via network interfaces
	fmt.Println("")
	fmt.Println("[2/4] Scanning for Pioneer devices...")
	scanNetwork()

	// Step 3: Connect to Pro DJ Link
	fmt.Println("")
	fmt.Println("[3/4] Connecting to Pro DJ Link network...")
	fmt.Println("      (Close Rekordbox if running)")
	fmt.Println("")

	network, err := prolink.Connect()
	if err != nil {
		fmt.Printf("      ERROR: %v\n", err)
		printTroubleshooting()
		return
	}
	fmt.Println("      Connected to network layer")

	// Try auto-configure
	fmt.Println("      Auto-configuring virtual CDJ ID...")
	if err := network.AutoConfigure(5 * time.Second); err != nil {
		fmt.Printf("      WARNING: Auto-configure failed: %v\n", err)
		fmt.Println("      (Normal if no CDJs broadcasting yet)")
	} else {
		fmt.Println("      Auto-configure successful!")
	}

	// Step 4: Wait for device discovery
	fmt.Println("")
	fmt.Println("[4/4] Waiting 10 seconds for device announcements...")

	dm := network.DeviceManager()

	// Listen for devices
	dm.OnDeviceAdded("test", prolink.DeviceListenerFunc(func(dev *prolink.Device) {
		fmt.Printf("      FOUND: %s (ID: %d, Type: %s, IP: %s)\n",
			dev.Name, dev.ID, dev.Type.String(), dev.IP.String())
	}))

	time.Sleep(10 * time.Second)

	// Results
	devices := dm.ActiveDevices()
	fmt.Println("")
	fmt.Println("========================================")
	fmt.Printf("  Results: %d device(s) discovered\n", len(devices))
	fmt.Println("========================================")

	if len(devices) == 0 {
		fmt.Println("")
		printTroubleshooting()
	} else {
		fmt.Println("")
		for _, dev := range devices {
			fmt.Printf("  [%d] %s\n", dev.ID, dev.Name)
			fmt.Printf("      Type: %s\n", dev.Type.String())
			fmt.Printf("      IP:   %s\n", dev.IP.String())
			fmt.Printf("      MAC:  %s\n", dev.MacAddr.String())
			fmt.Println("")
		}
		fmt.Println("SUCCESS! Pro DJ Link is working.")
	}
}

func checkPorts() {
	ports := []int{50000, 50001, 50002}
	allAvailable := true

	for _, port := range ports {
		addr := fmt.Sprintf(":%d", port)
		conn, err := net.ListenPacket("udp", addr)
		if err != nil {
			fmt.Printf("      UDP %d: BLOCKED (likely Rekordbox)\n", port)
			allAvailable = false
		} else {
			fmt.Printf("      UDP %d: Available\n", port)
			conn.Close()
		}
	}

	if !allAvailable {
		fmt.Println("")
		fmt.Println("      To free ports, run: taskkill /IM rekordbox.exe /F")
	}
}

func scanNetwork() {
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Printf("      Error getting interfaces: %v\n", err)
		return
	}

	foundPioneer := false
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok || ipNet.IP.To4() == nil {
				continue
			}

			mac := strings.ToLower(iface.HardwareAddr.String())
			for _, prefix := range pioneerOUIPrefixes {
				if strings.HasPrefix(mac, strings.ToLower(strings.ReplaceAll(prefix, ":", "-"))) {
					fmt.Printf("      FOUND: %s (MAC: %s) - Pioneer Device!\n", ipNet.IP.String(), iface.HardwareAddr.String())
					foundPioneer = true
				}
			}
		}
	}

	if !foundPioneer {
		fmt.Println("      No Pioneer devices in local interfaces")
		fmt.Println("      (XDJ should appear via Pro DJ Link discovery below)")
	}
}

func printTroubleshooting() {
	fmt.Println("TROUBLESHOOTING:")
	fmt.Println("")
	fmt.Println("1. CLOSE REKORDBOX")
	fmt.Println("   Rekordbox blocks UDP port 50000. Close it first:")
	fmt.Println("   > taskkill /IM rekordbox.exe /F")
	fmt.Println("")
	fmt.Println("2. CHECK WINDOWS FIREWALL")
	fmt.Println("   Run as Administrator:")
	fmt.Println("   > New-NetFirewallRule -DisplayName 'Pro DJ Link UDP 50000' -Direction Inbound -Protocol UDP -LocalPort 50000 -Action Allow")
	fmt.Println("   > New-NetFirewallRule -DisplayName 'Pro DJ Link UDP 50001' -Direction Inbound -Protocol UDP -LocalPort 50001 -Action Allow")
	fmt.Println("   > New-NetFirewallRule -DisplayName 'Pro DJ Link UDP 50002' -Direction Inbound -Protocol UDP -LocalPort 50002 -Action Allow")
	fmt.Println("")
	fmt.Println("3. XDJ CHECKLIST")
	fmt.Println("   [ ] XDJ powered on")
	fmt.Println("   [ ] Connected via Ethernet (same switch/router as PC)")
	fmt.Println("   [ ] USB drive inserted with Rekordbox library")
	fmt.Println("   [ ] Track loaded on the deck")
	fmt.Println("   [ ] Link indicator lit on XDJ display")
	fmt.Println("")
	fmt.Println("4. VERIFY XDJ IP")
	fmt.Println("   Check your router's device list or XDJ network settings")
	fmt.Println("   XDJ MAC prefix: c8:3d:fc (AlphaTheta)")
	fmt.Println("   Try: ping <XDJ-IP-ADDRESS>")
}
