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
	"context"
	"fmt"

	"github.com/nordix/meridio/test/e2e/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Scaling", func() {

	Context("With one trench containing a stream with 2 VIP addresses and 4 target pods running", func() {

		var (
			replicas int
			scale    *autoscalingv1.Scale
		)

		BeforeEach(func() {
			replicas = numberOfTargetA
			scale = &autoscalingv1.Scale{
				ObjectMeta: metav1.ObjectMeta{
					Name:      config.targetADeploymentName,
					Namespace: config.k8sNamespace,
				},
				Spec: autoscalingv1.ScaleSpec{
					Replicas: int32(replicas),
				},
			}
		})

		JustBeforeEach(func() {
			// scale
			scale.Spec.Replicas = int32(replicas)
			_, err := clientset.AppsV1().Deployments(config.k8sNamespace).UpdateScale(context.Background(), config.targetADeploymentName, scale, metav1.UpdateOptions{})
			Expect(err).NotTo(HaveOccurred())
			// wait for all targets to be in Running mode
			Eventually(func() bool {
				deployment, err := clientset.AppsV1().Deployments(config.k8sNamespace).Get(context.Background(), config.targetADeploymentName, metav1.GetOptions{})
				if err != nil {
					return false
				}
				listOptions := metav1.ListOptions{
					LabelSelector: fmt.Sprintf("app=%s", config.targetADeploymentName),
				}
				pods, err := clientset.CoreV1().Pods(config.k8sNamespace).List(context.Background(), listOptions)
				if err != nil {
					return false
				}
				return len(pods.Items) == int(deployment.Status.Replicas) && deployment.Status.ReadyReplicas == deployment.Status.Replicas
			}, timeout, interval).Should(BeTrue())
			// wait for all identifiers to be in NFQLB
			listOptions := metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app=%s", config.statelessLbFeDeploymentName),
			}
			pods, err := clientset.CoreV1().Pods(config.k8sNamespace).List(context.Background(), listOptions)
			Expect(err).NotTo(HaveOccurred())
			Eventually(func() bool {
				for _, pod := range pods.Items {
					nfqlbOutput, err := utils.PodExec(&pod, "load-balancer", []string{"nfqlb", "show", fmt.Sprintf("--shm=tshm-%v", config.streamAI)})
					Expect(err).NotTo(HaveOccurred())
					if utils.ParseNFQLB(nfqlbOutput) != int(scale.Spec.Replicas) {
						return false
					}
				}
				return true
			}, timeout, interval).Should(BeTrue())
		})

		AfterEach(func() {
			// scale
			scale.Spec.Replicas = int32(numberOfTargetA)
			_, err := clientset.AppsV1().Deployments(config.k8sNamespace).UpdateScale(context.Background(), config.targetADeploymentName, scale, metav1.UpdateOptions{})
			Expect(err).NotTo(HaveOccurred())
			// wait for all targets to be in Running mode
			Eventually(func() bool {
				deployment, err := clientset.AppsV1().Deployments(config.k8sNamespace).Get(context.Background(), config.targetADeploymentName, metav1.GetOptions{})
				if err != nil {
					return false
				}
				listOptions := metav1.ListOptions{
					LabelSelector: fmt.Sprintf("app=%s", config.targetADeploymentName),
				}
				pods, err := clientset.CoreV1().Pods(config.k8sNamespace).List(context.Background(), listOptions)
				if err != nil {
					return false
				}
				return len(pods.Items) == int(deployment.Status.Replicas) && deployment.Status.ReadyReplicas == deployment.Status.Replicas
			}, timeout, interval).Should(BeTrue())
			// wait for all identifiers to be in NFQLB
			listOptions := metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app=%s", config.statelessLbFeDeploymentName),
			}
			pods, err := clientset.CoreV1().Pods(config.k8sNamespace).List(context.Background(), listOptions)
			Expect(err).NotTo(HaveOccurred())
			Eventually(func() bool {
				for _, pod := range pods.Items {
					nfqlbOutput, err := utils.PodExec(&pod, "load-balancer", []string{"nfqlb", "show", fmt.Sprintf("--shm=tshm-%v", config.streamAI)})
					Expect(err).NotTo(HaveOccurred())
					if utils.ParseNFQLB(nfqlbOutput) != int(scale.Spec.Replicas) {
						return false
					}
				}
				return true
			}, timeout, interval).Should(BeTrue())
		})

		When("scaling targets down by 1", func() {
			BeforeEach(func() {
				replicas = numberOfTargetA - 1
			})
			It("should receive the traffic correctly", func() {
				if !utils.IsIPv6(config.ipFamily) { // Don't send traffic with IPv4 if the tests are only IPv6
					By("Checking if all targets have receive ipv4 traffic with no traffic interruption (no lost connection)")
					lastingConnections, lostConnections := trafficGeneratorHost.SendTraffic(trafficGenerator, config.trenchA, config.k8sNamespace, utils.VIPPort(config.vip1V4, config.flowAZTcpDestinationPort0), "tcp")
					Expect(lostConnections).To(Equal(0))
					Expect(len(lastingConnections)).To(Equal(replicas))
				}
				if !utils.IsIPv4(config.ipFamily) { // Don't send traffic with IPv6 if the tests are only IPv4
					By("Checking if all targets have receive ipv6 traffic with no traffic interruption (no lost connection)")
					lastingConnections, lostConnections := trafficGeneratorHost.SendTraffic(trafficGenerator, config.trenchA, config.k8sNamespace, utils.VIPPort(config.vip1V6, config.flowAZTcpDestinationPort0), "tcp")
					Expect(lostConnections).To(Equal(0))
					Expect(len(lastingConnections)).To(Equal(replicas))
				}
			})
		})

		When("scaling targets up by 1", func() {
			BeforeEach(func() {
				replicas = numberOfTargetA + 1
			})
			It("should receive the traffic correctly", func() {
				if !utils.IsIPv6(config.ipFamily) { // Don't send traffic with IPv4 if the tests are only IPv6
					By("Checking if all targets have receive ipv4 traffic with no traffic interruption (no lost connection)")
					lastingConnections, lostConnections := trafficGeneratorHost.SendTraffic(trafficGenerator, config.trenchA, config.k8sNamespace, utils.VIPPort(config.vip1V4, config.flowAZTcpDestinationPort0), "tcp")
					Expect(lostConnections).To(Equal(0))
					Expect(len(lastingConnections)).To(Equal(replicas))
				}
				if !utils.IsIPv4(config.ipFamily) { // Don't send traffic with IPv6 if the tests are only IPv4
					By("Checking if all targets have receive ipv6 traffic with no traffic interruption (no lost connection)")
					lastingConnections, lostConnections := trafficGeneratorHost.SendTraffic(trafficGenerator, config.trenchA, config.k8sNamespace, utils.VIPPort(config.vip1V6, config.flowAZTcpDestinationPort0), "tcp")
					Expect(lostConnections).To(Equal(0))
					Expect(len(lastingConnections)).To(Equal(replicas))
				}
			})
		})

	})

})
