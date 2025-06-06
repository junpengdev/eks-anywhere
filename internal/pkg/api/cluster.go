package api

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

type ClusterFiller func(c *anywherev1.Cluster)

// ClusterToConfigFiller updates the Cluster in the cluster.Config by applying all the fillers.
func ClusterToConfigFiller(fillers ...ClusterFiller) ClusterConfigFiller {
	return func(c *cluster.Config) {
		for _, f := range fillers {
			f(c.Cluster)
		}
	}
}

// JoinClusterConfigFillers creates one single ClusterConfigFiller from a collection of fillers.
func JoinClusterConfigFillers(fillers ...ClusterConfigFiller) ClusterConfigFiller {
	return func(c *cluster.Config) {
		for _, f := range fillers {
			f(c)
		}
	}
}

func WithKubernetesVersion(v anywherev1.KubernetesVersion) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		c.Spec.KubernetesVersion = v
	}
}

// WithLicenseToken sets LicenseToken with the provided token value to use.
func WithLicenseToken(licenseToken string) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		c.Spec.LicenseToken = licenseToken
	}
}

// WithBundlesRef sets BundlesRef with the provided name to use.
func WithBundlesRef(name string, namespace string, apiVersion string) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		c.Spec.BundlesRef = &anywherev1.BundlesRef{Name: name, Namespace: namespace, APIVersion: apiVersion}
	}
}

// WithEksaVersion sets EksaVersion with the provided name to use.
func WithEksaVersion(eksaVersion *anywherev1.EksaVersion) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		c.Spec.EksaVersion = eksaVersion
	}
}

func WithCiliumPolicyEnforcementMode(mode anywherev1.CiliumPolicyEnforcementMode) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		if c.Spec.ClusterNetwork.CNIConfig == nil {
			c.Spec.ClusterNetwork.CNIConfig = &anywherev1.CNIConfig{Cilium: &anywherev1.CiliumConfig{}}
		}
		c.Spec.ClusterNetwork.CNIConfig.Cilium.PolicyEnforcementMode = mode
	}
}

// WithCiliumEgressMasqueradeInterfaces sets the egressMasqueradeInterfaces with the provided interface option to use.
func WithCiliumEgressMasqueradeInterfaces(interfaceName string) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		if c.Spec.ClusterNetwork.CNIConfig == nil {
			c.Spec.ClusterNetwork.CNIConfig = &anywherev1.CNIConfig{Cilium: &anywherev1.CiliumConfig{}}
		}
		c.Spec.ClusterNetwork.CNIConfig.Cilium.EgressMasqueradeInterfaces = interfaceName
	}
}

// WithCiliumSkipUpgrade enables skip upgrade for EKSA Cilium installations.
func WithCiliumSkipUpgrade() ClusterFiller {
	return func(c *anywherev1.Cluster) {
		network := c.Spec.ClusterNetwork
		if network.CNIConfig != nil && network.CNIConfig.Cilium != nil {
			fmt.Println("Enable ciliun skip upgrade")
			network.CNIConfig.Cilium.SkipUpgrade = ptr.Bool(true)
		}
	}
}

// WithCiliumRoutingMode sets the tunnel mode with the provided mode option to use.
func WithCiliumRoutingMode(mode anywherev1.CiliumRoutingMode) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		if c.Spec.ClusterNetwork.CNIConfig == nil {
			c.Spec.ClusterNetwork.CNIConfig = &anywherev1.CNIConfig{Cilium: &anywherev1.CiliumConfig{}}
		}
		c.Spec.ClusterNetwork.CNIConfig.Cilium.RoutingMode = mode
	}
}

func WithClusterNamespace(ns string) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		c.Namespace = ns
	}
}

func WithControlPlaneCount(r int) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		c.Spec.ControlPlaneConfiguration.Count = r
	}
}

func WithControlPlaneEndpointIP(value string) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		c.Spec.ControlPlaneConfiguration.Endpoint.Host = value
	}
}

func WithControlPlaneTaints(taints []corev1.Taint) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		c.Spec.ControlPlaneConfiguration.Taints = taints
	}
}

