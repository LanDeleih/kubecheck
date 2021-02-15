package command

import (
	"errors"
	"fmt"
	"github.com/lanDeleih/kubecheck/app/cmd"
	"github.com/urfave/cli"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"log"
	"os"
	"runtime"
)

var winError = errors.New("does not support windows at this moment")

func NewKubeCheckCommand(VERSION string, logger *zap.SugaredLogger) cli.App {

	clientSet, err := getKubernetesClientSet()
	if err != nil {
		logger.Fatalf("failed to create kubernetes clientSet with kubeConfig: %s", err)
	}

	c := cmd.CheckOpts{
		Client:           clientSet,
		Logger:           logger,
		ContextNamespace: getNamespaceFromConfigContext(),
	}

	return cli.App{
		Name:        "kubecheck",
		Description: "check your application for readiness to production",
		Version:     VERSION,
		Commands: []cli.Command{
			cmd.NewScanCommand(c),
		},
		Flags: CheckFlags(),
	}

}

func CheckFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:  "namespace,n",
			Usage: "Specify namespace name. Example [kubecheck -n default]",
		},
	}
}

func getKubeConfig() (*rest.Config, error) {

	kubeConfigPath, err := getKubeConfigPath()
	if errors.Is(err, winError) {
		return nil, err
	}
	return clientcmd.BuildConfigFromFlags("", kubeConfigPath)
}

func getKubernetesClientSet() (*kubernetes.Clientset, error) {
	kubeConfig, err := getKubeConfig()
	if err != nil {
		log.Fatalf("failed to get kubernetes kubeConfig: %s", err)
	}
	return kubernetes.NewForConfig(kubeConfig)
}

func getKubeConfigPath() (string, error) {
	var kubeConfigPath, home string

	switch operationSystem := runtime.GOOS; operationSystem {
	case "darwin":
		home = fmt.Sprintf("%s/.kube/config", homedir.HomeDir())
	case "linux":
		home = fmt.Sprintf("%s/.kube/config", homedir.HomeDir())
	case "windows":
		return "", winError
	}

	if os.Getenv("KUBECONFIG_PATH") != "" {
		kubeConfigPath = os.Getenv("KUBECONFIG_PATH")
	} else {
		return home, nil
	}
	return kubeConfigPath, nil
}

func getNamespaceFromConfigContext() string {
	cc, _ := clientcmd.NewDefaultClientConfigLoadingRules().Load()
	for k, v := range cc.Contexts {
		if cc.CurrentContext == k {
			return v.Namespace
		}
	}
	return ""
}
