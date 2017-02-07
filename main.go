package main

import (
	"log"
	"net/http"
	"os/exec"
	"strings"

	albcontroller "git.tm.tmcs/kubernetes/alb-ingress/pkg/cmd/controller"
	"git.tm.tmcs/kubernetes/alb-ingress/pkg/config"
	"k8s.io/ingress/core/pkg/ingress"
	"k8s.io/ingress/core/pkg/ingress/controller"
	"k8s.io/ingress/core/pkg/ingress/defaults"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
)

type ALBController struct {
	ingress.Controller
	kubeClient *internalclientset.Clientset
}

func main() {
	ac := newALBController()
	ic := controller.NewIngressController(ac)

	// We need a kubeClient to look up the nginx daemonset IPs
	ac.kubeClient = albcontroller.KubeClient()

	defer func() {
		log.Printf("Shutting down ingress controller...")
		ic.Stop()
	}()
	ic.Start()
}

func newALBController() *ALBController {
	return &ALBController{}
}

func (ac ALBController) SetConfig(cfgMap *api.ConfigMap) {
	log.Printf("Config map %+v", cfgMap)
}

func (ac ALBController) Reload(data []byte) ([]byte, bool, error) {
	out, err := exec.Command("echo", string(data)).CombinedOutput()
	if err != nil {
		log.Printf("Reloaded new config %s", out)
	} else {
		return out, false, err
	}
	return out, true, err
}

func (ac ALBController) Test(file string) *exec.Cmd {
	return exec.Command("echo", file)
}

func (ac ALBController) OnUpdate(updatePayload ingress.Configuration) ([]byte, error) {
	log.Printf("Received OnUpdate notification")
	for _, b := range updatePayload.Backends {
		eps := []string{}
		for _, e := range b.Endpoints {
			eps = append(eps, e.Address)
		}
		log.Printf("%v: %v", b.Name, strings.Join(eps, ", "))
	}
	for _, b := range updatePayload.Servers {
		log.Printf("%v", b.Hostname)
	}
	return []byte(`<string containing a configuration file>`), nil
}

func (ac ALBController) BackendDefaults() defaults.Backend {
	// Just adopt nginx's default backend config
	return config.NewDefault().Backend
}

func (ac ALBController) Name() string {
	return "dummy Controller"
}

func (ac ALBController) Check(_ *http.Request) error {
	return nil
}

func (ac ALBController) Info() *ingress.BackendInfo {
	return &ingress.BackendInfo{
		Name:       "ALB Controller",
		Release:    "0.0.1",
		Build:      "git-00000000",
		Repository: "git://git.tm.tmcs/kubernetes/alb-ingress-controller",
	}
}