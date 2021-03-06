package app

import (
	"io/ioutil"
	"path"

	"k8s.io/apimachinery/pkg/util/json"
	kyaml "k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/informers"
	"k8s.io/kubernetes/cmd/kube-controller-manager/app/config"
	"k8s.io/kubernetes/cmd/kube-controller-manager/app/options"
)

var InformerFactoryOverride informers.SharedInformerFactory

func ShimForOpenShift(controllerManagerOptions *options.KubeControllerManagerOptions, controllerManager *config.Config) error {
	if len(controllerManager.OpenShiftContext.OpenShiftConfig) == 0 {
		return nil
	}

	// TODO this gets removed when no longer take flags and no longer build a recycler template
	openshiftConfig, err := getOpenShiftConfig(controllerManager.OpenShiftContext.OpenShiftConfig)
	if err != nil {
		return err
	}

	// TODO this should be replaced by using a flex volume to inject service serving cert CAs into pods instead of adding it to the sa token
	if err := applyOpenShiftServiceServingCertCAFunc(path.Dir(controllerManager.OpenShiftContext.OpenShiftConfig), openshiftConfig); err != nil {
		return err
	}

	// skip GC on some openshift resources
	// TODO this should be replaced by discovery information in some way
	if err := applyOpenShiftGCConfig(controllerManager); err != nil {
		return err
	}

	if err := applyOpenShiftConfigDefaultProjectSelector(controllerManagerOptions, openshiftConfig); err != nil {
		return err
	}

	// Overwrite the informers, because we have our custom generic informers for quota.
	// TODO update quota to create its own informer like garbage collection
	if informers, err := newInformerFactory(controllerManager.Kubeconfig); err != nil {
		return err
	} else {
		InformerFactoryOverride = informers
	}

	return nil
}

func getOpenShiftConfig(configFile string) (map[string]interface{}, error) {
	configBytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	jsonBytes, err := kyaml.ToJSON(configBytes)
	if err != nil {
		return nil, err
	}
	config := map[string]interface{}{}
	if err := json.Unmarshal(jsonBytes, &config); err != nil {
		return nil, err
	}

	return config, nil
}

func applyOpenShiftConfigDefaultProjectSelector(controllerManagerOptions *options.KubeControllerManagerOptions, openshiftConfig map[string]interface{}) error {
	projectConfig, ok := openshiftConfig["projectConfig"]
	if !ok {
		return nil
	}

	castProjectConfig := projectConfig.(map[string]interface{})
	defaultNodeSelector, ok := castProjectConfig["defaultNodeSelector"]
	if !ok {
		return nil
	}
	controllerManagerOptions.OpenShiftContext.OpenShiftDefaultProjectNodeSelector = defaultNodeSelector.(string)

	return nil
}
