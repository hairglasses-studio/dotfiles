package inventory

// specsCatalog maps ProductLine names to default technical specs.
// Used by quick_list to auto-enrich listings and by import_json to pre-populate
// the Specs column when items have a matching ProductLine.
var specsCatalog = map[string]map[string]string{
	// GPUs
	"RTX 4090": {
		"vram":        "24GB GDDR6X",
		"tdp":         "450W",
		"interface":   "PCIe 4.0 x16",
		"form_factor": "Triple-slot",
	},
	"RTX 4080 Super": {
		"vram":        "16GB GDDR6X",
		"tdp":         "320W",
		"interface":   "PCIe 4.0 x16",
		"form_factor": "Triple-slot",
	},
	"RTX 4080": {
		"vram":        "16GB GDDR6X",
		"tdp":         "320W",
		"interface":   "PCIe 4.0 x16",
		"form_factor": "Triple-slot",
	},
	"RTX 4070 Ti Super": {
		"vram":        "16GB GDDR6X",
		"tdp":         "285W",
		"interface":   "PCIe 4.0 x16",
		"form_factor": "Dual-slot",
	},
	"RTX 4070 Ti": {
		"vram":        "12GB GDDR6X",
		"tdp":         "285W",
		"interface":   "PCIe 4.0 x16",
		"form_factor": "Dual-slot",
	},
	"RTX 4070 Super": {
		"vram":        "12GB GDDR6X",
		"tdp":         "220W",
		"interface":   "PCIe 4.0 x16",
		"form_factor": "Dual-slot",
	},
	"RTX 4070": {
		"vram":        "12GB GDDR6X",
		"tdp":         "200W",
		"interface":   "PCIe 4.0 x16",
		"form_factor": "Dual-slot",
	},
	"RTX 3080": {
		"vram":        "10GB GDDR6X",
		"tdp":         "320W",
		"interface":   "PCIe 4.0 x16",
		"form_factor": "Triple-slot",
	},
	"RTX 3070": {
		"vram":        "8GB GDDR6",
		"tdp":         "220W",
		"interface":   "PCIe 4.0 x16",
		"form_factor": "Dual-slot",
	},
	// NVMe SSDs
	"Sabrent Rocket 5": {
		"interface":   "PCIe 5.0 x4 NVMe",
		"form_factor": "M.2 2280",
		"speed":       "14,000 MB/s read, 12,000 MB/s write",
	},
	"Sabrent Rocket 4 Plus": {
		"interface":   "PCIe 4.0 x4 NVMe",
		"form_factor": "M.2 2280",
		"speed":       "7,100 MB/s read, 6,600 MB/s write",
	},
	"Samsung 990 Pro": {
		"interface":   "PCIe 4.0 x4 NVMe",
		"form_factor": "M.2 2280",
		"speed":       "7,450 MB/s read, 6,900 MB/s write",
	},
	"Samsung 990 Evo": {
		"interface":   "PCIe 5.0 x2 / PCIe 4.0 x4 NVMe",
		"form_factor": "M.2 2280",
		"speed":       "5,000 MB/s read, 4,200 MB/s write",
	},
	"Samsung 980 Pro": {
		"interface":   "PCIe 4.0 x4 NVMe",
		"form_factor": "M.2 2280",
		"speed":       "7,000 MB/s read, 5,100 MB/s write",
	},
	"WD SN850X": {
		"interface":   "PCIe 4.0 x4 NVMe",
		"form_factor": "M.2 2280",
		"speed":       "7,300 MB/s read, 6,600 MB/s write",
	},
	"Crucial T700": {
		"interface":   "PCIe 5.0 x4 NVMe",
		"form_factor": "M.2 2280",
		"speed":       "12,400 MB/s read, 11,800 MB/s write",
	},
	"Crucial T500": {
		"interface":   "PCIe 4.0 x4 NVMe",
		"form_factor": "M.2 2280",
		"speed":       "7,400 MB/s read, 7,000 MB/s write",
	},
	// DDR5 RAM
	"Corsair Vengeance DDR5": {
		"memory":      "DDR5",
		"speed":       "5600 MT/s",
		"form_factor": "DIMM",
	},
	"G.Skill Trident Z5": {
		"memory":      "DDR5",
		"speed":       "6000 MT/s",
		"form_factor": "DIMM",
	},
	"Kingston Fury Beast DDR5": {
		"memory":      "DDR5",
		"speed":       "5600 MT/s",
		"form_factor": "DIMM",
	},
	// DDR4 RAM
	"Corsair Vengeance DDR4": {
		"memory":      "DDR4",
		"speed":       "3200 MT/s",
		"form_factor": "DIMM",
	},
	// Networking
	"Netgear Nighthawk": {
		"interface":   "WiFi 6E",
		"speed":       "AXE7800",
		"form_factor": "Desktop router",
	},
	"UniFi Dream Machine": {
		"interface":   "WiFi 6",
		"speed":       "1 Gbps routing",
		"form_factor": "All-in-one gateway",
	},
	// Thunderbolt
	"CalDigit TS4": {
		"interface":   "Thunderbolt 4",
		"speed":       "40 Gbps",
		"form_factor": "Desktop dock",
	},
	"OWC Thunderbolt Hub": {
		"interface":   "Thunderbolt 4",
		"speed":       "40 Gbps",
		"form_factor": "Hub",
	},
	// CPUs
	"Core i9-14900K": {
		"chipset":     "LGA 1700",
		"tdp":         "253W",
		"speed":       "24 cores (8P+16E), 6.0 GHz boost",
	},
	"Core i7-14700K": {
		"chipset":     "LGA 1700",
		"tdp":         "253W",
		"speed":       "20 cores (8P+12E), 5.6 GHz boost",
	},
	"Ryzen 9 7950X": {
		"chipset":     "AM5",
		"tdp":         "170W",
		"speed":       "16 cores, 5.7 GHz boost",
	},
	"Ryzen 7 7800X3D": {
		"chipset":     "AM5",
		"tdp":         "120W",
		"speed":       "8 cores, 5.0 GHz boost, 96MB L3 cache",
	},
	// PSUs
	"Corsair RM1000x": {
		"capacity":    "1000W",
		"interface":   "ATX 3.0, PCIe 5.0",
		"form_factor": "Full modular ATX",
	},
	"Corsair SF750": {
		"capacity":    "750W",
		"interface":   "ATX, 80+ Platinum",
		"form_factor": "SFX",
	},
	// SATA SSDs
	"Samsung 870 Evo": {
		"interface":   "SATA III 6 Gbps",
		"form_factor": "2.5-inch",
		"speed":       "560 MB/s read, 530 MB/s write",
	},
	"Samsung 870 QVO": {
		"interface":   "SATA III 6 Gbps",
		"form_factor": "2.5-inch",
		"speed":       "560 MB/s read, 530 MB/s write",
		"nand":        "QLC",
	},
	"Samsung 860 Evo": {
		"interface":   "SATA III 6 Gbps",
		"form_factor": "2.5-inch",
		"speed":       "550 MB/s read, 520 MB/s write",
		"nand":        "V-NAND MLC",
	},
	// HDDs
	"Seagate IronWolf Pro": {
		"interface":   "SATA III 6 Gbps",
		"form_factor": "3.5-inch",
		"speed":       "260 MB/s sustained",
		"rpm":         "7200",
		"cache":       "256MB",
	},
	// DDR4 RAM
	"G.Skill Trident Z RGB DDR4": {
		"memory":      "DDR4",
		"speed":       "3600 MT/s",
		"form_factor": "DIMM",
	},
	// Networking
	"TP-Link TL-SX1008": {
		"interface": "10GbE RJ45",
		"speed":     "10 Gbps per port",
		"ports":     "8",
	},
	"TP-Link TX401": {
		"interface":   "PCIe 3.0 x4",
		"speed":       "10 Gbps",
		"form_factor": "PCIe NIC",
	},
}

// LookupSpecs returns default specs for a ProductLine, or nil if not found.
func LookupSpecs(productLine string) map[string]string {
	if specs, ok := specsCatalog[productLine]; ok {
		// Return a copy to avoid mutation
		result := make(map[string]string, len(specs))
		for k, v := range specs {
			result[k] = v
		}
		return result
	}
	return nil
}