func WithControlPlaneLabel(key string, val string) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		if c.Spec.ControlPlaneConfiguration.Labels == nil {
			c.Spec.ControlPlaneConfiguration.Labels = map[string]string{}
		}
		c.Spec.ControlPlaneConfiguration.Labels[key] = val
	}
}

// WithControlPlaneAPIServerExtraArgs adds the APIServerExtraArgs to the cluster spec.
func WithControlPlaneAPIServerExtraArgs() ClusterFiller {
	return func(c *anywherev1.Cluster) {
		if c.Spec.ControlPlaneConfiguration.APIServerExtraArgs == nil {
			c.Spec.ControlPlaneConfiguration.APIServerExtraArgs = map[string]string{}
		}
		issuerURL := "https://" + c.Spec.ControlPlaneConfiguration.Endpoint.Host
		c.Spec.ControlPlaneConfiguration.APIServerExtraArgs["service-account-jwks-uri"] = issuerURL + "/openid/v1/jwks"
	}
}

// WithControlPlaneKubeletConfig adds the Kubelet config to the control plane in cluster spec.
func WithControlPlaneKubeletConfig(kc *unstructured.Unstructured) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		c.Spec.ControlPlaneConfiguration.KubeletConfiguration = kc
	}
}

// RemoveAllAPIServerExtraArgs removes all the API server flags from the cluster spec.
func RemoveAllAPIServerExtraArgs() ClusterFiller {
	return func(c *anywherev1.Cluster) {
		for k := range c.Spec.ControlPlaneConfiguration.APIServerExtraArgs {
			delete(c.Spec.ControlPlaneConfiguration.APIServerExtraArgs, k)
		}
	}
}

// WithPodCidr sets an explicit pod CIDR, overriding the provider's default.
func WithPodCidr(podCidr string) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		c.Spec.ClusterNetwork.Pods.CidrBlocks = strings.Split(podCidr, ",")
	}
}

func WithServiceCidr(svcCidr string) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		c.Spec.ClusterNetwork.Services.CidrBlocks = []string{svcCidr}
	}
}

// WithWorkerKubernetesVersion sets the kubernetes version field for the given worker group.
func WithWorkerKubernetesVersion(name string, version *anywherev1.KubernetesVersion) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		pos := -1
		for i, wng := range c.Spec.WorkerNodeGroupConfigurations {
			if wng.Name == name {
				wng.KubernetesVersion = version
				pos = i
				c.Spec.WorkerNodeGroupConfigurations[pos] = wng
				break
			}
		}
		// Append the worker node group if not already found in existing configuration
		if pos == -1 {
			c.Spec.WorkerNodeGroupConfigurations = append(c.Spec.WorkerNodeGroupConfigurations, workerNodeWithKubernetesVersion(name, version))
		}
	}
}

func workerNodeWithKubernetesVersion(name string, version *anywherev1.KubernetesVersion) anywherev1.WorkerNodeGroupConfiguration {
	return anywherev1.WorkerNodeGroupConfiguration{
		Name:              name,
		Count:             ptr.Int(1),
		KubernetesVersion: version,
	}
}

// WithWorkerNodeKubeletConfig adds the Kubelet config to the worker node groups in cluster spec.
func WithWorkerNodeKubeletConfig(kc *unstructured.Unstructured) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		if len(c.Spec.WorkerNodeGroupConfigurations) == 0 {
			c.Spec.WorkerNodeGroupConfigurations = []anywherev1.WorkerNodeGroupConfiguration{{}}
		}
		c.Spec.WorkerNodeGroupConfigurations[0].KubeletConfiguration = kc
	}
}

func WithWorkerNodeCount(r int) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		if len(c.Spec.WorkerNodeGroupConfigurations) == 0 {
			c.Spec.WorkerNodeGroupConfigurations = []anywherev1.WorkerNodeGroupConfiguration{{Count: ptr.Int(0)}}
		}
		c.Spec.WorkerNodeGroupConfigurations[0].Count = &r
	}
}

