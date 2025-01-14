package controller

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/performance"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

const namespace = "vmware"

var (
	prometheusHostPowerState = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "host",
		Name:      "power_state",
		Help:      "poweredOn 1, poweredOff 2, standBy 3, other 0",
	}, []string{"host_name"})
	prometheusHostBoot = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "host",
		Name:      "boot_timestamp_seconds",
		Help:      "Uptime host",
	}, []string{"host_name"})
	prometheusTotalCpu = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "host",
		Name:      "cpu_max",
		Help:      "CPU total",
	}, []string{"host_name"})
	prometheusUsageCpu = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "host",
		Name:      "cpu_usage",
		Help:      "CPU Usage",
	}, []string{"host_name"})
	prometheusHostCPUCores = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "host",
		Name:      "num_cpu",
		Help:      "The number of CPU cores per host",
	}, []string{"host_name"})
	prometheusTotalMem = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "host",
		Name:      "memory_max",
		Help:      "Memory max",
	}, []string{"host_name"})
	prometheusUsageMem = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "host",
		Name:      "memory_usage",
		Help:      "Memory Usage",
	}, []string{"host_name"})
	prometheusTotalDs = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "datastore",
		Name:      "capacity_size",
		Help:      "Datastore total",
	}, []string{"ds_name", "host_name"})
	prometheusUsageDs = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "datastore",
		Name:      "freespace_size",
		Help:      "Datastore free",
	}, []string{"ds_name", "host_name"})
	prometheusProvisionedDs = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "datastore",
		Name:      "provisioned_size",
		Help:      "Datastore provisioned",
	}, []string{"ds_name", "host_name"})
	prometheusVmBoot = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "vm",
		Name:      "boot_timestamp_seconds",
		Help:      "VMWare VM boot time in seconds",
	}, []string{"vm_name", "host_name"})
	prometheusVmCpuAval = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "vm",
		Name:      "cpu_avaleblemhz",
		Help:      "VMWare VM usage CPU",
	}, []string{"vm_name", "host_name"})
	prometheusVmCpuUsage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "vm",
		Name:      "cpu_usagemhz",
		Help:      "VMWare VM usage CPU",
	}, []string{"vm_name", "host_name"})
	prometheusVmNumCpu = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "vm",
		Name:      "num_cpu",
		Help:      "Available number of cores",
	}, []string{"vm_name", "host_name"})
	prometheusVmMemAval = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "vm",
		Name:      "mem_available",
		Help:      "Available memory",
	}, []string{"vm_name", "host_name"})
	prometheusVmMemUsage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "vm",
		Name:      "mem_usage",
		Help:      "Usage memory",
	}, []string{"vm_name", "host_name"})
	prometheusVmNetRec = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "vm",
		Name:      "net_rec",
		Help:      "Usage memory",
	}, []string{"vm_name", "host_name"})
	prometheusVmPowerState = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "vm",
		Name:      "power_state",
		Help:      "poweredOn 1, poweredOff 2, standBy 3, other 0",
	}, []string{"vm_name", "host_name"})
	prometheusVmSwapUsage = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "vm",
		Name:      "mem_swapped_average",
		Help:      "Average VM swap usage",
	}, []string{"vm_name", "host_name"})
	prometheusVmMaxDiskLatency = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "vm",
		Name:      "disk_max_total_latency_latest",
		Help:      "Max VM disk latency",
	}, []string{"vm_name", "host_name"})
)

func totalCpu(hs mo.HostSystem) float64 {
	totalCPU := int64(hs.Summary.Hardware.CpuMhz) * int64(hs.Summary.Hardware.NumCpuCores)
	return float64(totalCPU)
}

func convertTime(vm mo.VirtualMachine) float64 {
	if vm.Summary.Runtime.BootTime == nil {
		return 0
	}
	return float64(vm.Summary.Runtime.BootTime.Unix())
}

type powerState interface {
	String() string
}

type HostSystemPowerState struct {
	*types.HostSystemPowerState
}

func (ps *HostSystemPowerState) String() string {
	return fmt.Sprintf("%s", *ps.HostSystemPowerState)
}

type VirtualMachinePowerState struct {
	*types.VirtualMachinePowerState
}

func (ps *VirtualMachinePowerState) String() string {
	return fmt.Sprintf("%s", *ps.VirtualMachinePowerState)
}

func getPowerState(p powerState) float64 {
	switch p.String() {
	case "poweredOn":
		return 1
	case "poweredOff":
		return 2
	case "standBy":
		return 3
	}
	return 0
}

func dsProvisionedSize(ds mo.Datastore) float64 {
	ps := ds.Summary.Capacity - ds.Summary.FreeSpace + ds.Summary.Uncommitted
	return float64(ps)
}

