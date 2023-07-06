/*
Copyright 2016 The Kubernetes Authors.

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

// Note: the example only works with the code within the same release/branch.
package utils

import (
	"context"
	"log"
	"path/filepath"
	"strconv"

	"google.golang.org/grpc/resolver"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type EndPoint struct {
	clientset             *kubernetes.Clientset
	NameSpace             string
	EndPointLabelSelector string             //"name=test"
	ResEndPoint           []resolver.Address //[{Addr: "10.42.0.47:22222", ServerName: "", }]
}

// initial clientset
func (k *EndPoint) Init() {
	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	} else {
		log.Fatalln("~/.kube/config not exists! ")
	}

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	// create the clientset
	k.clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
}

// watching specifical endpoint
func (k *EndPoint) Watch(resConn resolver.ClientConn) {
	log.Println("watching Endpoint...")
	watcher, err := k.clientset.CoreV1().Endpoints(k.NameSpace).Watch(context.Background(), metav1.ListOptions{LabelSelector: k.EndPointLabelSelector})
	if err != nil {
		log.Fatalln(err.Error())
	}
	for event := range watcher.ResultChan() {
		item := event.Object.(*corev1.Endpoints)

		switch event.Type {
		case watch.Modified:
			k.ResEndPoint = []resolver.Address{}
			for _, v := range item.Subsets {
				for _, EndPoint := range v.Addresses {
					for _, port := range v.Ports {
						addPort := strconv.Itoa(int(port.Port))
						k.ResEndPoint = append(k.ResEndPoint, resolver.Address{Addr: EndPoint.IP + ":" + addPort})
					}
					// fmt.Printf("Hostname: %s, HostIP: %s \n", EndPoint.Hostname, EndPoint.IP)
				}

			}
			resConn.UpdateState(resolver.State{Addresses: k.ResEndPoint})
			log.Println("EndPoints has been modified: ", k.ResEndPoint)

		case watch.Bookmark:
		case watch.Error:
		case watch.Deleted:
		case watch.Added:
			// fmt.Println(item.GetName())
		}
	}
}

// Geting specifical endpoint
func (k *EndPoint) Get() []resolver.Address {
	log.Println("get EndPoint... ")
	EndPoint, err := k.clientset.CoreV1().Endpoints(k.NameSpace).List(context.Background(), metav1.ListOptions{LabelSelector: k.EndPointLabelSelector})
	if err != nil {
		log.Fatalln(err.Error())
	}
	for _, item := range EndPoint.Items {
		for _, v := range item.Subsets {
			for _, EndPoint := range v.Addresses {
				for _, port := range v.Ports {
					addPort := strconv.Itoa(int(port.Port))
					k.ResEndPoint = append(k.ResEndPoint, resolver.Address{Addr: EndPoint.IP + ":" + addPort})
					// fmt.Printf("Hostname: %s, HostIP: %s \n", EndPoint.Hostname, EndPoint.IP)
				}
			}
		}
	}
	log.Println("List EndPoints: ", k.ResEndPoint)
	return k.ResEndPoint
}