// WithWorkerNodeAutoScalingConfig adds an autoscaling configuration with a given min and max count.
func WithWorkerNodeAutoScalingConfig(min int, max int) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		if len(c.Spec.WorkerNodeGroupConfigurations) == 0 {
			c.Spec.WorkerNodeGroupConfigurations = []anywherev1.WorkerNodeGroupConfiguration{{Count: ptr.Int(min)}}
		}
		c.Spec.WorkerNodeGroupConfigurations[0].AutoScalingConfiguration = &anywherev1.AutoScalingConfiguration{
			MinCount: min,
			MaxCount: max,
		}
	}
}

func WithOIDCIdentityProviderRef(name string) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		c.Spec.IdentityProviderRefs = append(c.Spec.IdentityProviderRefs,
			anywherev1.Ref{Name: name, Kind: anywherev1.OIDCConfigKind})
	}
}

func WithGitOpsRef(name, kind string) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		c.Spec.GitOpsRef = &anywherev1.Ref{Name: name, Kind: kind}
	}
}

func WithExternalEtcdTopology(count int) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		if c.Spec.ExternalEtcdConfiguration == nil {
			c.Spec.ExternalEtcdConfiguration = &anywherev1.ExternalEtcdConfiguration{}
		}
		c.Spec.ExternalEtcdConfiguration.Count = count
	}
}

func WithEtcdCountIfExternal(count int) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		if c.Spec.ExternalEtcdConfiguration != nil {
			c.Spec.ExternalEtcdConfiguration.Count = count
		}
	}
}

func WithExternalEtcdMachineRef(kind string) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		if c.Spec.ExternalEtcdConfiguration == nil {
			c.Spec.ExternalEtcdConfiguration = &anywherev1.ExternalEtcdConfiguration{}
			c.Spec.ExternalEtcdConfiguration.Count = 1
		}

		if c.Spec.ExternalEtcdConfiguration.MachineGroupRef == nil {
			c.Spec.ExternalEtcdConfiguration.MachineGroupRef = &anywherev1.Ref{}
		}
		c.Spec.ExternalEtcdConfiguration.MachineGroupRef.Kind = kind
		c.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name = providers.GetEtcdNodeName(c.Name)
	}
}

func WithStackedEtcdTopology() ClusterFiller {
	return func(c *anywherev1.Cluster) {
		c.Spec.ExternalEtcdConfiguration = nil
	}
}

func WithProxyConfig(httpProxy, httpsProxy string, noProxy []string) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		if c.Spec.ProxyConfiguration == nil {
			c.Spec.ProxyConfiguration = &anywherev1.ProxyConfiguration{}
		}
		c.Spec.ProxyConfiguration.HttpProxy = httpProxy
		c.Spec.ProxyConfiguration.HttpsProxy = httpProxy
		c.Spec.ProxyConfiguration.NoProxy = noProxy
	}
}

// WithRegistryMirror adds a registry mirror configuration.
func WithRegistryMirror(endpoint, port string, caCert string, authenticate bool, insecureSkipVerify bool, ociNamespaces ...anywherev1.OCINamespace) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		if c.Spec.RegistryMirrorConfiguration == nil {
			c.Spec.RegistryMirrorConfiguration = &anywherev1.RegistryMirrorConfiguration{}
		}
		c.Spec.RegistryMirrorConfiguration.Endpoint = endpoint
		c.Spec.RegistryMirrorConfiguration.Port = port
		c.Spec.RegistryMirrorConfiguration.CACertContent = caCert
		c.Spec.RegistryMirrorConfiguration.Authenticate = authenticate
		c.Spec.RegistryMirrorConfiguration.InsecureSkipVerify = insecureSkipVerify

		if len(ociNamespaces) != 0 {
			c.Spec.RegistryMirrorConfiguration.OCINamespaces = ociNamespaces
		}
	}
}

func WithManagementCluster(name string) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		c.Spec.ManagementCluster.Name = name
	}
}