// Many types returned by the VMWare API contain embedded
// ManagedObjectReference pointers. We have to handle getting the values
// without causing panics if the embedded pointer is nil.
func lookupRefValue(objRef *types.ManagedObjectReference) (*string, bool) {
	if objRef == nil {
		return nil, false
	}
	return &objRef.Value, true
}

func RegistredMetrics() {
	prometheus.MustRegister(
		prometheusHostPowerState,
		prometheusHostBoot,
		prometheusHostCPUCores,
		prometheusTotalCpu,
		prometheusUsageCpu,
		prometheusTotalMem,
		prometheusUsageMem,
		prometheusTotalDs,
		prometheusUsageDs,
		prometheusProvisionedDs,
		prometheusVmBoot,
		prometheusVmCpuAval,
		prometheusVmNumCpu,
		prometheusVmMemAval,
		prometheusVmMemUsage,
		prometheusVmCpuUsage,
		prometheusVmNetRec,
		prometheusVmPowerState,
		prometheusVmSwapUsage,
		prometheusVmMaxDiskLatency,
	)
}

func NewVmwareHostMetrics(host string, username string, password string, logger *log.Logger) {
	ctx := context.Background()
	c, err := NewClient(ctx, host, username, password)
	if err != nil {
		logger.Fatal(err)
	}
	defer c.Logout(ctx)
	m := view.NewManager(c.Client)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"HostSystem"}, true)
	if err != nil {
		logger.Fatal(err)
	}
	defer v.Destroy(ctx)
	var hss []mo.HostSystem
	err = v.Retrieve(ctx, []string{"HostSystem"}, []string{"summary"}, &hss)
	if err != nil {
		logger.Fatal(err)
	}
	for _, hs := range hss {
		hsname := hs.Summary.Config.Name
		ps := HostSystemPowerState{&hs.Summary.Runtime.PowerState}
		prometheusHostPowerState.WithLabelValues(hsname).Set(getPowerState(&ps))
		prometheusHostBoot.WithLabelValues(hsname).Set(float64(hs.Summary.Runtime.BootTime.Unix()))
		prometheusHostCPUCores.WithLabelValues(hsname).Set(float64(hs.Summary.Hardware.NumCpuCores))
		prometheusTotalCpu.WithLabelValues(hsname).Set(totalCpu(hs))
		prometheusUsageCpu.WithLabelValues(hsname).Set(float64(hs.Summary.QuickStats.OverallCpuUsage))
		prometheusTotalMem.WithLabelValues(hsname).Set(float64(hs.Summary.Hardware.MemorySize))
		prometheusUsageMem.WithLabelValues(hsname).Set(float64(hs.Summary.QuickStats.OverallMemoryUsage) * 1024 * 1024)
	}
}

func NewVmwareDsMetrics(host string, username string, password string, logger *log.Logger) {
	ctx := context.Background()
	c, err := NewClient(ctx, host, username, password)
	if err != nil {
		logger.Fatal(err)
	}
	defer c.Logout(ctx)
	m := view.NewManager(c.Client)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"Datastore"}, true)
	if err != nil {
		logger.Fatal(err)
	}
	defer v.Destroy(ctx)
	var dss []mo.Datastore
	err = v.Retrieve(ctx, []string{"Datastore"}, []string{"summary"}, &dss)
	if err != nil {
		logger.Fatal(err)
	}
	for _, ds := range dss {
		dsname := ds.Summary.Name
		prometheusTotalDs.WithLabelValues(dsname, host).Set(float64(ds.Summary.Capacity))
		prometheusUsageDs.WithLabelValues(dsname, host).Set(float64(ds.Summary.FreeSpace))
		prometheusProvisionedDs.WithLabelValues(dsname, host).Set(dsProvisionedSize(ds))
	}
}

