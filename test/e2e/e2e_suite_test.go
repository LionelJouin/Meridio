/*
Copyright (c) 2021-2022 Nordix Foundation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package e2e_test

import (
	"bytes"
	"context"
	"flag"
	"os/exec"
	"testing"
	"time"

	"github.com/nordix/meridio/test/e2e/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	clientset *kubernetes.Clientset

	trafficGeneratorHost *utils.TrafficGeneratorHost
	trafficGenerator     utils.TrafficGenerator

	numberOfTargetA int
	numberOfTargetB int

	config *e2eTestConfiguration
)

type e2eTestConfiguration struct {
	trafficGeneratorCMD string
	script              string

	k8sNamespace              string
	targetADeploymentName     string
	trenchA                   string
	attractorA1               string
	conduitA1                 string
	streamAI                  string
	streamAII                 string
	flowAZTcp                 string
	flowAZTcpDestinationPort0 int
	flowAZUdp                 string
	flowAZUdpDestinationPort0 int
	flowAXTcp                 string
	flowAXTcpDestinationPort0 int
	vip1V4                    string
	vip1V6                    string
	targetBDeploymentName     string
	trenchB                   string
	conduitB1                 string
	streamBI                  string
	vip2V4                    string
	vip2V6                    string

	statelessLbFeDeploymentName string
	ipFamily                    string
}

const (
	timeout  = time.Minute * 3
	interval = time.Second * 2
)

func init() {
	config = &e2eTestConfiguration{}
	flag.StringVar(&config.trafficGeneratorCMD, "traffic-generator-cmd", "", "Command to use to connect to the traffic generator. All occurences of '{trench}' will be replaced with the trench name.")
	flag.StringVar(&config.script, "script", "", "Path + script used by the e2e tests")

	flag.StringVar(&config.k8sNamespace, "k8s-namespace", "", "Name of the namespace")
	flag.StringVar(&config.targetADeploymentName, "target-a-deployment-name", "", "Name of the namespace")
	flag.StringVar(&config.trenchA, "trench-a", "", "Name of the trench")
	flag.StringVar(&config.attractorA1, "attractor-a-1", "", "Name of the attractor")
	flag.StringVar(&config.conduitA1, "conduit-a-1", "", "Name of the conduit")
	flag.StringVar(&config.streamAI, "stream-a-I", "", "Name of the stream")
	flag.StringVar(&config.streamAII, "stream-a-II", "", "Name of the stream")
	flag.StringVar(&config.flowAZTcp, "flow-a-z-tcp", "", "Name of the flow")
	flag.IntVar(&config.flowAZTcpDestinationPort0, "flow-a-z-tcp-destination-port-0", 4000, "Destination port 0")
	flag.StringVar(&config.flowAZUdp, "flow-a-z-udp", "", "Name of the flow")
	flag.IntVar(&config.flowAZUdpDestinationPort0, "flow-a-z-udp-destination-port-0", 4000, "Destination port 0")
	flag.StringVar(&config.flowAXTcp, "flow-a-x-tcp", "", "Name of the flow")
	flag.IntVar(&config.flowAXTcpDestinationPort0, "flow-a-x-tcp-destination-port-0", 4000, "Destination port 0")
	flag.StringVar(&config.vip1V4, "vip-1-v4", "", "Address of the vip v4 number 1")
	flag.StringVar(&config.vip1V6, "vip-1-v6", "", "Address of the vip v6 number 1")
	flag.StringVar(&config.targetBDeploymentName, "target-b-deployment-name", "", "Name of the target deployment")
	flag.StringVar(&config.trenchB, "trench-b", "", "Name of the trench")
	flag.StringVar(&config.conduitB1, "conduit-b-1", "", "Name of the conduit")
	flag.StringVar(&config.streamBI, "stream-b-I", "", "Name of the stream")
	flag.StringVar(&config.vip2V4, "vip-2-v4", "", "Address of the vip v4 number 2")
	flag.StringVar(&config.vip2V6, "vip-2-v6", "", "Address of the vip v6 number 2")

	flag.StringVar(&config.statelessLbFeDeploymentName, "stateless-lb-fe-deployment-name", "", "Name of stateless-lb-fe deployment in trench-a")
	flag.StringVar(&config.ipFamily, "ip-family", "", "IP Family")
}

func TestE2e(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2e Suite")
}

var _ = BeforeSuite(func() {
	var err error
	clientset, err = utils.GetClientSet()
	Expect(err).ToNot(HaveOccurred())

	trafficGeneratorHost = &utils.TrafficGeneratorHost{
		TrafficGeneratorCommand: config.trafficGeneratorCMD,
	}
	trafficGenerator = &utils.MConnect{
		NConn: 400,
	}

	cmd := exec.Command(config.script, "init")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	Expect(stderr.String()).To(BeEmpty())
	Expect(err).ToNot(HaveOccurred())

	deploymentTargetA, err := clientset.AppsV1().Deployments(config.k8sNamespace).Get(context.Background(), config.targetADeploymentName, metav1.GetOptions{})
	Expect(err).ToNot(HaveOccurred())
	numberOfTargetA = int(*deploymentTargetA.Spec.Replicas)

	deploymentTargetB, err := clientset.AppsV1().Deployments(config.k8sNamespace).Get(context.Background(), config.targetADeploymentName, metav1.GetOptions{})
	Expect(err).ToNot(HaveOccurred())
	numberOfTargetB = int(*deploymentTargetB.Spec.Replicas)
})

var _ = AfterSuite(func() {
	cmd := exec.Command(config.script, "end")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	Expect(stderr.String()).To(BeEmpty())
	Expect(err).ToNot(HaveOccurred())
})
