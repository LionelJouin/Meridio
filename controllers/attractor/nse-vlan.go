package attractor

import (
	"fmt"

	meridiov1alpha1 "github.com/nordix/meridio-operator/api/v1alpha1"
	common "github.com/nordix/meridio-operator/controllers/common"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	nseImage       = "nse-vlan"
	nseEnvItf      = "NSE_VLAN_BASE_IFNAME"
	nseEnvID       = "NSE_VLAN_ID"
	nseEnvSerive   = "NSE_SERVICE_NAME"
	nseEnvPrefixV4 = "NSE_CIDR_PREFIX"
	nseEnvPrefixV6 = "NSE_IPV6_PREFIX"
)

type NseDeployment struct {
	model     *appsv1.Deployment
	attractor *meridiov1alpha1.Attractor
	exec      *common.Executor
	trench    *meridiov1alpha1.Trench
}

func NewNSE(e *common.Executor, attr *meridiov1alpha1.Attractor, t *meridiov1alpha1.Trench) (*NseDeployment, error) {
	nse := &NseDeployment{
		attractor: attr,
		exec:      e,
		trench:    t,
	}
	err := nse.getModel()
	if err != nil {
		return nil, err
	}
	return nse, nil
}

func (i *NseDeployment) getEnvVars(allEnv []corev1.EnvVar) []corev1.EnvVar {
	// if envVars are set in the cr, use the values
	// else return default envVars
	env := []corev1.EnvVar{
		{
			Name:  nseEnvItf,
			Value: i.attractor.Spec.Interface.NSMVlan.BaseInterface,
		},
		{
			Name:  nseEnvID,
			Value: fmt.Sprint(*i.attractor.Spec.Interface.NSMVlan.VlanID),
		},
		{
			Name:  nseEnvSerive,
			Value: common.VlanNtwkSvcName(i.trench),
		},
		{
			Name:  nseEnvPrefixV4,
			Value: i.attractor.Spec.Interface.PrefixIPv4,
		},
		{
			Name:  nseEnvPrefixV6,
			Value: i.attractor.Spec.Interface.PrefixIPv6,
		},
		{
			Name:  "NSE_NAME",
			Value: common.NSEDeploymentName(i.attractor),
		},
	}

	for _, e := range allEnv {
		// append all hard coded envVars
		if e.Name == "SPIFFE_ENDPOINT_SOCKET" ||
			e.Name == "NSE_CONNECT_TO" ||
			e.Name == "NSE_POINT2POINT" {
			env = append(env, e)
		}
	}
	return env
}

func (i *NseDeployment) insertParameters(dep *appsv1.Deployment) *appsv1.Deployment {
	ret := dep.DeepCopy()
	nseVLANDeploymentName := common.NSEDeploymentName(i.attractor)
	ret.ObjectMeta.Name = nseVLANDeploymentName
	ret.ObjectMeta.Namespace = i.attractor.ObjectMeta.Namespace
	ret.ObjectMeta.Labels["app"] = nseVLANDeploymentName
	ret.Spec.Selector.MatchLabels["app"] = nseVLANDeploymentName
	ret.Spec.Template.ObjectMeta.Labels["app"] = nseVLANDeploymentName

	ret.Spec.Template.Spec.ImagePullSecrets = common.GetImagePullSecrets()

	for x, container := range ret.Spec.Template.Spec.Containers {
		switch name := container.Name; name {
		case "nse":
			if container.Image == "" {
				container.Image = fmt.Sprintf("%s/%s/%s:%s", common.Registry, common.OrganizationNsm, nseImage, common.Tag)
				container.ImagePullPolicy = corev1.PullAlways
			}
			if container.StartupProbe == nil {
				container.StartupProbe = common.GetProbe(common.StartUpTimer,
					common.GetProbeCommand(true, "unix:///tmp/listen.on", ""))
			}
			if container.ReadinessProbe == nil {
				container.ReadinessProbe = common.GetProbe(common.ReadinessTimer,
					common.GetProbeCommand(true, "unix:///tmp/listen.on", ""))
			}
			if container.LivenessProbe == nil {
				container.LivenessProbe = common.GetProbe(common.LivenessTimer,
					common.GetProbeCommand(true, "unix:///tmp/listen.on", ""))
			}
			container.Env = i.getEnvVars(container.Env)
		default:
			i.exec.LogError(fmt.Errorf("container %s not expected", name), "get container error")
		}
		ret.Spec.Template.Spec.Containers[x] = container
	}
	return ret
}

func (i *NseDeployment) getModel() error {
	model, err := common.GetDeploymentModel("deployment/nse-vlan.yaml")
	if err != nil {
		return err
	}
	i.model = model
	return nil
}

func (i *NseDeployment) getSelector() client.ObjectKey {
	return client.ObjectKey{
		Namespace: i.attractor.ObjectMeta.Namespace,
		Name:      common.NSEDeploymentName(i.attractor),
	}
}

func (i *NseDeployment) getDesiredStatus() *appsv1.Deployment {
	return i.insertParameters(i.model)

}

// getReconciledDesiredStatus gets the desired status of nse deployment after it's created
// more paramters than what are defined in the model could be added by K8S
func (i *NseDeployment) getReconciledDesiredStatus(cd *appsv1.Deployment) *appsv1.Deployment {
	template := cd.DeepCopy()
	template.Spec.Template.Spec.InitContainers = i.model.Spec.Template.Spec.InitContainers
	template.Spec.Template.Spec.Containers = i.model.Spec.Template.Spec.Containers
	template.Spec.Template.Spec.Volumes = i.model.Spec.Template.Spec.Volumes
	return i.insertParameters(template)
}

func (i *NseDeployment) getCurrentStatus() (*appsv1.Deployment, error) {
	currentStatus := &appsv1.Deployment{}
	selector := i.getSelector()
	err := i.exec.GetObject(selector, currentStatus)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return currentStatus, nil
}

func (i *NseDeployment) getAction() error {
	cs, err := i.getCurrentStatus()
	if err != nil {
		return err
	}
	if cs == nil {
		ds := i.getDesiredStatus()
		if err != nil {
			return err
		}
		i.exec.AddCreateAction(ds)
	} else {
		ds := i.getReconciledDesiredStatus(cs)
		if !equality.Semantic.DeepEqual(ds.Spec, cs.Spec) {
			i.exec.AddUpdateAction(ds)
		}
	}
	return nil
}
