// Copyright © 2019 Banzai Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package portforward

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"

	"emperror.dev/errors"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"sigs.k8s.io/controller-runtime/pkg/client"

	k8sclient "github.com/banzaicloud/backyards-cli/pkg/k8s/client"
)

type Portforward struct {
	namespace  string
	podname    string
	localPort  int
	remotePort int
	url        *url.URL

	stopChannel  chan struct{}
	readyChannel chan struct{}

	config *rest.Config
}

func New(k8sClient k8sclient.Client, config *rest.Config, matchLabels map[string]string, namespace string, localPort, remotePort int) (*Portforward, error) {
	var pods v1.PodList
	err := k8sClient.List(context.Background(), &pods, client.InNamespace(namespace), client.MatchingLabels(matchLabels))
	if err != nil {
		return nil, errors.WrapIfWithDetails(err, "could not list pods", "namespace", namespace)
	}

	podName := ""
	for _, pod := range pods.Items {
		if pod.Status.Phase == v1.PodRunning {
			podName = pod.Name
			break
		}
	}

	if podName == "" {
		return nil, errors.NewWithDetails("no running pods found", "matchLabels", matchLabels, "namespace", namespace)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.WrapIf(err, "could not get k8s clientset")
	}

	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(namespace).
		Name(podName).
		SubResource("portforward")

	if localPort == 0 {
		localPort, err = getEphemeralPort()
		if err != nil {
			return nil, errors.WrapIf(err, "could not get ephemeral port")
		}
	}

	return &Portforward{
		namespace:  namespace,
		podname:    podName,
		localPort:  localPort,
		remotePort: remotePort,
		url:        req.URL(),

		stopChannel:  make(chan struct{}, 1),
		readyChannel: make(chan struct{}),

		config: config,
	}, nil
}

func (pf *Portforward) Stop() {
	if pf.stopChannel != nil {
		close(pf.stopChannel)
	}
}

func (pf *Portforward) WaitForStop() {
	<-pf.stopChannel
}

// GetURL returns the URL for the port-forward connection
func (pf *Portforward) GetURL(path string) string {
	return fmt.Sprintf("http://127.0.0.1:%d%s", pf.localPort, path)
}

func (pf *Portforward) Run() error {
	failure := make(chan error)

	go func() {
		if err := pf.run(); err != nil {
			failure <- err
		}

		select {
		case <-pf.stopChannel:
			// stopCh was closed, do nothing
		default:
			// pf.run() returned for some other reason, close stopCh
			pf.Stop()
		}
	}()

	select {
	case <-pf.readyChannel:
		log.Debug("port forward initialized successfully")
	case err := <-failure:
		err = errors.WrapIf(err, "port forward failed")
		return err
	}

	return nil
}

func (pf *Portforward) run() error {
	var err error

	transport, upgrader, err := spdy.RoundTripperFor(pf.config)
	if err != nil {
		return errors.WrapIf(err, "could not initialize round tripper")
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", pf.url)
	fw, err := portforward.New(dialer, []string{fmt.Sprintf("%d:%d", pf.localPort, pf.remotePort)}, pf.stopChannel, pf.readyChannel, ioutil.Discard, ioutil.Discard)
	if err != nil {
		return errors.WrapIf(err, "could not create port forwarder")
	}

	err = fw.ForwardPorts()
	if err != nil {
		return errors.WrapIf(err, "could not forward port")
	}

	return nil
}

// getEphemeralPort selects a port for the port-forwarding
// It binds to a free ephemeral port and returns the port number
func getEphemeralPort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, errors.WrapIf(err, "could not listen on port zero")
	}

	defer listener.Close()

	tcpAddr, ok := listener.Addr().(*net.TCPAddr)
	if !ok {
		return 0, errors.NewWithDetails("invalid listen address", "address", listener.Addr())
	}

	return tcpAddr.Port, nil
}