func WithAWSIamIdentityProviderRef(name string) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		c.Spec.IdentityProviderRefs = append(c.Spec.IdentityProviderRefs,
			anywherev1.Ref{Name: name, Kind: anywherev1.AWSIamConfigKind})
	}
}

func RemoveAllWorkerNodeGroups() ClusterFiller {
	return func(c *anywherev1.Cluster) {
		c.Spec.WorkerNodeGroupConfigurations = make([]anywherev1.WorkerNodeGroupConfiguration, 0)
	}
}

func RemoveWorkerNodeGroup(name string) ClusterFiller {
	logger.Info("removing", "name", name)
	return func(c *anywherev1.Cluster) {
		logger.Info("before deleting", "w", c.Spec.WorkerNodeGroupConfigurations)
		for i, w := range c.Spec.WorkerNodeGroupConfigurations {
			if w.Name == name {
				copy(c.Spec.WorkerNodeGroupConfigurations[i:], c.Spec.WorkerNodeGroupConfigurations[i+1:])
				c.Spec.WorkerNodeGroupConfigurations[len(c.Spec.WorkerNodeGroupConfigurations)-1] = anywherev1.WorkerNodeGroupConfiguration{}
				c.Spec.WorkerNodeGroupConfigurations = c.Spec.WorkerNodeGroupConfigurations[:len(c.Spec.WorkerNodeGroupConfigurations)-1]
				logger.Info("after deleting", "w", c.Spec.WorkerNodeGroupConfigurations)
				return
			}
		}
	}
}

func WithWorkerNodeGroup(name string, fillers ...WorkerNodeGroupFiller) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		var nodeGroup *anywherev1.WorkerNodeGroupConfiguration
		position := -1
		for i, w := range c.Spec.WorkerNodeGroupConfigurations {
			if w.Name == name {
				logger.Info("Updating worker node group", "name", name)
				nodeGroup = &w
				position = i
				break
			}
		}

		if nodeGroup == nil {
			logger.Info("Adding worker node group", "name", name)
			nodeGroup = &anywherev1.WorkerNodeGroupConfiguration{Name: name}
			c.Spec.WorkerNodeGroupConfigurations = append(c.Spec.WorkerNodeGroupConfigurations, *nodeGroup)
			position = len(c.Spec.WorkerNodeGroupConfigurations) - 1
		}

		FillWorkerNodeGroup(nodeGroup, fillers...)
		c.Spec.WorkerNodeGroupConfigurations[position] = *nodeGroup
	}
}

// WithPodIamFiller configures pod IAM config to enable IRSA.
func WithPodIamFiller(issuerURL string) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		c.Spec.PodIAMConfig = &anywherev1.PodIAMConfig{
			ServiceAccountIssuer: issuerURL,
		}
	}
}

// WithEtcdEncryptionFiller configures EtcdEncyption on the cluster.
func WithEtcdEncryptionFiller(kms *anywherev1.KMS, resources []string) ClusterFiller {
	return func(c *anywherev1.Cluster) {
		c.Spec.EtcdEncryption = &[]anywherev1.EtcdEncryption{
			{
				Providers: []anywherev1.EtcdEncryptionProvider{
					{
						KMS: kms,
					},
				},
				Resources: resources,
			},
		}
	}
}

// WithInPlaceUpgradeStrategy configures the UpgradeStrategy on Control-plane and Worker node groups to InPlace.
func WithInPlaceUpgradeStrategy() ClusterFiller {
	return func(c *anywherev1.Cluster) {
		c.Spec.ControlPlaneConfiguration.UpgradeRolloutStrategy = &anywherev1.ControlPlaneUpgradeRolloutStrategy{
			Type: anywherev1.InPlaceStrategyType,
		}
		for idx, wng := range c.Spec.WorkerNodeGroupConfigurations {
			wng.UpgradeRolloutStrategy = &anywherev1.WorkerNodesUpgradeRolloutStrategy{
				Type: anywherev1.InPlaceStrategyType,
			}
			c.Spec.WorkerNodeGroupConfigurations[idx] = wng
		}
	}
}
