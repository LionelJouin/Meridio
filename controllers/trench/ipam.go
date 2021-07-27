package trench

import (
	"fmt"

	meridiov1alpha1 "github.com/nordix/meridio-operator/api/v1alpha1"
	common "github.com/nordix/meridio-operator/controllers/common"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	imageIpam   = "ipam"
	ipamEnvName = "IPAM_PORT"
)

type IpamDeployment struct {
	trench *meridiov1alpha1.Trench
	model  *appsv1.Deployment
}

func NewIPAM(t *meridiov1alpha1.Trench) (*IpamDeployment, error) {
	l := &IpamDeployment{
		trench: t.DeepCopy(),
	}

	// get model
	if err := l.getModel(); err != nil {
		return nil, err
	}
	return l, nil
}

func (i *IpamDeployment) insertParamters(dep *appsv1.Deployment) *appsv1.Deployment {
	// if status ipam deployment parameters are specified in the cr, use those
	// else use the default parameters
	ret := dep.DeepCopy()
	ipamDeploymentName := common.IPAMDeploymentName(i.trench)
	ret.ObjectMeta.Name = ipamDeploymentName
	ret.ObjectMeta.Namespace = i.trench.ObjectMeta.Namespace
	ret.ObjectMeta.Labels["app"] = ipamDeploymentName
	ret.Spec.Selector.MatchLabels["app"] = ipamDeploymentName
	ret.Spec.Template.ObjectMeta.Labels["app"] = ipamDeploymentName
	ret.Spec.Template.Spec.Containers[0].Image = fmt.Sprintf("%s/%s/%s:%s", common.Registry, common.Organization, imageIpam, common.Tag)
	ret.Spec.Template.Spec.Containers[0].ImagePullPolicy = common.PullPolicy
	ret.Spec.Template.Spec.Containers[0].LivenessProbe = common.GetLivenessProbe(i.trench)
	ret.Spec.Template.Spec.Containers[0].ReadinessProbe = common.GetReadinessProbe(i.trench)
	return ret
}

func (i *IpamDeployment) getModel() error {
	model, err := common.GetDeploymentModel("deployment/ipam.yaml")
	if err != nil {
		return err
	}
	i.model = model
	return nil
}

func (i *IpamDeployment) getSelector() client.ObjectKey {
	return client.ObjectKey{
		Namespace: i.trench.ObjectMeta.Namespace,
		Name:      common.IPAMDeploymentName(i.trench),
	}
}

func (i *IpamDeployment) getDesiredStatus() *appsv1.Deployment {
	return i.insertParamters(i.model)
}

// getIpamDeploymentReconciledDesiredStatus gets the desired status of ipam deployment after it's created
// more paramters than what are defined in the model could be added by K8S
func (i *IpamDeployment) getReconciledDesiredStatus(cd *appsv1.Deployment) *appsv1.Deployment {
	return i.insertParamters(cd)
}

func (i *IpamDeployment) getCurrentStatus(e *common.Executor) (*appsv1.Deployment, error) {
	currentStatus := &appsv1.Deployment{}
	selector := i.getSelector()
	err := e.GetObject(selector, currentStatus)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return currentStatus, nil
}

func (i *IpamDeployment) getAction(e *common.Executor) (common.Action, error) {
	elem := common.IPAMDeploymentName(i.trench)
	var action common.Action
	cs, err := i.getCurrentStatus(e)
	if err != nil {
		return nil, err
	}
	if cs == nil {
		ds := i.getDesiredStatus()
		if err != nil {
			return nil, err
		}
		e.LogInfo(fmt.Sprintf("add action: create %s", elem))
		action = common.NewCreateAction(ds, fmt.Sprintf("create %s", elem))
	} else {
		ds := i.getReconciledDesiredStatus(cs)
		if !equality.Semantic.DeepEqual(ds, cs) {
			e.LogInfo(fmt.Sprintf("add action: update %s", elem))
			action = common.NewUpdateAction(ds, fmt.Sprintf("update %s", elem))
		}
	}
	return action, nil
}
