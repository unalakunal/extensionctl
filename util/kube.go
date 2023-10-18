package util

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"

	"github.com/fatih/color"

	appv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func KubeGetDeployment(deploymentName string, namespace string) (*appv1.Deployment, error) {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	var deployment *appv1.Deployment

	deployment, err = clientset.AppsV1().Deployments(namespace).Get(context.TODO(), deploymentName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		color.Red("Deployment %s in namespace %s not found", deploymentName, namespace)
		return nil, err
	} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
		color.Red("Status error while getting deployment %s in namespace %s: %v", deploymentName, namespace, statusError.ErrStatus.Message)
		return nil, err
	} else if err != nil {
		color.Red("Can not get deployment %s in namespace %s", deploymentName, namespace)
		return nil, err
	}
	return deployment, nil
}

func GetEnvVarFromDeployment(deployment *appv1.Deployment, envVarName string) (string, error) {
	val := ""
	valFound := false
	containers := deployment.Spec.Template.Spec.Containers
	if len(containers) > 1 {
		color.Yellow(
			"More than one (%d) containers found in %s, using the first one with image %s", len(containers), deployment.Name, containers[0].Image)
	}
	for _, envVar := range containers[0].Env {
		if envVar.Name == envVarName {
			valFound = true
			val = envVar.Value
			break
		}
	}
	if !valFound {
		return "", fmt.Errorf("KAAPANA_BUILD_VERSION does not exist in env variables of deployment %s", deployment.Name)
	}
	color.Blue("Variable %s has value %s", envVarName, val)
	return val, nil
}
