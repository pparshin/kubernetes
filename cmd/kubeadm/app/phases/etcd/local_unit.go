package etcd

import (
	"fmt"

	"github.com/pkg/errors"
	"k8s.io/klog/v2"

	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	etcdutil "k8s.io/kubernetes/cmd/kubeadm/app/util/etcd"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/initsystem"
)

func getServiceUnitSpec(cfg *kubeadmapi.ClusterConfiguration, endpoint *kubeadmapi.APIEndpoint, nodeName string, initialCluster []etcdutil.Member) initsystem.UnitSpec {
	return initsystem.UnitSpec{
		Description:   "etcd",
		Documentation: "https://github.com/coreos",
		Service: initsystem.UnitService{
			ExecStartCmd: getEtcdCommand(cfg, endpoint, nodeName, initialCluster),
			Restart:      initsystem.UnitRestartAlways,
			RestartSec:   initsystem.DefaultUnitRestartSec,
		},
		Install: initsystem.UnitInstall{
			WantedBy: initsystem.MultiUserTarget,
		},
	}
}

// CreateServiceUnitFile creates the requested service unit file.
func CreateServiceUnitFile(unitsDir string, nodeName string, cfg *kubeadmapi.ClusterConfiguration, endpoint *kubeadmapi.APIEndpoint) error {
	if cfg.Etcd.External != nil {
		return errors.New("etcd unit file cannot be generated for cluster using external etcd")
	}

	component := kubeadmconstants.Etcd
	spec := getServiceUnitSpec(cfg, endpoint, nodeName, []etcdutil.Member{})

	if err := initsystem.WriteUnitToDisk(component, unitsDir, spec); err != nil {
		return errors.Wrap(err, "failed to create service unit file for etcd")
	}

	klog.V(1).Infof("[control-plane] wrote service unit for etcd to %q\n", kubeadmconstants.GetSystemUnitFilepath(component, unitsDir))

	return nil
}

// RunStackedEtcdService will write local etcd service unit file
// for an additional etcd member that is joining an existing local/stacked etcd cluster and run the service.
// Other members of the etcd cluster will be notified of the joining node in beforehand as well.
func RunStackedEtcdService(unitsDir, nodeName string, cfg *kubeadmapi.ClusterConfiguration, endpoint *kubeadmapi.APIEndpoint, isDryRun bool, certificatesDir string) error {
	klog.V(1).Info("Creating client that connects to etcd cluster")
	etcdClientAddress := etcdutil.GetClientURL(endpoint)
	etcdClient, err := etcdutil.NewFromEndpoint(etcdClientAddress, certificatesDir)
	if err != nil {
		return err
	}

	etcdPeerAddress := etcdutil.GetPeerURL(endpoint)

	var cluster []etcdutil.Member
	if isDryRun {
		fmt.Printf("[dryrun] Would add etcd member: %s\n", etcdPeerAddress)
	} else {
		klog.V(1).Infof("[etcd] Adding etcd member: %s", etcdPeerAddress)
		cluster, err = etcdClient.AddMember(nodeName, etcdPeerAddress)
		if err != nil {
			return err
		}
		fmt.Println("[etcd] Announced new etcd member joining to the existing etcd cluster")
		klog.V(1).Infof("Updated etcd member list: %v", cluster)
	}

	component := kubeadmconstants.Etcd

	fmt.Printf("[etcd] Creating unix service unit file for %q\n", component)
	spec := getServiceUnitSpec(cfg, endpoint, nodeName, cluster)

	if err := initsystem.WriteUnitToDisk(component, unitsDir, spec); err != nil {
		return errors.Wrap(err, "[etcd] failed to create service unit file")
	}

	if isDryRun {
		fmt.Println("[dryrun] Would wait for the new etcd member to join the cluster")
		return nil
	}

	err = RunService()
	if err != nil {
		return err
	}

	fmt.Printf("[etcd] Waiting for the new etcd member to join the cluster. This can take up to %v\n", etcdHealthyCheckInterval*etcdHealthyCheckRetries)
	if _, err := etcdClient.WaitForClusterAvailable(etcdHealthyCheckRetries, etcdHealthyCheckInterval); err != nil {
		return err
	}

	return nil
}

// RunService runs etcd as system daemon.
func RunService() error {
	sys, err := initsystem.GetInitSystem()
	if err != nil {
		return err
	}

	return sys.ServiceStart(kubeadmconstants.Etcd)
}