// Retrieve performance metrics for VMs via a PerfManager object
func getVmwareVmPerfMetrics(ctx context.Context, c *govmomi.Client, v *view.ContainerView, vms []mo.VirtualMachine, hss []mo.HostSystem, logger *log.Logger) error {
	vmsRefs, err := v.Find(ctx, []string{"VirtualMachine"}, nil)
	if err != nil {
		return err
	}

	// Create a PerfManager
	perfManager := performance.NewManager(c.Client)

	// Create PerfQuerySpec
	spec := types.PerfQuerySpec{
		MaxSample:  1,
		MetricId:   []types.PerfMetricId{{Instance: "*"}},
		IntervalId: int32(20),
	}

	// Query metrics
	sample, err := perfManager.SampleByName(ctx, spec, []string{"disk.maxTotalLatency.latest"}, vmsRefs)
	if err != nil {
		return err
	}

	result, err := perfManager.ToMetricSeries(ctx, sample)
	if err != nil {
		return err
	}

	// Read result
	for _, metric := range result {
		// Get the human readable vm name from the object referenced by vmRef
		var name string
		var vmhost string
		metricval := &metric.Entity.Value // Perf metrics don't use pointers for their references

		for _, vm := range vms {
			vmname := vm.Summary.Config.Name
			vmval, ok := lookupRefValue(vm.Summary.Vm)
			if !ok {
				logger.Warnf("Could not lookup ManagedObjectReference for %v", vmname)
				continue
			}

			if *metricval == *vmval {
				name = vmname
				vmhostval, ok := lookupRefValue(vm.Summary.Runtime.Host)
				if !ok {
					logger.Warnf("Could not lookup host ManagedObjectReference for %v", vmname)
					continue
				}

				for _, host := range hss {
					hostname := host.Summary.Config.Name
					hostval, ok := lookupRefValue(host.Summary.Host)
					if !ok {
						logger.Warnf("Could not lookup ManagedObjectReference for %v", hostname)
					}

					if *vmhostval == *hostval {
						vmhost = hostname
					}
				}
			}
		}

		if name == "" || vmhost == "" {
			logger.Errorf("Could not resolve references for performance metric %v", metric.Entity.Type)
		}
		for _, v := range metric.Value {
			if len(v.Value) != 0 {
				prometheusVmMaxDiskLatency.WithLabelValues(name, vmhost).Set(float64(v.Value[0]))
			}
		}
	}

	return nil
}

func NewVmwareVmMetrics(host string, username string, password string, logger *log.Logger) {
	ctx := context.Background()
	c, err := NewClient(ctx, host, username, password)
	if err != nil {
		logger.Fatal(err)
	}
	defer c.Logout(ctx)
	m := view.NewManager(c.Client)
	v, err := m.CreateContainerView(ctx, c.ServiceContent.RootFolder, []string{"HostSystem", "VirtualMachine"}, true)
	if err != nil {
		logger.Fatal(err)
	}
	defer v.Destroy(ctx)
	var vms []mo.VirtualMachine
	err = v.Retrieve(ctx, []string{"VirtualMachine"}, []string{"summary"}, &vms)
	if err != nil {
		logger.Fatal(err)
	}
	var hss []mo.HostSystem
	err = v.Retrieve(ctx, []string{"HostSystem"}, []string{"summary"}, &hss)
	if err != nil {
		logger.Fatal(err)
	}

	// Get VM metrics
	for _, vm := range vms {
		vmname := vm.Summary.Config.Name

		// get human readable name of the hostsystem the vm is on
		// from the actual object referenced by hostRef
		var vmhost string
		vmval, ok := lookupRefValue(vm.Summary.Runtime.Host)
		if !ok {
			logger.Warnf("Could not lookup ManagedObjectReference for %v", vmhost)
			continue
		}

		for _, host := range hss {
			hostname := host.Summary.Config.Name
			hostval, ok := lookupRefValue(vm.Summary.Runtime.Host)
			if !ok {
				logger.Warnf("Could not lookup ManagedObjectReference for %v", hostname)
				continue
			}

			if *vmval == *hostval {
				vmhost = hostname
			}
		}

		logger.Debugf("VM: %s -- %s", vmname, vmhost)
		if vmhost == "" {
			logger.Warnf("Could not determine host for %v", vmname)
			continue
		}

		ps := VirtualMachinePowerState{&vm.Summary.Runtime.PowerState}

		prometheusVmBoot.WithLabelValues(vmname, vmhost).Set(convertTime(vm))
		prometheusVmCpuAval.WithLabelValues(vmname, vmhost).Set(float64(vm.Summary.Runtime.MaxCpuUsage) * 1000 * 1000)
		prometheusVmCpuUsage.WithLabelValues(vmname, vmhost).Set(float64(vm.Summary.QuickStats.OverallCpuUsage) * 1000 * 1000)
		prometheusVmNumCpu.WithLabelValues(vmname, vmhost).Set(float64(vm.Summary.Config.NumCpu))
		prometheusVmMemAval.WithLabelValues(vmname, vmhost).Set(float64(vm.Summary.Config.MemorySizeMB))
		prometheusVmMemUsage.WithLabelValues(vmname, vmhost).Set(float64(vm.Summary.QuickStats.GuestMemoryUsage) * 1024 * 1024)
		prometheusVmPowerState.WithLabelValues(vmname, vmhost).Set(float64(getPowerState(&ps)))
		prometheusVmSwapUsage.WithLabelValues(vmname, vmhost).Set(float64(vm.Summary.QuickStats.SwappedMemory))
	}

	// Get VM performance metrics
	err = getVmwareVmPerfMetrics(ctx, c, v, vms, hss, logger)
	if err != nil {
		logger.Fatal(err)
	}
}
