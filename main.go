package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

type SystemInfo struct {
	Host   HostInfo   `json:"host"`
	CPU    CPUInfo    `json:"cpu"`
	Memory MemoryInfo `json:"memory"`
	Disk   DiskInfo   `json:"disk"`
}

type HostInfo struct {
	Hostname string `json:"hostname"`
	OS       string `json:"os"`
	Platform string `json:"platform"`
	Uptime   uint64 `json:"uptime_seconds"`
}

type CPUInfo struct {
	Cores       int       `json:"cores"`
	ModelName   string    `json:"model_name"`
	UsagePercent []float64 `json:"usage_percent"`
}

type MemoryInfo struct {
	Total       uint64  `json:"total_bytes"`
	Used        uint64  `json:"used_bytes"`
	Free        uint64  `json:"free_bytes"`
	UsedPercent float64 `json:"used_percent"`
}

type DiskInfo struct {
	Total       uint64  `json:"total_bytes"`
	Used        uint64  `json:"used_bytes"`
	Free        uint64  `json:"free_bytes"`
	UsedPercent float64 `json:"used_percent"`
}

func getSystemInfo() (*SystemInfo, error) {
	// Host info
	hostInfo, err := host.Info()
	if err != nil {
		return nil, err
	}

	// CPU info
	cpuInfo, err := cpu.Info()
	if err != nil {
		return nil, err
	}
	cpuPercent, err := cpu.Percent(0, true)
	if err != nil {
		return nil, err
	}

	modelName := ""
	if len(cpuInfo) > 0 {
		modelName = cpuInfo[0].ModelName
	}

	// Memory info
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	// Disk info
	diskInfo, err := disk.Usage("/")
	if err != nil {
		return nil, err
	}

	return &SystemInfo{
		Host: HostInfo{
			Hostname: hostInfo.Hostname,
			OS:       hostInfo.OS,
			Platform: hostInfo.Platform,
			Uptime:   hostInfo.Uptime,
		},
		CPU: CPUInfo{
			Cores:        len(cpuInfo),
			ModelName:    modelName,
			UsagePercent: cpuPercent,
		},
		Memory: MemoryInfo{
			Total:       memInfo.Total,
			Used:        memInfo.Used,
			Free:        memInfo.Free,
			UsedPercent: memInfo.UsedPercent,
		},
		Disk: DiskInfo{
			Total:       diskInfo.Total,
			Used:        diskInfo.Used,
			Free:        diskInfo.Free,
			UsedPercent: diskInfo.UsedPercent,
		},
	}, nil
}

func handleSystemInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	info, err := getSystemInfo()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(info)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func main() {
	http.HandleFunc("/api/system", handleSystemInfo)
	http.HandleFunc("/health", handleHealth)

	log.Println("Server starting on :8088...")
	log.Fatal(http.ListenAndServe(":8088", nil))
}
